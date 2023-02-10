package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/lib/pq"
	"github.com/pkg/errors"
)

const (
	CreateUserQuery                 = `INSERT INTO USERS(name, password, email, phone_number, role) VALUES ($1, $2, $3, $4, $5 ) returning user_id`
	getUserByEmail                  = `SELECT * FROM users WHERE email=$1`
	AddMovieQuery                   = `INSERT INTO MOVIES(title, language, release_date, genre, duration) VALUES ($1, $2, $3, $4, $5) returning movie_id`
	getMultiplexesByName            = `Select * FROM multiplexes WHERE name=$1`
	getAllMultiplexeByLocationID    = `SELECT * FROM MULTIPLEXES WHERE location_id=$1`
	AddScreenQuery                  = `INSERT INTO SCREENS (screen_number, total_seats, sound_system, screen_dimension, multiplex_id) VALUES ($1, $2, $3, $4, $5) returning screen_id`
	AddLocationQuery                = `INSERT INTO LOCATIONS (city, state, pincode) VALUES ($1, $2, $3) returning location_id`
	AddMultiplexQuery               = `INSERT INTO MULTIPLEXES (name, contact, total_screens, locality, location_id) VALUES ($1, $2, $3, $4, $5) returning multiplex_id`
	getLocationIdByCity             = `SELECT * from locations WHERE Lower(city)=Lower($1)`
	getMultiplexeByID               = `Select multiplex_id FROM multiplexes WHERE multiplex_id=$1`
	getAllShowsByMultiplexIDandDate = `SELECT movies.title AS movie_title, movies.language, movies.duration, movies.genre, movies.movie_id, shows.show_id, shows.start_time, multiplexes.locality, multiplexes.name, shows.show_date
										FROM movies JOIN shows ON movies.movie_id = shows.movie_id JOIN multiplexes ON shows.multiplex_id = multiplexes.multiplex_id WHERE shows.show_date = $1 AND multiplexes.multiplex_id = $2;`

	GetAllShowsByMovieAndDate = `SELECT movies.title AS movie_title, movies.language, movies.duration, movies.genre, movies.movie_id, shows.show_id, shows.start_time, multiplexes.locality, multiplexes.name, shows.show_date
									FROM multiplexes JOIN locations ON multiplexes.location_id = locations.location_id JOIN shows ON multiplexes.multiplex_id = shows.multiplex_id JOIN movies ON shows.movie_id = movies.movie_id WHERE locations.city = $1 AND movies.title = $2 AND shows.show_date = $3`

	AddShowQuery = `INSERT INTO shows (show_date, start_time, end_time, screen_id, movie_id, multiplex_id)
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
	getMovieByTitle                 = `Select * From MOVIES where title=$1`
	AddSeatsQuery                   = `INSERT INTO SEATS (seat_number, price, show_id, status) VALUES ($1, $2, $3, $4)`
	getAllSeatsByShowIDQuery        = `SELECT * FROM SEATS WHERE show_id=$1`
	// bookSeatsQuery                  = `WITH updated_seats AS (UPDATE seats SET status = 'sold' WHERE seat_id = ANY($1) AND status = 'Available' RETURNING * ) INSERT INTO bookings (status, email, seat_id, show_id)
	//   SELECT 'sold', $2, seat_id, show_id
	//   FROM updated_seats`
	bookSeatsQuery = `BEGIN;
	WITH available_seats AS (
	  SELECT seat_id, seat_number
	  FROM seats
	  WHERE seat_id = ANY($1) AND status = 'available'
	)
	UPDATE seats
	SET status = 'sold'
	WHERE seat_id IN (SELECT seat_id FROM available_seats)
	RETURNING seat_id;
	
	INSERT INTO bookings (status, email, seats, show_id)
	VALUES ('sold', $2, (SELECT array_agg(seat_number) FROM available_seats), $3);
	
	COMMIT;`
	checkIfAvailable  = `WITH seats_status AS (SELECT seat_id, status FROM seats WHERE seat_id = ANY($1)) SELECT count(*) FROM seats_status WHERE status = 'Available' HAVING count(*) = (SELECT count(*) FROM seats_status)`
	getSeatsByID      = `SELECT * FROM seats WHERE seat_id = ANY($1)`
	getInvoiceDetails = `SELECT m.title, m.language, s.screen_number, sh.start_time, m.duration, mu.name as multiplex_name, mu.locality
	FROM shows sh
	INNER JOIN screens s ON sh.screen_id = s.screen_id
	INNER JOIN movies m ON sh.movie_id = m.movie_id
	INNER JOIN multiplexes mu ON sh.multiplex_id = mu.multiplex_id
	WHERE sh.show_id = $1;
	`
	getUpcomingMovies = `SELECT * FROM MOVIES WHERE release_date > $1`
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
	Email      int    `json:"Email" db:"email"`
	Seat_id    int    `json:"seat_id" db:"seat_id"`
	Show_id    int    `json:"show_id" db:"show_id"`
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

func (s *store) AddLocation(ctx context.Context, l Location) (location_id int, err error) {

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

func (s *store) GetAllMultiplexesByLocationID(ctx context.Context, location_id int) (m []Multiplexe, err error) {
	var rows *sql.Rows
	err = WithDefaultTimeout(ctx, func(ctx context.Context) error {
		rows, err = s.db.QueryContext(ctx, getAllMultiplexeByLocationID, location_id)
		return err
	})
	defer rows.Close()
	for rows.Next() {
		var multiplex Multiplexe
		err = rows.Scan(&multiplex.Multiplex_id, &multiplex.Name, &multiplex.Contact, &multiplex.Total_screens, &multiplex.Locality, &multiplex.Location_id)
		if err != nil {
			return
		}
		m = append(m, multiplex)
	}
	log.Println(m, err)
	if err = rows.Err(); err != nil {
		return
	}
	if err == sql.ErrNoRows {
		return m, errors.New("No multiplexes found.")
	}
	return
}

func (s *store) GetLocationIdByCity(ctx context.Context, city string) (location Location, err error) {

	err = WithDefaultTimeout(ctx, func(ctx context.Context) error {
		err = s.db.GetContext(ctx, &location, getLocationIdByCity, city)
		return err
	})
	log.Println(city)
	if err == sql.ErrNoRows {
		return location, errors.New("location doesn't exist")
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
	// log.Println(sh.Start_time, sh.End_time)

	ctxWithTx := newContext(ctx, tx)
	err = WithDefaultTimeout(ctxWithTx, func(ctx context.Context) error {
		err := s.db.GetContext(ctx, &show_id, AddShowQuery, sh.Show_date, sh.Start_time, sh.End_time, sh.Screen_id, sh.Movie_id, sh.Multiplex_id)

		if err != nil {
			log.Println("error in add show:", err)
			return err
		}
		return nil

	})

	return

}

func (s *store) GetScreenByNumberAndMultiplexID(ctx context.Context, s_no int, m_id int) (sn Screen, err error) {

	err = WithDefaultTimeout(ctx, func(ctx context.Context) error {
		err = s.db.GetContext(ctx, &sn, getScreenByNumberAndMultiplexID, s_no, m_id)

		return err
	})

	if err == sql.ErrNoRows {
		return sn, errors.New("screen doesn't exist")
	}
	return
}

func (s *store) GetMovieByTitle(ctx context.Context, title string) (m Movie, err error) {

	err = WithDefaultTimeout(ctx, func(ctx context.Context) error {
		err = s.db.GetContext(ctx, &m, getMovieByTitle, title)
		return err
	})

	if err != nil && err == sql.ErrNoRows {
		return m, errors.New("movie doesn't exist")
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

func (s *store) GetAllShowsByDateAndMultiplexId(ctx context.Context, date time.Time, multiplex_id int) (m []MultiplexShow, err error) {

	var rows *sql.Rows
	err = WithDefaultTimeout(ctx, func(ctx context.Context) error {
		rows, err = s.db.QueryContext(ctx, getAllShowsByMultiplexIDandDate, date, multiplex_id)
		log.Println("ff", err)
		return err
	})
	defer rows.Close()
	for rows.Next() {
		var mShow MultiplexShow
		err = rows.Scan(&mShow.Title, &mShow.Language, &mShow.Duration, &mShow.Genre, &mShow.Movie_id, &mShow.Show_id, &mShow.Start_time, &mShow.Locality, &mShow.Multiplex_name, &mShow.Date)
		if err != nil {
			return
		}
		m = append(m, mShow)
	}

	err = rows.Err()
	if err != nil && err == sql.ErrNoRows {
		return m, errors.New("No shows found.")
	}
	return

}
func (s *store) GetAllShowsByMovieAndDate(ctx context.Context, title string, city string, date time.Time) (m []MultiplexShow, err error) {

	var rows *sql.Rows
	err = WithDefaultTimeout(ctx, func(ctx context.Context) error {
		rows, err = s.db.QueryContext(ctx, GetAllShowsByMovieAndDate, city, title, date)
		log.Println("ff", err)
		return err
	})
	defer rows.Close()
	for rows.Next() {
		var mShow MultiplexShow
		err = rows.Scan(&mShow.Title, &mShow.Language, &mShow.Duration, &mShow.Genre, &mShow.Movie_id, &mShow.Show_id, &mShow.Start_time, &mShow.Locality, &mShow.Multiplex_name, &mShow.Date)
		if err != nil {
			return
		}
		m = append(m, mShow)
	}

	err = rows.Err()
	if err != nil && err == sql.ErrNoRows {
		return m, errors.New("No shows found.")
	}
	return

}

func (s *store) GetSeatsByShowID(ctx context.Context, id int) (seats []Seats, err error) {

	var rows *sql.Rows
	err = WithDefaultTimeout(ctx, func(ctx context.Context) error {
		rows, err = s.db.QueryContext(ctx, getAllSeatsByShowIDQuery, id)
		log.Println("ff", err)
		return err
	})
	defer rows.Close()
	for rows.Next() {
		var mShow Seats
		err = rows.Scan(&mShow.Seat_id, &mShow.Seat_number, &mShow.Price, &mShow.Status, &mShow.Show_id)
		if err != nil {
			return
		}
		seats = append(seats, mShow)
	}

	err = rows.Err()
	if err != nil && err == sql.ErrNoRows {
		return seats, errors.New("No seats found.")
	}
	return

}

func (s *store) AddBookingsBySeatId(ctx context.Context, seats []int, email string, show_id int, seat_num []int) (err error) {

	updateStmt, err := s.db.Prepare("UPDATE seats SET status = 'sold' WHERE seat_id = $1")
	if err != nil {

		return
	}

	defer updateStmt.Close()

	insertStmt, err := s.db.Prepare("INSERT INTO bookings (status, email, seat_id, show_id) VALUES ('sold', $1, $2, $3) RETURNING booking_id")
	if err != nil {
		// Handle error
		return
	}
	defer insertStmt.Close()

	tx, err := s.db.Begin()
	if err != nil {

		return
	}

	for _, seatID := range seats {

		_, err = tx.Stmt(updateStmt).Exec(seatID)
		if err != nil {

			tx.Rollback()
			return
		}
	}

	var bookingID int
	err = tx.Stmt(insertStmt).QueryRow(email, pq.Array(seat_num), show_id).Scan(&bookingID)
	if err != nil {

		tx.Rollback()
		return
	}

	err = tx.Commit()
	if err != nil {

		return
	}
	return
}
func (s *store) CheckAvailability(ctx context.Context, seats []int) (bool, error) {
	// var status int
	seatIDsString := "{"
	for i, seatID := range seats {
		seatIDsString += fmt.Sprintf("%d", seatID)
		if i < len(seats)-1 {
			seatIDsString += ","
		}
	}
	seatIDsString += "}"
	log.Println(pq.Array(seats), "ogseats:", seats)

	var count int
	err := WithDefaultTimeout(ctx, func(ctx context.Context) error {
		err := s.db.GetContext(ctx, &count, checkIfAvailable, pq.Array(seats))
		return err
	})
	// log.Println(err)
	// log.Println("status", count)
	if err != nil && err == sql.ErrNoRows {
		return false, err
	} else if err != nil {
		return false, err
	}

	return true, nil
}

func (s *store) GetSeatsByID(ctx context.Context, id []int) (seats []Seats, err error) {

	var rows *sql.Rows
	err = WithDefaultTimeout(ctx, func(ctx context.Context) error {
		rows, err = s.db.QueryContext(ctx, getSeatsByID, pq.Array(id))
		log.Println("ff", err)
		return err
	})
	err = rows.Err()
	if err != nil && err == sql.ErrNoRows {
		return seats, errors.New("Err fetching seats")
	}
	defer rows.Close()
	for rows.Next() {
		var mShow Seats
		err = rows.Scan(&mShow.Seat_id, &mShow.Seat_number, &mShow.Price, &mShow.Status, &mShow.Show_id)
		if err != nil {
			return
		}
		seats = append(seats, mShow)
	}

	return
}

func (s *store) GetInvoiceDetails(ctx context.Context, show_id int) (invoice Invoice, err error) {
	// log.Print(show_id)
	var rows *sql.Rows
	err = WithDefaultTimeout(ctx, func(ctx context.Context) error {
		rows, err = s.db.QueryContext(ctx, getInvoiceDetails, show_id)
		log.Println("ff", err)
		return err
	})
	err = rows.Err()

	if err != nil && err == sql.ErrNoRows {
		return invoice, errors.New("Err fetching invoice details")
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&invoice.Movie, &invoice.Language, &invoice.Screen, &invoice.Start_time, &invoice.Duration, &invoice.Multiplex_name, &invoice.Localtiy)
		if err != nil {
			return
		}
	}
	log.Println(invoice)

	return
}

func (s *store) GetUpcomingMovies(ctx context.Context, date string) (m []Movie, err error) {
	var rows *sql.Rows
	rDate, _ := time.Parse("2006-01-02", date)
	err = WithDefaultTimeout(ctx, func(ctx context.Context) error {
		rows, err = s.db.QueryContext(ctx, getUpcomingMovies, rDate)
		return err
	})
	// err = rows.Err()
	if err != nil && err == sql.ErrNoRows {
		return []Movie{{}}, errors.New("No movies available")
	}
	defer rows.Close()
	var movie Movie
	for rows.Next() {
		err = rows.Scan(&movie.Movie_id, &movie.Title, &movie.Language, &movie.Poster, &movie.Release_date, &movie.Genre, &movie.Duration)
		if err != nil {
			return
		}
		m = append(m, movie)
	}

	return

}

func (s *store) DeleteByBookingByID(ctx context.Context, id int) (err error) {

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec("UPDATE seats SET status = 'Available' WHERE (seat_number = ANY(ARRAY(SELECT seat_id FROM bookings WHERE booking_id = $1))) AND show_id = (SELECT show_id FROM bookings WHERE booking_id = $1)", id)
	if err != nil {
		return err
	}

	_, err = tx.Exec("UPDATE bookings SET status = 'Cancelled' WHERE booking_id = $1", id)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil

}
