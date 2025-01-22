package auth

import (
	"github.com/tavo-wasd-gh/gosmtp"
)

func IsStudent(Production bool, email, passwd string) bool {
	if Production {
		s := smtp.Client("smtp.ucr.ac.cr", "587", passwd)
		if err := s.Validate(email); err != nil {
			return false
		}
	}

	return true
}
