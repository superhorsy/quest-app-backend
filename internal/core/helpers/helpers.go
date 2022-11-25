package helpers

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/mailjet/mailjet-apiv3-go"
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

var htmlBody = `
<html>
<head>
   <meta http-equiv="Content-Type" content="text/html; charset=utf-8" />
   <title>Hello, World</title>
</head>
<body>
   <p>This is an email using Go</p>
</body>
`

func SendEmail(email string, name string) error {
	mailjetClient := mailjet.NewMailjetClient("514b16f4b57472627979bd8b47abeaa4", "946dba2f181cf2a125c601d76dd0352b")
	messagesInfo := []mailjet.InfoMessagesV31{
		{
			From: &mailjet.RecipientV31{
				Email: "adamuk.kirill@gmail.com",
				Name:  "Questly",
			},
			To: &mailjet.RecipientsV31{
				mailjet.RecipientV31{
					Email: email,
					Name:  name,
				},
			},
			Subject:  "Greetings from Questly.",
			TextPart: "My first Questly email",
			HTMLPart: "<h3>Dear passenger 1, welcome to <a href='https://www.mailjet.com/'>Questly</a>!</h3><br />May the delivery force be with you!",
			CustomID: "AppGettingStartedTest",
		},
	}
	messages := mailjet.MessagesV31{Info: messagesInfo}
	_, err := mailjetClient.SendMailV31(&messages)
	if err != nil {
		return err
	}
	return nil
}
