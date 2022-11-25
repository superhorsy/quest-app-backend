package helpers

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/superhorsy/quest-app-backend/internal/core/errors"
	"golang.org/x/crypto/bcrypt"
	"math/rand"
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

type TokenData struct {
	sub string
	exp int
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func generateTokenSecret() string {
	rand.Seed(time.Now().UnixNano())

	return randSeq(10)
}

var TokenSecret = generateTokenSecret()

func CreateJwtToken(id string, err error) string {
	tokenContent := jwt.MapClaims{
		"sub": id,
		//"exp": time.Now().Add(time.Minute * 60).Unix(),
	}
	jwtToken := jwt.NewWithClaims(jwt.GetSigningMethod("HS256"), tokenContent)
	token, err := jwtToken.SignedString([]byte(TokenSecret))
	return token
}

func ParseToken(jwtToken string) (jwt.MapClaims, error) {
	cleanJWT := strings.Replace(jwtToken, "Bearer ", "", -1)
	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(cleanJWT, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(TokenSecret), nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("Token is invalid")
	}
	return claims, nil
}
