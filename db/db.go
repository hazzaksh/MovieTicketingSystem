package db

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
)

type ctxKey int

const (
	dbKey          ctxKey = 0
	defaultTimeout        = 1 * time.Second
)

type Storer interface {
	CreateUser(ctx context.Context, u User) (user_id uint, err error)
	GetUserByEmail(ctx context.Context, email string) (u User, err error)
	AddMovie(ctx context.Context, m Movie) (movie_id uint, err error)
	AddScreen(ctx context.Context, m Screen) (screen_id uint, err error)
	// GetMultiplexesByName(ctx context.Context, name string) (m Multiplexe, err error)
	GetMultiplexeByID(ctx context.Context, id int) (m_id uint, err error)
	AddMultiplex(ctx context.Context, m Multiplexe) (multiplex_id uint, err error)
	AddLocation(ctx context.Context, l Location) (location_id int, err error)
	GetLocationIdByCity(ctx context.Context, city string) (l Location, err error)
	AddShow(ctx context.Context, s Show) (show_id uint, err error)
	GetScreenByNumberAndMultiplexID(ctx context.Context, s_no int, m_id int) (s Screen, err error)
	GetMovieByTitle(ctx context.Context, title string) (m Movie, err error)
	AddSeats(ctx context.Context, num_of_seats int, show_id int) (err error)
	GetAllMultiplexesByLocationID(ctx context.Context, location_id int) (m []Multiplexe, err error)
	GetAllShowsByDateAndMultiplexId(ctx context.Context, date time.Time, multiplex_id int) (m []MultiplexShow, err error)
	GetAllShowsByMovieAndDate(ctx context.Context, title string, city string, date time.Time) (m []MultiplexShow, err error)
	GetSeatsByShowID(ctx context.Context, id int) (s []Seats, err error)
	AddBookingsBySeatId(ctx context.Context, seats []int, email string, show_id int, seat_num []int) (err error)
	CheckAvailability(ctx context.Context, seats []int) (bool, error)
	GetSeatsByID(ctx context.Context, id []int) (seats []Seats, err error)
	GetInvoiceDetails(ctx context.Context, show_id int) (invoice Invoice, err error)
	GetUpcomingMovies(ctx context.Context, date string) (m []Movie, err error)
	DeleteByBookingByID(ctx context.Context, id int) (err error)
}

type store struct {
	db *sqlx.DB
}

func NewStorer(d *sqlx.DB) Storer {
	return &store{
		db: d,
	}
}

func newContext(ctx context.Context, tx *sqlx.Tx) context.Context {
	return context.WithValue(ctx, dbKey, tx)
}

func WithTimeout(ctx context.Context, timeout time.Duration, op func(ctx context.Context) error) (err error) {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	return op(ctxWithTimeout)
}

func WithDefaultTimeout(ctx context.Context, op func(ctx context.Context) error) (err error) {
	return WithTimeout(ctx, defaultTimeout, op)
}
