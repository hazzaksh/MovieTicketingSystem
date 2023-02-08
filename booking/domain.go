package booking

import (
	"github.com/golang-jwt/jwt"
)

type NewUser struct {
	Name         string `json:"name"`
	Email        string `json:"email"`
	Password     string `json:"password"`
	Phone_number string `json:"phone_number"`
	Role         string `json:"role"`
}

type NewUserResponse struct {
	Name         string `json:"name"`
	Email        string `json:"email"`
	Password     string `json:"password"`
	Phone_number string `json:"phone_number"`
}

type Authentication struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Claims struct {
	Role  string `json:"role"`
	Email string `json:"email"`
	jwt.StandardClaims
}

type LoginResp struct {
	Token string `json:"token"`
	Mssg  string `json:"message"`
}

type NewMovie struct {
	Title        string  `json:"title"`
	Language     string  `json:"language"`
	Release_date string  `json:"release_date"`
	Genre        string  `json:"genre"`
	Duration     float64 `json:"duration"`
}

type NewScreen struct {
	Screen_number    int    `json:"screen"`
	Total_seats      int    `json:"total_seats"`
	Sound_system     string `json:"sound_system"`
	Screen_dimension string `json:"screen_dimension"`
	Multiplex_id     int    `json:"muliplex_id"`
}

type NewMultiplex struct {
	Name          string `json:"name"`
	Contact       string `json:"contact"`
	Total_screens int    `json:"total_screens"`
	Locality      string `json:"locality"`
	City          string `json:"city"`
	State         string `json:"state"`
	Pincode       int    `json:"pincode"`
	Location_id   int    `json:"location_id"`
}

type NewLocation struct {
	City    string `json:"city"`
	State   string `json:"state"`
	Pincode int    `json:"pincode"`
}

type NewShow struct {
	Date         string `json:"show_date"`
	Start_time   string `json:"start_time"`
	End_time     string `json:"end_time"`
	Movie        string `json:"movie"`
	Screen       int    `json:"screen"`
	Screen_id    int    `json:"screen_id"`
	Movie_id     int    `json:"movie_id"`
	Multiplex_id int    `json:"multiplex_id"`
}
