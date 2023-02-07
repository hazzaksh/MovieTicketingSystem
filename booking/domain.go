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
