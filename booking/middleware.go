package booking

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
)

func generateJWT(email string, role string) (tokenString string, tokenExpirationTime time.Time, err error) {
	tokenExpirationTime = time.Now().Add(5 * time.Minute)
	claims := &Claims{
		Email: email,
		Role:  role,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: tokenExpirationTime.Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err = token.SignedString(secretKey)
	if err != nil {
		err = fmt.Errorf("error generating token, err: %v", err)
		return
	}
	return
}

func ValidateJWT(tokenString string) (claims *Claims, err error) {
	claims = &Claims{}

	_, err = jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})
	if err != nil {
		err = fmt.Errorf("unauthorized, err: %v", err)
		return
	}
	return
}

func CheckPasswordHash(authPass string, dbPass string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(dbPass), []byte(authPass))
	return err == nil
}
