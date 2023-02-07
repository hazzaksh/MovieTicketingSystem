package booking

import (
	"encoding/json"
	"net/http"
	"net/mail"
	"regexp"
	"strings"

	"github.com/Coderx44/MovieTicketingPortal/app"
)

func PingHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("pong"))
}

func CreateNewUser(s Service) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

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
