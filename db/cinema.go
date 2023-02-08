package db

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/pkg/errors"
)

const (
	CreateUserQuery      = `INSERT INTO USERS(name, password, email, phone_number, role) VALUES ($1, $2, $3, $4, $5 ) returning user_id`
	getUserByEmail       = `SELECT * FROM users WHERE email=$1`
	AddMovieQuery        = `INSERT INTO MOVIES(title, language, release_date, genre, duration) VALUES ($1, $2, $3, $4, $5) returning movie_id`
	getMultiplexesByName = `Select * FROM multiplexes WHERE name=$1`
	AddScreenQuery       = `INSERT INTO SCREENS (screen_number, total_seats, sound_system, screen_dimension, multiplex_id) VALUES ($1, $2, $3, $4, $5) returning screen_id`
	AddLocationQuery     = `INSERT INTO LOCATIONS (city, state, pincode) VALUES ($1, $2, $3) returning location_id`
	AddMultiplexQuery    = `INSERT INTO MULTIPLEXES (name, contact, total_screens, locality, location_id) VALUES ($1, $2, $3, $4, $5) returning multiplex_id`
	getLocationIdByCity  = `SELECT location_id from locations WHERE city=$1`
	getMultiplexeByID    = `Select multiplex_id FROM multiplexes WHERE multiplex_id=$1`
	AddShowQuery         = `INSERT INTO shows (show_date, start_time, end_time, screen_id, movie_id, multiplex_id)
	SELECT $1, $2, $3, $4, $5, $6
	WHERE NOT EXISTS (
		SELECT 1 FROM shows
		WHERE show_date = $1
		AND screen_id = $4
		AND (
			($2 BETWEEN start_time AND end_time)
			OR ($3 BETWEEN start_time AND end_time)
			OR (start_time BETWEEN $2 AND $3)
			OR (end_time BETWEEN $2 AND $3)
		)
	)
	RETURNING show_id;
	`
	getScreenByNumberAndMultiplexID = `Select * From screens WHERE screen_number=$1 and multiplex_id=$2`
	getMovieByTitle                 = `Select movie_id From MOVIES where title=$1`
	AddSeatsQuery                   = `INSERT INTO SEATS (seat_number, price, show_id, status) VALUES ($1, $2, $3, $4)`
)

type User struct {
	User_id     int    `json:"user_id" db:"user_id"`
	Name        string `json:"name" db:"name"`
	Email       string `json:"email" db:"email"`
	Password    string `json:"-" db:"password"`
	PhoneNumber string `json:"phone_number" db:"phone_number"`
	Role        string `json:"role" db:"role"`
}

type Movie struct {
	Movie_id     int       `json:"movie_id" db:"movie_id"`
	Title        string    `json:"title" db:"title"`
	Language     string    `json:"language" db:"language"`
	Poster       []byte    `json:"poster" db:"poster"`
	Release_date time.Time `json:"release_date" db:"release_date"`
	Genre        string    `json:"genre" db:"genre"`
	Duration     float64   `json:"duration" db:"duration"`
}

type Location struct {
	Location_id int    `json:"location_id" db:"location_id"`
	City        string `json:"city" db:"city"`
	State       string `json:"state" db:"state"`
	Pincode     int    `json:"pincode" sb:"pincode"`
}

type Multiplexe struct {
	Multiplex_id  int    `json:"multiplex_id" db:"multiplex_id"`
	Name          string `json:"name" db:"name"`
	Contact       string `json:"contact" db:"contact"`
	Total_screens int    `json:"total_screens" db:"total_screens"`
	Locality      string `json:"locality" db:"locality"`
	Location_id   int    `json:"location_id" db:"location_id"`
}

type Screen struct {
	Screen_id        int    `json:"screen_id" db:"screen_id"`
	Screen_number    int    `json:"screen_number" db:"screen_number"`
	Total_seats      int    `json:"total_seats" db:"total_seats"`
	Sound_system     string `json:"sound_system" db:"sound_system"`
	Screen_dimension string `json:"screen_dimension" db:"screen_dimension"`
	Multiplex_id     int    `json:"multiplex_id" db:"multiplex_id"`
}

type Show struct {
	Show_id      int       `json:"show_id" db:"show_id"`
	Show_date    time.Time `json:"show_date" db:"show_date"`
	Start_time   time.Time `json:"start_time" db:"start_time"`
	End_time     time.Time `json:"end_time" db:"end_time"`
	Screen_id    int       `json:"screen_id" db:"screen_id"`
	Movie_id     int       `json:"movie_id" db:"movie_id"`
	Multiplex_id int       `json:"multiplex_id" db:"multiplex_id"`
}

type Seat struct {
	Seat_id int    `json:"seat_id" db:"seat_id"`
	Price   int    `json:"price" db:"price"`
	Status  string `json:"status" db:"status"`
	Show_id int    `json:"show_id" db:"show_id"`
}

type Booking struct {
	Booking_id int    `json:"booking_id" db:"booking_id"`
	Price      int    `json:"price" db:"price"`
	Status     string `json:"status" db:"status"`
	User_id    int    `json:"user_id" db:"user_id"`
	Seat_id    int    `json:"seat_id" db:"seat_id"`
	Show_id    int    `json:"show_id" db:"show_id"`
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

func (s *store) AddMovie(ctx context.Context, m Movie) (movie_id uint, err error) {
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
		if err := s.db.GetContext(ctx, &movie_id, AddMovieQuery, m.Title, m.Language, m.Release_date, m.Genre, m.Duration); err != nil {
			return err
		}
		return nil

	})
	return

}

func (s *store) GetMultiplexesByName(ctx context.Context, name string) (m Multiplexe, err error) {

	err = WithDefaultTimeout(ctx, func(ctx context.Context) error {
		err = s.db.GetContext(ctx, &m, getUserByEmail, name)
		return err
	})

	if err == sql.ErrNoRows {
		return m, errors.New("multiplex doesn't exist.")
	}
	return

}
func (s *store) AddScreen(ctx context.Context, sn Screen) (screen_id uint, err error) {

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
		if err := s.db.GetContext(ctx, &screen_id, AddScreenQuery, sn.Screen_number, sn.Total_seats, sn.Sound_system, sn.Screen_dimension, sn.Multiplex_id); err != nil {
			return err
		}
		return nil

	})

	return

}

func (s *store) AddLocation(ctx context.Context, l Location) (location_id uint, err error) {

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
		if err := s.db.GetContext(ctx, &location_id, AddLocationQuery, l.City, l.State, l.Pincode); err != nil {
			return err
		}
		return nil

	})

	return

}

func (s *store) AddMultiplex(ctx context.Context, m Multiplexe) (muliplex_id uint, err error) {

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
		if err := s.db.GetContext(ctx, &muliplex_id, AddMultiplexQuery, m.Name, m.Contact, m.Total_screens, m.Locality, m.Location_id); err != nil {
			return err
		}
		return nil

	})

	return
}

func (s *store) GetLocationIdByCity(ctx context.Context, city string) (location_id uint, err error) {

	err = WithDefaultTimeout(ctx, func(ctx context.Context) error {
		err = s.db.GetContext(ctx, &location_id, getLocationIdByCity, city)
		return err
	})
	log.Println(city)
	if err == sql.ErrNoRows {
		return location_id, errors.New("location doesn't exist")
	}
	return
}

func (s *store) GetMultiplexeByID(ctx context.Context, id int) (m_id uint, err error) {

	err = WithDefaultTimeout(ctx, func(ctx context.Context) error {
		err = s.db.GetContext(ctx, &m_id, getMultiplexeByID, id)
		return err
	})

	if err == sql.ErrNoRows {
		return m_id, errors.New("multiplex doesn't exist.")
	}
	return

}

func (s *store) AddShow(ctx context.Context, sh Show) (show_id uint, err error) {

	tx, err := s.db.BeginTxx(ctx, &sql.TxOptions{})
	if err != nil {
		return
	}
	log.Println(sh)

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
	log.Println(sh.Start_time, sh.End_time)
	ctxWithTx := newContext(ctx, tx)
	err = WithDefaultTimeout(ctxWithTx, func(ctx context.Context) error {
		err := s.db.GetContext(ctx, &show_id, AddShowQuery, sh.Show_date, sh.Start_time, sh.End_time, sh.Screen_id, sh.Movie_id, sh.Multiplex_id)

		if err != nil {
			log.Println(err, show_id)
			return err
		}
		return nil

	})

	return

}

func (s *store) GetScreenByNumberAndMultiplexID(ctx context.Context, s_no int, m_id int) (sn Screen, err error) {

	err = WithDefaultTimeout(ctx, func(ctx context.Context) error {
		err = s.db.GetContext(ctx, &sn, getScreenByNumberAndMultiplexID, s_no, m_id)
		log.Println("select seat_id", err)
		return err
	})

	if err == sql.ErrNoRows {
		return sn, errors.New("screen doesn't exist")
	}
	return
}

func (s *store) GetMovieByTitle(ctx context.Context, title string) (movie_id uint, err error) {

	err = WithDefaultTimeout(ctx, func(ctx context.Context) error {
		err = s.db.GetContext(ctx, &movie_id, getMovieByTitle, title)
		return err
	})

	if err == sql.ErrNoRows {
		return movie_id, errors.New("movie doesn't exist")
	}
	return

}

func (s *store) AddSeats(ctx context.Context, num_of_seats int, show_id int) (err error) {

	for i := 0; i < num_of_seats; i++ {
		err = WithDefaultTimeout(ctx, func(ctx context.Context) error {
			_, err = s.db.ExecContext(ctx, AddSeatsQuery, i+1, 300, show_id, "Available")
			return err
		})
	}
	return
}
