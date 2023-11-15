package user

import "fmt"

type User struct {
	ID             int64
	Username       string
	FirstName      string
	LastName       string
	AdminStatus    int64 // 0 - normal user, 1 - org admin, 2 - global admin
	OrganizationID int64
}

type UserSecret struct {
	User
	Password string
}

func (u *User) FullName() string {
	return fmt.Sprintf("%s %s", u.FirstName, u.LastName)
}
