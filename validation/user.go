package validation

import "fmt"

const (
	FirstNameMinLength = 3
	LastNameMinLength  = 3
	PasswordMinLength  = 4
	PasswordMaxLength  = 64
)

type UserValidationErrors struct {
	FirstName string
	LastName  string
}

type PasswordValidationErrors struct {
	Password string
}

func ValidateUserUpdate(first string, last string) (valid bool, errors *UserValidationErrors) {
	errors = &UserValidationErrors{}
	valid = true
	if len(first) <= FirstNameMinLength {
		errors.FirstName = fmt.Sprintf("First name must have at least %d characters", FirstNameMinLength)
		valid = false
	}
	if len(last) <= LastNameMinLength {
		errors.LastName = fmt.Sprintf("Last name must have at least %d characters", LastNameMinLength)
		valid = false
	}
	return
}

func ValidatePasswordUpdate(password string, confirmPassword string) (valid bool, errors *PasswordValidationErrors) {
	valid = true
	errors = &PasswordValidationErrors{}
	if password != confirmPassword {
		errors.Password = "Both passwords must match"
		valid = false
		return
	}

	if len(password) <= PasswordMinLength {
		errors.Password = fmt.Sprintf("Password must be at least %d characters long", PasswordMinLength)
		valid = false
		return
	}

	if len(password) > PasswordMaxLength {
		errors.Password = fmt.Sprintf("Password cannot be longer than %d characters", PasswordMaxLength)
		valid = false
		return
	}

	//TODO: password requirements (eg. digits, special chars, blacklisted words, etc.)

	return
}
