package booking

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/mail"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Coderx44/MovieTicketingPortal/app"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

func PingHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	t, _ := time.Parse("15:04:05", "18:00:00")
	st := t.Format(time.Kitchen)
	t, _ = time.Parse("15:04:05", st)
	log.Println(id, t)
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
		// log.Println(authUser)
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

		var newM NewMovie
		var newmovie = make(map[string]string)
		json.NewDecoder(r.Body).Decode(&newmovie)
		rDate, err := time.Parse(DateOnly, newmovie["release_date"])
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Err: Invalid release date"))
			return
		}
		newM.Title = newmovie["title"]
		newM.Language = newmovie["language"]
		newM.Genre = newmovie["genre"]
		newM.Release_date = rDate
		newM.Duration, err = strconv.ParseFloat(newmovie["duration"], 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Err: Invalid duration"))
			return
		}

		if newM.Title == "" || newM.Language == "" || newM.Release_date.IsZero() || newM.Genre == "" || newM.Duration == 0.0 {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Provide the required parameters"))
			return
		}

		movie_id, err := s.AddMovie(r.Context(), newM)

		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Err - Internal Server Error - Failed to Add Movie"))
			return
		}

		respBytes, _ := json.Marshal(movie_id)
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(respBytes)

	})
}

func AddScreen(s Service) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		vars := mux.Vars(r)
		id, ok := vars["id"]
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode("Invalid Multiplex id")
			return
		}
		var newSn NewScreen
		json.NewDecoder(r.Body).Decode(&newSn)
		newSn.Multiplex_id, _ = strconv.Atoi(id)

		if newSn.Screen_number == 0 || newSn.Total_seats == 0 || newSn.Sound_system == "" || newSn.Screen_dimension == "" || newSn.Multiplex_id == 0 {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Provide the required parameters"))
			return
		}
		log.Println("newsnn", newSn)
		screen_id, err := s.AddScreen(r.Context(), newSn)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Err - Internal Server Error - Failed to add screen"))
			return
		}

		respBytes, _ := json.Marshal(screen_id)
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(respBytes)

	})
}

func AddMultiplex(s Service) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		var newM NewMultiplex
		json.NewDecoder(r.Body).Decode(&newM)
		if newM.Name == "" || newM.Contact == "" || newM.Total_screens == 0 || newM.Locality == "" || newM.City == "" || newM.State == "" || newM.Pincode == 0 {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Provide the required parameters"))
			return
		}

		multiplex_id, err := s.AddMultiplex(r.Context(), newM)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Err - Internal Server Error - Failed to add multiplex"))
			return
		}

		respBytes, _ := json.Marshal(multiplex_id)
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(respBytes)

	})
}

func AddShow(s Service) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		vars := mux.Vars(r)
		id, ok := vars["id"]

		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode("Invalid Multiplex id")
			return
		}
		var newShow NewShow
		json.NewDecoder(r.Body).Decode(&newShow)
		// log.Println("newsh", newShow)
		var err error
		newShow.Multiplex_id, err = strconv.Atoi(id)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Invalid id"))
			return
		}

		show_id, err := s.AddShow(r.Context(), newShow)

		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Err: Internal Server Error - Failed to add show"))
			return
		}

		respBytes, _ := json.Marshal(show_id)
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(respBytes)
	})

}

func GetAllMultiplexesByCity(s Service) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		vars := mux.Vars(r)
		city, ok := vars["city"]
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode("city required")
			return
		}

		multiplexes, err := s.GetAllMultiplexesByCity(r.Context(), city)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Err: Internal Server Error - Failed to fetch multiplexes"))
			return
		}

		json.NewEncoder(w).Encode(&multiplexes)

	})
}

func GetAllShowsByDateAndMultiplexId(s Service) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		values := r.URL.Query()
		log.Println(values["date"][0])

		multiplex_id, _ := strconv.Atoi(values["multiplex_id"][0])
		resp, err := s.GetAllShowsByDateAndMultiplexId(r.Context(), values["date"][0], multiplex_id)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Err: Internal Server Error - Failed to fetch all shows"))
			return
		}
		json.NewEncoder(w).Encode(resp)
	})
}

func GetAllShowsByMovieAndDate(s Service) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		values := r.URL.Query()
		log.Println(values["date"][0])

		resp, err := s.GetAllShowsByMovieAndDate(r.Context(), values["date"][0], values["title"][0], values["city"][0])
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Err: Internal Server Error - Failed to fetch all shows"))
			return
		}
		json.NewEncoder(w).Encode(resp)
	})
}

func GetAllSeatsByShowID(s Service) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		s_id, ok := vars["show_id"]
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Err: show id required"))
			return
		}
		show_id, _ := strconv.Atoi(s_id)

		resp, err := s.GetAllSeatsByShowID(r.Context(), show_id)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		json.NewEncoder(w).Encode(resp)
	})

}

func BookSeats(s Service) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		var seats = make(map[string][]int)

		claims := r.Context().Value("token")

		claim := claims.(*Claims)

		json.NewDecoder(r.Body).Decode(&seats)
		log.Println(seats)
		invoice, err := s.AddBookingsBySeatId(r.Context(), seats["seats"], claim.Email)
		if err != nil && err.Error() == "Seats not available" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		} else if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		json.NewEncoder(w).Encode(invoice)
	})
}

func GetUpcomingMovies(s Service) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		currentDate := time.Now()
		rdate := currentDate.Format("2006-01-02")
		log.Println(len(rdate), rdate)
		movies, err := s.GetUpcomingMovies(r.Context(), rdate)
		if err != nil {
			// log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		json.NewEncoder(w).Encode(movies)

	})
}

func GetMovieByTitle(s Service) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		title, ok := vars["title"]
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Err: title required"))
			return
		}
		movie, err := s.GetMovieByTitle(r.Context(), title)
		if err != nil && err.Error() == "movie not found" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(err.Error()))
			return
		} else if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		json.NewEncoder(w).Encode(movie)

	})
}
