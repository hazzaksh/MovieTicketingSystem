package booking

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Coderx44/MovieTicketingPortal/db"
	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
)

func generateJWT(email string, role string) (tokenString string, tokenExpirationTime time.Time, err error) {
	tokenExpirationTime = time.Now().Add(30 * time.Minute)
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
	log.Println(tokenString)
	_, err = jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})
	log.Println(err)
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

func getLocationID(ctx context.Context, b *bookingService, city string, state string, pincode int) (location_id uint, err error) {

	location_id, err = b.store.GetLocationIdByCity(ctx, city)

	var lerr = "location doesn't exist"
	if err.Error() == lerr {

		newL := db.Location{
			City:    city,
			State:   state,
			Pincode: pincode,
		}
		location_id, err = b.store.AddLocation(ctx, newL)

		if err != nil {
			if err != nil {
				b.logger.Errorf("Err: Adding Location: %v", err.Error())
				return
			}
		}
		return
	}

	return

}

func MultiplexIdExists(b *bookingService, ctx context.Context, multiplex_id int) bool {

	_, err := b.store.GetMultiplexeByID(ctx, multiplex_id)
	log.Println(multiplex_id, err)
	return err == nil

}

func ScreenExists(b *bookingService, ctx context.Context, screen int, multpx_id int) (db.Screen, bool) {

	s, err := b.store.GetScreenByNumberAndMultiplexID(ctx, screen, multpx_id)
	return s, err == nil
}

func MovieExists(b *bookingService, ctx context.Context, title string) (int, bool) {
	movie_id, err := b.store.GetMovieByTitle(ctx, title)
	return int(movie_id), err == nil
}
