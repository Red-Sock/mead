package server

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog/log"

	"go.redsock.ru/mead/internal/auth"
	"go.redsock.ru/mead/internal/config"
	"go.redsock.ru/mead/internal/log_key"
)

// SOCKS5 protocol constants (RFC 1928).
const (
	socks5Version = uint8(5)

	// Auth methods
	authNone     = uint8(0x00)
	authPassword = uint8(0x02)
	authNoAccept = uint8(0xFF)

	// Commands
	cmdConnect = uint8(0x01)
	// cmdBind      = uint8(0x02) // not implemented
	// cmdUDPAssoc  = uint8(0x03) // not implemented

	// Address types
	atypIPv4   = uint8(0x01)
	atypDomain = uint8(0x03)
	atypIPv6   = uint8(0x04)

	// Reply codes
	replySuccess          = uint8(0x00)
	replyGeneralFailure   = uint8(0x01)
	replyNotAllowed       = uint8(0x02)
	replyNetUnreachable   = uint8(0x03)
	replyHostUnreachable  = uint8(0x04)
	replyConnRefused      = uint8(0x05)
	replyCmdNotSupported  = uint8(0x07)
	replyAddrNotSupported = uint8(0x08)
)

// Server is a production-ready SOCKS5 proxy server.
type Server struct {
	cfg          *config.Config
	auth         auth.Authenticator
	listener     net.Listener
	dialer       *net.Dialer
	blockedHosts map[string]struct{}
	allowedNets  []*net.IPNet

	activeConns atomic.Int64
	wg          sync.WaitGroup
	closeOnce   sync.Once
	closed      chan struct{}
}

// New creates a Server from cfg. The auth parameter may be nil only when
// cfg.AllowNoAuth is true.
func New(cfg *config.Config, authenticator auth.Authenticator) (*Server, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config must not be nil")
	}
	if !cfg.AllowNoAuth && authenticator == nil {
		return nil, fmt.Errorf("authenticator required when allow_no_auth is false")
	}

	s := &Server{
		cfg:          cfg,
		auth:         authenticator,
		blockedHosts: make(map[string]struct{}),
		closed:       make(chan struct{}),
		dialer: &net.Dialer{
			Timeout:   cfg.DialTimeout,
			KeepAlive: 30 * time.Second,
		},
	}

	for _, h := range cfg.BlockedHosts {
		s.blockedHosts[h] = struct{}{}
	}

	for _, cidr := range cfg.AllowedCIDRs {
		_, network, err := net.ParseCIDR(cidr)
		if err != nil {
			return nil, fmt.Errorf("invalid CIDR %q: %w", cidr, err)
		}
		s.allowedNets = append(s.allowedNets, network)
	}

	return s, nil
}

// ListenAndServe starts the SOCKS5 listener and blocks until the server is
// closed or an unrecoverable error occurs.
func (s *Server) ListenAndServe(ctx context.Context) error {
	ln, err := net.Listen("tcp", s.cfg.Address)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", s.cfg.Address, err)
	}
	return s.Serve(ctx, ln)
}

// Serve accepts connections from ln. It takes ownership of the listener.
func (s *Server) Serve(ctx context.Context, ln net.Listener) error {
	s.listener = ln
	log.Info().
		Str("address", ln.Addr().String()).
		Msg("SOCKS5 server listening")

	// Close listener when context is cancelled.
	go func() {
		select {
		case <-ctx.Done():
			s.Close()
		case <-s.closed:
		}
	}()

	for {
		conn, err := ln.Accept()
		if err != nil {
			select {
			case <-s.closed:
				log.Info().Msg("server shutting down, waiting for active connections")
				s.wg.Wait()
				return nil
			default:
			}
			// Transient errors — back off briefly.
			var ne net.Error
			if errors.As(err, &ne) && ne.Timeout() {
				time.Sleep(5 * time.Millisecond)
				continue
			}
			return fmt.Errorf("accept: %w", err)
		}

		if s.cfg.MaxConnections > 0 &&
			s.activeConns.Load() >= int64(s.cfg.MaxConnections) {
			log.Warn().
				Str(log_key.RemoteAddr, conn.RemoteAddr().String()).
				Msg("max connections reached, rejecting")
			conn.Close()
			continue
		}

		if !s.isAllowedSource(conn.RemoteAddr()) {
			log.Warn().
				Str(log_key.RemoteAddr, conn.RemoteAddr().String()).
				Msg("source IP not in allowlist, rejecting")
			conn.Close()
			continue
		}

		s.wg.Add(1)
		s.activeConns.Add(1)
		go func() {
			defer s.wg.Done()
			defer s.activeConns.Add(-1)
			s.handleConn(conn)
		}()
	}
}

// Close shuts down the server gracefully.
func (s *Server) Close() {
	s.closeOnce.Do(func() {
		close(s.closed)
		if s.listener != nil {
			s.listener.Close()
		}
	})
}

// ActiveConnections returns the current number of active client connections.
func (s *Server) ActiveConnections() int64 {
	return s.activeConns.Load()
}

// isAllowedSource returns true when the source address is permitted.
func (s *Server) isAllowedSource(addr net.Addr) bool {
	if len(s.allowedNets) == 0 {
		return true
	}
	host, _, err := net.SplitHostPort(addr.String())
	if err != nil {
		return false
	}
	ip := net.ParseIP(host)
	if ip == nil {
		return false
	}
	for _, network := range s.allowedNets {
		if network.Contains(ip) {
			return true
		}
	}
	return false
}

// handleConn drives the SOCKS5 handshake and data relay for one client.
func (s *Server) handleConn(conn net.Conn) {
	defer conn.Close()
	remote := conn.RemoteAddr().String()
	log.Debug().
		Str(log_key.RemoteAddr, remote).
		Msg("new connection")

	if s.cfg.ReadTimeout > 0 {
		conn.SetDeadline(time.Now().Add(s.cfg.ReadTimeout))
	}

	// --- Negotiate authentication method ---
	if err := s.negotiateAuth(conn); err != nil {
		log.Warn().
			Err(err).
			Str(log_key.RemoteAddr, remote).
			Msg("auth negotiation failed")
		return
	}

	// --- Authenticate (RFC 1929) ---
	if !s.cfg.AllowNoAuth {
		if err := s.authenticate(conn); err != nil {
			log.Warn().
				Err(err).
				Str(log_key.RemoteAddr, remote).
				Msg("authentication failed")
			return
		}
	}

	// --- Handle CONNECT request ---
	target, err := s.readRequest(conn)
	if err != nil {
		log.Warn().
			Err(err).
			Str(log_key.RemoteAddr, remote).
			Msg("bad request")
		return
	}

	// Reset deadline for the long-lived relay phase.
	conn.SetDeadline(time.Time{})

	log.Info().
		Str(log_key.RemoteAddr, remote).
		Str(log_key.TargetAddr, target).
		Msg("Connection established")

	if s.isBlocked(target) {
		log.Warn().
			Str(log_key.RemoteAddr, remote).
			Str(log_key.TargetAddr, target).
			Msg("blocked destination ")
		writeReply(conn, replyNotAllowed)
		return
	}

	// --- Dial target ---
	dst, err := s.dialer.Dial("tcp", target)
	if err != nil {
		log.Warn().
			Err(err).
			Str(log_key.TargetAddr, target).
			Msg("dial failed")
		writeReply(conn, dialErrToReply(err))
		return
	}
	defer dst.Close()

	// --- Send success reply ---
	err = writeSuccessReply(conn, dst.LocalAddr())
	if err != nil {
		log.Warn().
			Err(err).
			Str(log_key.TargetAddr, target).
			Msg("Write reply ")
		return
	}

	// --- Relay data ---
	relay(conn, dst)
	log.Debug().
		Str(log_key.RemoteAddr, remote).
		Str(log_key.TargetAddr, target).
		Msg("connection closed")
}

// negotiateAuth sends/receives the method selection handshake.
func (s *Server) negotiateAuth(conn net.Conn) error {
	// Read VER + NMETHODS
	header := make([]byte, 2)
	if _, err := io.ReadFull(conn, header); err != nil {
		return fmt.Errorf("read header: %w", err)
	}
	if header[0] != socks5Version {
		return fmt.Errorf("unsupported SOCKS version: %d", header[0])
	}
	nMethods := int(header[1])
	if nMethods == 0 {
		return fmt.Errorf("no auth methods offered")
	}

	methods := make([]byte, nMethods)
	if _, err := io.ReadFull(conn, methods); err != nil {
		return fmt.Errorf("read methods: %w", err)
	}

	selected := authNoAccept
	for _, m := range methods {
		if s.cfg.AllowNoAuth && m == authNone {
			selected = authNone
			break
		}
		if !s.cfg.AllowNoAuth && m == authPassword {
			selected = authPassword
			break
		}
	}

	_, err := conn.Write([]byte{socks5Version, selected})
	if err != nil {
		return fmt.Errorf("write method selection: %w", err)
	}
	if selected == authNoAccept {
		return fmt.Errorf("no acceptable auth method")
	}
	return nil
}

// authenticate performs RFC 1929 sub-negotiation.
func (s *Server) authenticate(conn net.Conn) error {
	// VER (0x01 for subnegotiation)
	ver := make([]byte, 1)
	if _, err := io.ReadFull(conn, ver); err != nil {
		return fmt.Errorf("read auth version: %w", err)
	}
	if ver[0] != 0x01 {
		return fmt.Errorf("unsupported auth version: %d", ver[0])
	}

	ulen := make([]byte, 1)
	if _, err := io.ReadFull(conn, ulen); err != nil {
		return fmt.Errorf("read ulen: %w", err)
	}
	username := make([]byte, ulen[0])
	if _, err := io.ReadFull(conn, username); err != nil {
		return fmt.Errorf("read username: %w", err)
	}

	plen := make([]byte, 1)
	if _, err := io.ReadFull(conn, plen); err != nil {
		return fmt.Errorf("read plen: %w", err)
	}
	password := make([]byte, plen[0])
	if _, err := io.ReadFull(conn, password); err != nil {
		return fmt.Errorf("read password: %w", err)
	}

	if err := s.auth.Authenticate(string(username), string(password)); err != nil {
		conn.Write([]byte{0x01, 0x01}) //nolint:errcheck — best-effort
		return fmt.Errorf("authenticate %q: %w", username, err)
	}

	_, err := conn.Write([]byte{0x01, 0x00}) // success
	return err
}

// readRequest parses a SOCKS5 CONNECT request and returns "host:port".
func (s *Server) readRequest(conn net.Conn) (string, error) {
	// VER CMD RSV ATYP
	hdr := make([]byte, 4)
	if _, err := io.ReadFull(conn, hdr); err != nil {
		return "", fmt.Errorf("read request header: %w", err)
	}
	if hdr[0] != socks5Version {
		return "", fmt.Errorf("bad version in request: %d", hdr[0])
	}
	if hdr[1] != cmdConnect {
		writeReply(conn, replyCmdNotSupported) //nolint:errcheck
		return "", fmt.Errorf("unsupported command: %d", hdr[1])
	}

	var host string
	switch hdr[3] {
	case atypIPv4:
		addr := make([]byte, 4)
		if _, err := io.ReadFull(conn, addr); err != nil {
			return "", fmt.Errorf("read IPv4: %w", err)
		}
		host = net.IP(addr).String()

	case atypIPv6:
		addr := make([]byte, 16)
		if _, err := io.ReadFull(conn, addr); err != nil {
			return "", fmt.Errorf("read IPv6: %w", err)
		}
		host = net.IP(addr).String()

	case atypDomain:
		dlen := make([]byte, 1)
		if _, err := io.ReadFull(conn, dlen); err != nil {
			return "", fmt.Errorf("read domain length: %w", err)
		}
		domain := make([]byte, dlen[0])
		if _, err := io.ReadFull(conn, domain); err != nil {
			return "", fmt.Errorf("read domain: %w", err)
		}
		host = string(domain)

	default:
		writeReply(conn, replyAddrNotSupported) //nolint:errcheck
		return "", fmt.Errorf("unsupported address type: %d", hdr[3])
	}

	portBytes := make([]byte, 2)
	if _, err := io.ReadFull(conn, portBytes); err != nil {
		return "", fmt.Errorf("read port: %w", err)
	}
	port := binary.BigEndian.Uint16(portBytes)

	return net.JoinHostPort(host, strconv.Itoa(int(port))), nil
}

// isBlocked returns true if the target host/IP should be rejected.
func (s *Server) isBlocked(target string) bool {
	if len(s.blockedHosts) == 0 {
		return false
	}
	host, _, err := net.SplitHostPort(target)
	if err != nil {
		return false
	}
	_, blocked := s.blockedHosts[host]
	return blocked
}

// writeReply sends a minimal SOCKS5 reply with the given status code.
func writeReply(conn net.Conn, rep uint8) error {
	// VER REP RSV ATYP BND.ADDR(4) BND.PORT(2)
	reply := []byte{socks5Version, rep, 0x00, atypIPv4, 0, 0, 0, 0, 0, 0}
	_, err := conn.Write(reply)
	return err
}

// writeSuccessReply sends a success reply populated with the bound address.
func writeSuccessReply(conn net.Conn, bound net.Addr) error {
	host, portStr, err := net.SplitHostPort(bound.String())
	if err != nil {
		return writeReply(conn, replySuccess)
	}
	port, _ := strconv.Atoi(portStr)
	ip := net.ParseIP(host)

	var reply []byte
	if ip4 := ip.To4(); ip4 != nil {
		reply = append([]byte{socks5Version, replySuccess, 0x00, atypIPv4}, ip4...)
	} else if ip6 := ip.To16(); ip6 != nil {
		reply = append([]byte{socks5Version, replySuccess, 0x00, atypIPv6}, ip6...)
	} else {
		reply = []byte{socks5Version, replySuccess, 0x00, atypIPv4, 0, 0, 0, 0}
	}
	reply = append(reply, byte(port>>8), byte(port))
	_, err = conn.Write(reply)
	return err
}

// dialErrToReply maps net.Dial errors to SOCKS5 reply codes.
func dialErrToReply(err error) uint8 {
	if err == nil {
		return replySuccess
	}
	var netErr *net.OpError
	if errors.As(err, &netErr) {
		if netErr.Timeout() {
			return replyHostUnreachable
		}
	}
	return replyGeneralFailure
}

// relay copies data bidirectionally between a and b until either side closes.
func relay(a, b net.Conn) {
	var wg sync.WaitGroup
	wg.Add(2)
	copy := func(dst, src net.Conn) {
		defer wg.Done()
		defer dst.(*net.TCPConn).CloseWrite() //nolint:errcheck
		io.Copy(dst, src)                     //nolint:errcheck
	}
	go copy(b, a)
	go copy(a, b)
	wg.Wait()
}
