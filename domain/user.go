package domain

type User struct {
	ID       string
	Username string
	Password string
}

type UserRepository interface {
	Save(user *User) (*User, error)
	FindByID(id string) (*User, error)
	FindByUsername(username string) (*User, error)
}
