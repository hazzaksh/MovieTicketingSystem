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
	router.HandleFunc("/ping", booking.PingHandler).Methods(http.MethodGet)
	router.HandleFunc("/create/user", booking.CreateNewUser(dep.BookingService)).Methods(http.MethodPost)
	router.HandleFunc("/create/admin", booking.CreateNewUser(dep.BookingService)).Methods(http.MethodPost)
	router.HandleFunc("/login", booking.Login(dep.BookingService)).Methods(http.MethodPost)
	router.HandleFunc("/movie/add", booking.AddMovie(dep.BookingService))
	return
}
