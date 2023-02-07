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
	// GetUserByName(ctx context.Context, name string) (u User, err error)
	// GetMultiplexesByCity(ctx context.Context, city string) (m Multiplexes, err error)
	// GetMultiplexesByName(ctx context.Context, name string) (m Multiplexes, err error)
	// GetMovieByTitle(ctx context.Context, title string) (m Movies, err error)
	// GetShowByMultiplexID(ctx context.Context, multiplex_id int) (s []Shows, err error)
	// GetScreenByID(ctx context.Context, id int) (s Screens, err error)
	// GetSeatsByShowID(ctx context.Context, id int) (s []Seats, err error)
	// getScreenTypeByClass(ctx context.Context, typee string) (st Screen_types, err error)
	// StartBooking(ctx context.Context, no_of_seats int) (err error)
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
