package domain

type UserAuth struct {
	User
	ProxyLink string
}

type User struct {
	Username   string
	TelegramId int64
}

type ListUsersReq struct {
}
