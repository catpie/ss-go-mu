package main

var storage Storage

func SetStorage(s Storage) {
	storage = s
}

type Storage interface {
	// GetUserInfo(UserInterface ) (UserInterface Info, error)
	// StoreUser(u UserInterface) error
	// Exists(u UserInterface) (bool, error)
	Del(u UserInterface) error
	ClearAll() error
	IncrSize(u UserInterface, size int) error
	GetSize(u UserInterface) (int64, error)
	SetSize(u UserInterface, size int) error
	MarkUserOnline(u UserInterface) error
	// IsUserOnline(u UserInterface) bool
	// GetOnlineUsersCount(u []UserInterface) int
}
