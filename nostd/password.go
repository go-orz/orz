package nostd

import (
	"golang.org/x/crypto/bcrypt"
)

func BcryptEncode(password []byte) ([]byte, error) {
	return bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
}

func BcryptMatch(hashedPassword, password []byte) error {
	return bcrypt.CompareHashAndPassword(hashedPassword, password)
}
