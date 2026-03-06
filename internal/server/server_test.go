package server

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"testing"
	"time"

	"go.redsock.ru/mead/internal/auth"
	"go.redsock.ru/mead/internal/config"
)

// ---- helpers ----------------------------------------------------------------

func newTestServer(t *testing.T, allowNoAuth bool, creds ...config.Credential) (addr string, cleanup func()) {
	t.Helper()

	cfg := config.Defaults()
	cfg.AllowNoAuth = allowNoAuth
	cfg.Credentials = creds
	cfg.DialTimeout = 5 * time.Second
	cfg.ReadTimeout = 10 * time.Second
	cfg.MaxConnections = 10

	var authenticator auth.Authenticator
	if !allowNoAuth {
		store := auth.NewCredentialStore()
		for _, cr := range creds {
			store.Add(cr.Username, cr.Password) //nolint:errcheck
		}
		authenticator = store
	}

	srv, err := New(cfg, authenticator)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	go srv.Serve(ctx, ln) //nolint:errcheck

	return ln.Addr().String(), func() {
		cancel()
		srv.Close()
	}
}

// echoServer listens on a random local port, echoes back every byte it reads,
// and closes after the first connection.
func echoServer(t *testing.T) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("echo server listen: %v", err)
	}
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		defer ln.Close()
		io.Copy(conn, conn) //nolint:errcheck
	}()
	return ln.Addr().String()
}

// socks5Connect dials the proxy, negotiates auth+CONNECT and returns the conn.
func socks5Connect(t *testing.T, proxyAddr, targetHost string, targetPort uint16, username, password string) net.Conn {
	t.Helper()

	conn, err := net.DialTimeout("tcp", proxyAddr, 5*time.Second)
	if err != nil {
		t.Fatalf("dial proxy: %v", err)
	}

	useAuth := username != ""

	// --- Method selection ---
	if useAuth {
		conn.Write([]byte{0x05, 0x01, 0x02}) //nolint:errcheck // VER NMETHODS [PASSWORD]
	} else {
		conn.Write([]byte{0x05, 0x01, 0x00}) //nolint:errcheck // VER NMETHODS [NONE]
	}

	resp := make([]byte, 2)
	if _, err := io.ReadFull(conn, resp); err != nil {
		t.Fatalf("read method response: %v", err)
	}
	if resp[0] != 0x05 {
		t.Fatalf("bad version in method response: %d", resp[0])
	}
	if useAuth && resp[1] != 0x02 {
		t.Fatalf("expected password auth (0x02), got 0x%02x", resp[1])
	}
	if !useAuth && resp[1] != 0x00 {
		t.Fatalf("expected no auth (0x00), got 0x%02x", resp[1])
	}

	// --- RFC 1929 sub-negotiation ---
	if useAuth {
		ulen := byte(len(username))
		plen := byte(len(password))
		msg := append([]byte{0x01, ulen}, []byte(username)...)
		msg = append(msg, plen)
		msg = append(msg, []byte(password)...)
		conn.Write(msg) //nolint:errcheck

		authResp := make([]byte, 2)
		if _, err := io.ReadFull(conn, authResp); err != nil {
			t.Fatalf("read auth response: %v", err)
		}
		if authResp[1] != 0x00 {
			t.Fatalf("auth failed: status 0x%02x", authResp[1])
		}
	}

	// --- CONNECT request ---
	hostBytes := []byte(targetHost)
	req := []byte{0x05, 0x01, 0x00, 0x03, byte(len(hostBytes))}
	req = append(req, hostBytes...)
	portBuf := make([]byte, 2)
	binary.BigEndian.PutUint16(portBuf, targetPort)
	req = append(req, portBuf...)
	conn.Write(req) //nolint:errcheck

	// --- Read reply ---
	hdr := make([]byte, 4)
	if _, err := io.ReadFull(conn, hdr); err != nil {
		t.Fatalf("read reply header: %v", err)
	}
	if hdr[1] != 0x00 {
		t.Fatalf("CONNECT failed: reply code 0x%02x", hdr[1])
	}

	// Consume bound address.
	switch hdr[3] {
	case 0x01: // IPv4
		io.ReadFull(conn, make([]byte, 4+2)) //nolint:errcheck
	case 0x04: // IPv6
		io.ReadFull(conn, make([]byte, 16+2)) //nolint:errcheck
	case 0x03:
		l := make([]byte, 1)
		io.ReadFull(conn, l)                         //nolint:errcheck
		io.ReadFull(conn, make([]byte, int(l[0])+2)) //nolint:errcheck
	}

	return conn
}

// ---- tests ------------------------------------------------------------------

func TestServer_AuthenticatedConnect(t *testing.T) {
	echoAddr := echoServer(t)
	host, portStr, _ := net.SplitHostPort(echoAddr)
	var port uint16
	fmt.Sscanf(portStr, "%d", &port)

	proxyAddr, cleanup := newTestServer(t, false,
		config.Credential{Username: "alice", Password: "s3cr3t"},
	)
	defer cleanup()

	conn := socks5Connect(t, proxyAddr, host, port, "alice", "s3cr3t")
	defer conn.Close()

	msg := []byte("hello socks5")
	conn.Write(msg) //nolint:errcheck

	buf := make([]byte, len(msg))
	if _, err := io.ReadFull(conn, buf); err != nil {
		t.Fatalf("echo read: %v", err)
	}
	if string(buf) != string(msg) {
		t.Errorf("echo mismatch: got %q, want %q", buf, msg)
	}
}

func TestServer_NoAuthMode(t *testing.T) {
	echoAddr := echoServer(t)
	host, portStr, _ := net.SplitHostPort(echoAddr)
	var port uint16
	fmt.Sscanf(portStr, "%d", &port)

	cfg := config.Defaults()
	cfg.AllowNoAuth = true

	srv, _ := New(cfg, nil)

	ln, _ := net.Listen("tcp", "127.0.0.1:0")
	ctx, cancel := context.WithCancel(context.Background())
	go srv.Serve(ctx, ln) //nolint:errcheck
	defer func() { cancel(); srv.Close() }()

	conn := socks5Connect(t, ln.Addr().String(), host, port, "", "")
	defer conn.Close()

	msg := []byte("no auth test")
	conn.Write(msg) //nolint:errcheck
	buf := make([]byte, len(msg))
	io.ReadFull(conn, buf) //nolint:errcheck
	if string(buf) != string(msg) {
		t.Errorf("echo mismatch: got %q, want %q", buf, msg)
	}
}

func TestServer_WrongPassword(t *testing.T) {
	proxyAddr, cleanup := newTestServer(t, false,
		config.Credential{Username: "alice", Password: "correct"},
	)
	defer cleanup()

	conn, err := net.DialTimeout("tcp", proxyAddr, 3*time.Second)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close()

	// Method selection — request password auth
	conn.Write([]byte{0x05, 0x01, 0x02}) //nolint:errcheck
	resp := make([]byte, 2)
	io.ReadFull(conn, resp) //nolint:errcheck

	// Sub-negotiation with wrong password
	user := "alice"
	pass := "wrong"
	msg := []byte{0x01, byte(len(user))}
	msg = append(msg, []byte(user)...)
	msg = append(msg, byte(len(pass)))
	msg = append(msg, []byte(pass)...)
	conn.Write(msg) //nolint:errcheck

	authResp := make([]byte, 2)
	io.ReadFull(conn, authResp) //nolint:errcheck
	if authResp[1] == 0x00 {
		t.Error("expected auth failure but got success")
	}
}

func TestServer_MaxConnections(t *testing.T) {
	cfg := config.Defaults()
	cfg.AllowNoAuth = true
	cfg.MaxConnections = 1

	srv, _ := New(cfg, nil)

	ln, _ := net.Listen("tcp", "127.0.0.1:0")
	ctx, cancel := context.WithCancel(context.Background())
	go srv.Serve(ctx, ln) //nolint:errcheck
	defer func() { cancel(); srv.Close() }()

	// First connection — accepted, sits in handshake
	c1, err := net.DialTimeout("tcp", ln.Addr().String(), 3*time.Second)
	if err != nil {
		t.Fatalf("first dial: %v", err)
	}
	defer c1.Close()

	// Give the server goroutine time to increment the counter.
	time.Sleep(50 * time.Millisecond)

	// Second connection — should be rejected immediately
	c2, err := net.DialTimeout("tcp", ln.Addr().String(), 3*time.Second)
	if err != nil {
		// OS-level rejection is also acceptable.
		return
	}
	defer c2.Close()

	// If accepted at TCP level, server should close it right away.
	c2.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
	buf := make([]byte, 1)
	_, err = c2.Read(buf)
	if err == nil {
		t.Error("expected second connection to be closed by server")
	}
}

func TestServer_ActiveConnections(t *testing.T) {
	cfg := config.Defaults()
	cfg.AllowNoAuth = true

	srv, _ := New(cfg, nil)

	ln, _ := net.Listen("tcp", "127.0.0.1:0")
	ctx, cancel := context.WithCancel(context.Background())
	go srv.Serve(ctx, ln) //nolint:errcheck
	defer func() { cancel(); srv.Close() }()

	if srv.ActiveConnections() != 0 {
		t.Errorf("expected 0 active connections, got %d", srv.ActiveConnections())
	}

	conn, _ := net.DialTimeout("tcp", ln.Addr().String(), 3*time.Second)
	defer conn.Close()
	time.Sleep(50 * time.Millisecond)

	if srv.ActiveConnections() != 1 {
		t.Errorf("expected 1 active connection, got %d", srv.ActiveConnections())
	}
}

func TestServer_InvalidSocksVersion(t *testing.T) {
	proxyAddr, cleanup := newTestServer(t, true)
	defer cleanup()

	conn, _ := net.DialTimeout("tcp", proxyAddr, 3*time.Second)
	defer conn.Close()

	// Send SOCKS4 header
	conn.Write([]byte{0x04, 0x01, 0x00}) //nolint:errcheck

	conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
	buf := make([]byte, 16)
	_, err := conn.Read(buf)
	// Server should close the connection
	if err == nil {
		t.Error("expected connection to be closed after invalid version")
	}
}
