package domain

type UserAuth struct {
	User
	ProxyLink string
}

type User struct {
	Username string
}

type ListUsersReq struct {
}
