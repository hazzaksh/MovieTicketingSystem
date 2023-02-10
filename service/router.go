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
	router.Use()
	router.HandleFunc("/pi/{id}", booking.ValidateJWT(booking.PingHandler)).Methods(http.MethodGet)
	router.HandleFunc("/create/user", booking.CreateNewUser(dep.BookingService)).Methods(http.MethodPost)
	router.HandleFunc("/create/admin", booking.CreateNewUser(dep.BookingService)).Methods(http.MethodPost)
	router.HandleFunc("/login", booking.Login(dep.BookingService)).Methods(http.MethodPost)
	router.HandleFunc("/movie/add", booking.ValidateJWT(booking.AddMovie(dep.BookingService))).Methods(http.MethodPost)
	router.HandleFunc("/multiplex", booking.ValidateJWT(booking.AddMultiplex(dep.BookingService))).Methods(http.MethodPost)
	router.HandleFunc("/multiplex/{id}/screen", booking.ValidateJWT(booking.AddScreen(dep.BookingService))).Methods(http.MethodPost)
	router.HandleFunc("/multiplex/{id}/show", booking.ValidateJWT(booking.AddShow(dep.BookingService))).Methods(http.MethodPost)

	return
}
