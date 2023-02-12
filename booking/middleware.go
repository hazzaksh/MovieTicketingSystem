package booking

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Coderx44/MovieTicketingPortal/db"
	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
)

func generateJWT(email string, role string) (tokenString string, err error) {
	var tokenExpirationTime = time.Now().Add(30 * time.Minute)
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

func ValidateJWT(next http.HandlerFunc, role string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}
		claims := &Claims{}
		_, err := jwt.ParseWithClaims(authHeader, claims, func(token *jwt.Token) (interface{}, error) {
			return secretKey, nil
		})
		if err != nil {
			http.Error(w, "Token is invalid", http.StatusUnauthorized)
			return
		}

		if claims.Role != role {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode("Unauthorized")
			return
		}

		ctx := context.WithValue(r.Context(), "token", claims)
		next.ServeHTTP(w, r.WithContext(ctx))

	})
}

// func ValidateJWT(tokenString string) (claims *Claims, err error) {
// 	claims = &Claims{}
// 	log.Println(tokenString)
// 	_, err = jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
// 		return secretKey, nil
// 	})
// 	log.Println(err)
// 	if err != nil {
// 		err = fmt.Errorf("unauthorized, err: %v", err)
// 		return
// 	}
// 	return
// }

func CheckPasswordHash(authPass string, dbPass string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(dbPass), []byte(authPass))
	return err == nil
}

func getLocationID(ctx context.Context, b *bookingService, city string, state string, pincode int) (location_id int, err error) {

	location, err := b.store.GetLocationIdByCity(ctx, city)
	location_id = int(location.Location_id)
	var lerr = "location doesn't exist"

	if err != nil && err.Error() == lerr {

		newL := db.Location{
			City:    city,
			State:   state,
			Pincode: pincode,
		}
		location_id, err = b.store.AddLocation(ctx, newL)

		// if err != nil {
		// 	b.logger.Errorf("Err: Adding Location: %v", err.Error())
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

func MovieExists(b *bookingService, ctx context.Context, title string) (movie NewMovie, err error) {
	m, err := b.store.GetMovieByTitle(ctx, title)
	return NewMovie(m), err
}
