package helpers

import (
	"golang.org/x/crypto/bcrypt"
)

func HandleError(err error) {
	if err != nil {
		panic(err.Error())
	}
}

func HashAndSalt(pass []byte) string {
	hashed, err := bcrypt.GenerateFromPassword(pass, bcrypt.MinCost)
	HandleError(err)
	return string(hashed)
}
