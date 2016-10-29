package main

var storage Storage

func SetStorage(s Storage) {
	storage = s
}

type Storage interface {
	// GetUserInfo(UserInterface ) (UserInterface Info, error)
	// StoreUser(UserInterface Info) error
	Exists(UserInterface) (bool, error)
	Del(UserInterface) error
	ClearAll() error
	IncrSize(u UserInterface, size int) error
	GetSize(u UserInterface) (int64, error)
	SetSize(u UserInterface, size int) error
	MarkUserOnline(u UserInterface) error
	IsUserOnline(u UserInterface) bool
	GetOnlineUsersCount(u []UserInterface) int
}
