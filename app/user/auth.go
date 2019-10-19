package user

import (
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password []byte) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
}

func VerifyPassword(hash []byte, inputPassword []byte) error {
	return bcrypt.CompareHashAndPassword(hash, inputPassword)
}