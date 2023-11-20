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

func (u *User) AdminStatusString() string {
	if u.AdminStatus == 0 {
		return "User"
	}
	if u.AdminStatus == 1 {
		return "Admin"
	}
	if u.AdminStatus == 2 {
		return "Global Admin"
	}
	return "N/A"
}
