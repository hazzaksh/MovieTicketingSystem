package db

import (
	"context"
	"database/sql"
	"time"

	"github.com/pkg/errors"
)

const (
	CreateUserQuery = `INSERT INTO USERS(name, password, email, phone_number, role) VALUES ($1, $2, $3, $4, $5 ) returning user_id`
	getUserByEmail  = `SELECT * FROM users WHERE email=$1`
)

type User struct {
	ID          int    `json:"user_id" db:"user_id"`
	Name        string `json:"name" db:"name"`
	Email       string `json:"email" db:"email"`
	Password    string `json:"-" db:"password"`
	PhoneNumber string `json:"phone_number" db:"phone_number"`
	Role        string `json:"role" db:"role"`
}

type Movies struct {
	ID           int       `json:"movie_id" db:"movie_id"`
	Title        string    `json:"title" db:"title"`
	Language     string    `json:"language" db:"language"`
	Poster       []byte    `json:"poster" db:"poster"`
	Release_date time.Time `json:"release_date" db:"release_date"`
}

type Locations struct {
	Location_id int    `json:"location_id" db:"location_id"`
	City        string `json:"city" db:"city"`
	State       string `json:"state" db:"state"`
	Pincode     int    `json:"pincode" sb:"pincode"`
}

type Multiplexes struct {
	Multiplex_id  int    `json:"multiplex_id" db:"multiplex_id"`
	Name          string `json:"name" db:"name"`
	Contact       string `json:"contact" db:"contact"`
	Total_screens int    `json:"total_screens" db:"total_screens"`
	Locality      string `json:"locality" db:"locality"`
	Location_id   int    `json:"location_id" db:"location_id"`
}

type Screens struct {
	Screen_id        int    `json:"screen_id" db:"screen_id"`
	Total_seats      int    `json:"total_seats" db:"total_seats"`
	Sound_system     string `json:"sound_system" db:"sound_system"`
	Screen_dimension string `json:"screen_dimension" db:"screen_dimension"`
	Multiplex_id     int    `json:"multiplex_id" db:"multiplex_id"`
}

type Shows struct {
	Show_id      int         `json:"show_id" db:"show_id"`
	Show_date    time.Time   `json:"show_date" db:"show_date"`
	Start_time   time.Time   `json:"start_time" db:"start_time"`
	End_time     time.Ticker `json:"end_time" db:"end_time"`
	Screen_id    int         `json:"screen_id" db:"screen_id"`
	Movie_id     int         `json:"movie_id" db:"movie_id"`
	Multiplex_id int         `json:"multiplex_id" db:"multiplex_id"`
}

type Screen_types struct {
	Screen_type_id int    `json:"screen_type_id" db:"screen_type_id"`
	Seat_number    int    `json:"seat_number" db:"seat_number"`
	Class          string `json:"class" db:"class"`
	Screen_id      int    `json:"screen_id" db:"screen_id"`
}

type Seats struct {
	Seat_id        int    `json:"seat_id" db:"seat_id"`
	Price          int    `json:"price" db:"price"`
	Status         string `json:"status" db:"status"`
	Screen_type_id int    `json:"screen_type_id" db:"screen_type_id"`
}

type Bookings struct {
	Booking_id      int       `json:"booking_id" db:"booking_id"`
	Number_of_seats int       `json:"number_of_seats" db:"number_of_seats"`
	Status          string    `json:"status" db:"status"`
	Expiry          time.Time `json:"expiry" db:"expiry"`
	User_id         int       `json:"user_id" db:"user_id"`
	Seat_id         int       `json:"seat_id" db:"seat_id"`
}

type Transactions struct {
	Transaction_id int       `json:"transaction_id" db:"transaction_id"`
	Price          int       `json:"price" db:"price"`
	Time_stamp     time.Time `json:"time_stamp" db:"time_stamp"`
	Booking_id     int       `json:"booking_id" db:"booking_id"`
}

func (s *store) CreateUser(ctx context.Context, u User) (user_id uint, err error) {
	tx, err := s.db.BeginTxx(ctx, &sql.TxOptions{})
	if err != nil {
		return
	}

	defer func() {
		if err != nil {
			e := tx.Rollback()
			if e != nil {
				err = errors.WithStack(e)
				return
			}
		}
		tx.Commit()
	}()

	ctxWithTx := newContext(ctx, tx)
	err = WithDefaultTimeout(ctxWithTx, func(ctx context.Context) error {
		if err := s.db.GetContext(ctx, &user_id, CreateUserQuery, u.Name, u.Password, u.Email, u.PhoneNumber, u.Role); err != nil {
			return err
		}
		return nil

	})
	return
}

func (s *store) GetUserByEmail(ctx context.Context, email string) (u User, err error) {

	err = WithDefaultTimeout(ctx, func(ctx context.Context) error {
		err = s.db.GetContext(ctx, &u, getUserByEmail, email)
		return err
	})

	if err == sql.ErrNoRows {
		return u, errors.New("user does not exist in db")
	}
	return

}
