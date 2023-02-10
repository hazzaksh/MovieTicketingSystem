package service

import (
	"fmt"
	"net/http"

	"github.com/Coderx44/MovieTicketingPortal/booking"
	"github.com/Coderx44/MovieTicketingPortal/config"
	"github.com/gorilla/mux"
)

func initRouter(dep dependencies) (router *mux.Router) {
	v1 := fmt.Sprintf("application/vnd.%s.v1", config.AppName())
	_ = v1
	router = mux.NewRouter()
	router.HandleFunc("/ping/{id}", booking.ValidateJWT(booking.PingHandler, "admin")).Methods(http.MethodGet)
	router.HandleFunc("/create/user", booking.CreateNewUser(dep.BookingService)).Methods(http.MethodPost)
	router.HandleFunc("/create/admin", booking.CreateNewUser(dep.BookingService)).Methods(http.MethodPost)
	router.HandleFunc("/login", booking.Login(dep.BookingService)).Methods(http.MethodPost)
	router.HandleFunc("/movie", booking.ValidateJWT(booking.AddMovie(dep.BookingService), "admin")).Methods(http.MethodPost)
	router.HandleFunc("/multiplex", booking.ValidateJWT(booking.AddMultiplex(dep.BookingService), "admin")).Methods(http.MethodPost)
	router.HandleFunc("/multiplex/{id}/screen", booking.ValidateJWT(booking.AddScreen(dep.BookingService), "admin")).Methods(http.MethodPost)
	router.HandleFunc("/multiplex/{id}/show", booking.ValidateJWT(booking.AddShow(dep.BookingService), "admin")).Methods(http.MethodPost)
	router.HandleFunc("/multiplex/{city}", booking.GetAllMultiplexesByCity(dep.BookingService)).Methods(http.MethodGet)
	router.HandleFunc("/shows/movie", booking.GetAllShowsByMovieAndDate(dep.BookingService)).Methods(http.MethodGet)
	router.HandleFunc("/shows", booking.GetAllShowsByDateAndMultiplexId(dep.BookingService)).Methods(http.MethodGet)
	router.HandleFunc("/seats/{show_id}", booking.GetAllSeatsByShowID(dep.BookingService)).Methods(http.MethodGet)
	router.HandleFunc("/bookseats", booking.ValidateJWT(booking.BookSeats(dep.BookingService), "user")).Methods(http.MethodPost)
	router.HandleFunc("/movies/new", booking.GetUpcomingMovies(dep.BookingService)).Methods(http.MethodGet)
	router.HandleFunc("/movie/{title}", (booking.GetMovieByTitle(dep.BookingService))).Methods(http.MethodGet)
	return
}
