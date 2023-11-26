package validation

import (
	"fmt"
	"net/mail"

	"github.com/nyaruka/phonenumbers"
	"github.com/slimnate/laser-beam/data/user"
)

const (
	FirstNameMinLength = 3
	LastNameMinLength  = 3
	PasswordMinLength  = 4
	PasswordMaxLength  = 64
)

// Validates the following properties on the user object:
// - FirstName
// - LastName
// - Email
// - Phone
// Also automatically formats the phone number in the phonenumbers.NATIONAL format
func ValidateUserUpdate(u *user.User) (valid bool, errors map[string]string) {
	valid = true
	errors = make(map[string]string)
	// Validate first name
	if len(u.FirstName) < FirstNameMinLength {
		errors["FirstName"] = fmt.Sprintf("First name must have at least %d characters", FirstNameMinLength)
		valid = false
	}

	// Validate last name
	if len(u.LastName) < LastNameMinLength {
		errors["LastName"] = fmt.Sprintf("Last name must have at least %d characters", LastNameMinLength)
		valid = false
	}

	// Validate email address
	_, errMail := mail.ParseAddress(u.Email)
	if errMail != nil {
		errors["Email"] = "Invalid email address format"
		valid = false
	}

	//Validate phone number
	p, errPhone := phonenumbers.Parse(u.Phone, "US")
	if errPhone != nil {
		errors["Phone"] = "Invalid phone number"
		valid = false
	}
	u.Phone = phonenumbers.Format(p, phonenumbers.NATIONAL)

	return
}

func ValidatePasswordUpdate(password string, confirmPassword string) (valid bool, errors map[string]string) {
	valid = true
	errors = make(map[string]string)
	if password != confirmPassword {
		errors["Password"] = "Both passwords must match"
		valid = false
		return
	}

	if len(password) < PasswordMinLength {
		errors["Password"] = fmt.Sprintf("Password must be at least %d characters long", PasswordMinLength)
		valid = false
		return
	}

	if len(password) > PasswordMaxLength {
		errors["Password"] = fmt.Sprintf("Password cannot be longer than %d characters", PasswordMaxLength)
		valid = false
		return
	}

	//TODO: password requirements (eg. digits, special chars, blacklisted words, etc.)

	return
}
