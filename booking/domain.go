package booking

import (
	"time"

	"github.com/golang-jwt/jwt"
)

type NewUser struct {
	User_id     int    `json:"user_id"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	Password    string `json:"password"`
	PhoneNumber string `json:"phone_number"`
	Role        string `json:"role"`
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
	Movie_id     int       `json:"movie_id"`
	Title        string    `json:"title"`
	Language     string    `json:"language"`
	Poster       []byte    `json:"-"`
	Release_date time.Time `json:"release_date"`
	Genre        string    `json:"genre"`
	Duration     float64   `json:"duration"`
}

type NewScreen struct {
	Screen_number    int    `json:"screen"`
	Total_seats      int    `json:"total_seats"`
	Sound_system     string `json:"sound_system"`
	Screen_dimension string `json:"screen_dimension"`
	Multiplex_id     int    `json:"muliplex_id"`
}

type NewMultiplex struct {
	Multiplex_id  int    `json:"multiplex_id"`
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
	Location_id int    `json:"location_id"`
	City        string `json:"city"`
	State       string `json:"state"`
	Pincode     int    `json:"pincode"`
}

type NewShow struct {
	Show_id      int    `json:"show_id"`
	Date         string `json:"show_date"`
	Start_time   string `json:"start_time"`
	End_time     string `json:"end_time"`
	Movie        string `json:"movie"`
	Screen       int    `json:"screen"`
	Screen_id    int    `json:"screen_id"`
	Movie_id     int    `json:"movie_id"`
	Multiplex_id int    `json:"multiplex_id"`
}

type MultiplexShow struct {
	Title          string    `json:"title"`
	Multiplex_name string    `json:"multiplex_name"`
	Language       string    `json:"language"`
	Duration       string    `json:"duration"`
	Genre          string    `json:"genre"`
	Movie_id       string    `json:"movie_id"`
	Show_id        string    `json:"show_id"`
	Start_time     time.Time `json:"show_time"`
	Locality       string    `json:"locality"`
	Date           time.Time `json:"show_date"`
}

type Seats struct {
	Seat_id     int    `json:"seat_id"`
	Seat_number int    `json:"seat_number"`
	Price       int    `json:"price"`
	Status      string `json:"status"`
	Show_id     int    `json:"show_id"`
}

type Invoice struct {
	Email          string    `json:"email"`
	Movie          string    `json:"movie"`
	Language       string    `json:"language"`
	Screen         string    `json:"screen"`
	Start_time     time.Time `json:"start_time"`
	Duration       string    `json:"duration"`
	Seats          []int     `json:"seats"`
	Total_price    int       `json:"price"`
	Multiplex_name string    `json:"multiplex"`
	Localtiy       string    `json:"locality"`
}
