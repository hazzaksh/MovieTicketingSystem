package booking

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/mail"
	"regexp"
	"strings"

	"github.com/Coderx44/MovieTicketingPortal/app"
	"golang.org/x/crypto/bcrypt"
)

func PingHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("pong"))
}

func CreateNewUser(s Service) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		role := strings.Split(path, "/")[2]
		var newUser NewUser
		err := json.NewDecoder(r.Body).Decode(&newUser)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Internal Server error"))
			return
		}

		if newUser.Email == "" || newUser.Password == "" || newUser.Name == "" || newUser.Phone_number == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Provide the required parameters"))
			return
		}

		if _, err := mail.ParseAddress(newUser.Email); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Err: Invalid email address"))
			return
		}
		newUser.Email = strings.Trim(newUser.Email, " ")
		re := regexp.MustCompile(`^\d{10}$`)
		if !re.MatchString(newUser.Phone_number) {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Err: Phone must contain 10 digits"))
			return
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newUser.Password), bcrypt.DefaultCost)
		if err != nil {
			panic(err)
		}

		newUser.Password = string(hashedPassword)
		newUser.Role = role
		newResp, err := s.CreateNewUser(r.Context(), newUser)

		if err != nil {
			if err.Error() == "account exists for the given email" {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("Err: User already exits for given email"))
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Err - Internal Server Error - Failure creating user account"))
			return
		}
		respBytes, err := json.Marshal(newResp)
		status := http.StatusOK
		if err != nil {
			app.GetLogger().Error(err)
			status = http.StatusInternalServerError
		}
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(status)
		w.Write(respBytes)

	})

}

func Login(s Service) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var authUser Authentication
		err := json.NewDecoder(r.Body).Decode(&authUser)
		log.Println(authUser)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(err)
			return
		}

		if authUser.Email == "" || authUser.Password == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Err: Email address and password must be provided"))
			return
		}
		if _, err := mail.ParseAddress(authUser.Email); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Err - Invalid email address"))
			return
		}
		authUser.Email = strings.Trim(authUser.Email, " ")
		tokenString, _, err := s.Login(r.Context(), authUser)

		if err != nil {
			if err == errors.New("unauthorized") {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte("Unauthorized"))
				return
			}
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Err: Internal Server Error"))
			return
		}
		var resp = LoginResp{
			Token: tokenString,
			Mssg:  "Successfully logged in",
		}
		json.NewEncoder(w).Encode(resp)

	})
}

func AddMovie(s Service) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header["Token"]

		claims, err := ValidateJWT(tokenString[0])
		if err != nil || claims.Role != "admin" {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode("Unauthorized")
			return
		}

		var newM NewMovie
		json.NewDecoder(r.Body).Decode(&newM)

		if newM.Title == "" || newM.Language == "" || newM.Release_date == "" || newM.Genre == "" || newM.Duration == 0.0 {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Provide the required parameters"))
			return
		}

		movie_id, err := s.AddMovie(r.Context(), newM)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Err - Internal Server Error - Failure creating user account"))
			return
		}

		respBytes, _ := json.Marshal(movie_id)
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(respBytes)

	})
}
