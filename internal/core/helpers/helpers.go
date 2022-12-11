package helpers

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/superhorsy/quest-app-backend/internal/core/errors"
	"golang.org/x/crypto/bcrypt"
	"os"
	"strings"
	"time"
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

func CreateJwtToken(id string) (*string, error) {
	tokenContent := jwt.MapClaims{
		"sub": id,
		//"exp": time.Now().Add(time.Minute * 60).Unix(),
	}
	jwtToken := jwt.NewWithClaims(jwt.GetSigningMethod("HS256"), tokenContent)
	token, err := jwtToken.SignedString([]byte(os.Getenv("JWT_PRIVATE_KEY")))
	if err != nil {
		return nil, err
	}
	return &token, nil
}

func ParseToken(authHeader string) (jwt.MapClaims, error) {
	cleanJWT := strings.Replace(authHeader, "Bearer ", "", -1)
	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(cleanJWT, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_PRIVATE_KEY")), nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("Token is invalid")
	}
	return claims, nil
}

func TimeNow() *time.Time {
	now := time.Now().UTC()
	return &now
}

func SliceContains[K comparable](s []K, e K) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
