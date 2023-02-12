package booking

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"testing"
	"time"

	"github.com/Coderx44/MovieTicketingPortal/db"
	"github.com/Coderx44/MovieTicketingPortal/db/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

type BookingServiceTestSuite struct {
	suite.Suite
	logger  *zap.SugaredLogger
	storer  *mocks.Storer
	service Service
}

func (suite *BookingServiceTestSuite) SetupTest() {
	suite.storer = &mocks.Storer{}
	suite.service = NewBookingService(suite.storer, suite.logger)
}

func (suite *BookingServiceTestSuite) TearDownTest() {
	t := suite.T()
	suite.storer.AssertExpectations(t)
}
func TestHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(BookingServiceTestSuite))
}

func (suite *BookingServiceTestSuite) TestCreateUser() {

	t := suite.T()
	type args struct {
		ctx     context.Context
		dbUser  db.User
		newUser NewUser
	}

	tests := []struct {
		name    string
		args    args
		wantErr error
		user_id uint
		prepare func(args, *mocks.Storer)
	}{
		{
			name: "success",
			args: args{
				ctx: context.TODO(),
				dbUser: db.User{
					Name:        "test1",
					Email:       "test1@test.com",
					Password:    "test1",
					PhoneNumber: "1234567890",
					Role:        "user",
				},
				newUser: NewUser{
					Name:        "test1",
					Email:       "test1@test.com",
					Password:    "test1",
					PhoneNumber: "1234567890",
					Role:        "user",
				},
			},
			wantErr: nil,
			user_id: 1,
			prepare: func(a args, s *mocks.Storer) {
				s.On("CreateUser", context.TODO(), a.dbUser).Return(uint(1), nil)
			},
		},
		{
			name: "failure",
			args: args{
				ctx: context.TODO(),
				dbUser: db.User{
					Name:        "test2",
					Email:       "test1@test.com",
					Password:    "test2",
					PhoneNumber: "1234565590",
					Role:        "user",
				},
				newUser: NewUser{
					Name:        "test2",
					Email:       "test1@test.com",
					Password:    "test2",
					PhoneNumber: "1234565590",
					Role:        "user",
				},
			},
			wantErr: errors.New("user exists for the given email"),
			user_id: 0,
			prepare: func(a args, s *mocks.Storer) {
				s.On("CreateUser", context.TODO(), a.dbUser).Return(uint(0), errors.New("pq: duplicate key value violates unique constraint \"users_email_key\"")).Once()
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.prepare(tt.args, suite.storer)
			u_id, err := suite.service.CreateNewUser(tt.args.ctx, tt.args.newUser)
			if tt.wantErr != nil {
				suite.ErrorContains(err, "user exists for the given email")
			} else {
				suite.ErrorIs(err, nil)
			}

			assert.Equal(t, tt.user_id, u_id)

		})
	}

}

func (suite *BookingServiceTestSuite) TestLogin() {

	t := suite.T()
	type args struct {
		ctx  context.Context
		auth Authentication
	}

	tests := []struct {
		name        string
		args        args
		wantErr     error
		tokenString string
		prepare     func(args, *mocks.Storer)
	}{
		{
			name: "Login success",
			args: args{
				ctx: context.TODO(),
				auth: Authentication{
					Email:    "test1@gmail.com",
					Password: "pass1",
				},
			},
			wantErr:     nil,
			tokenString: "token",

			prepare: func(a args, s *mocks.Storer) {

				s.On("GetUserByEmail", mock.Anything, "test1@gmail.com").Return(db.User{
					Name:        "test1",
					Email:       "test1@gmail.com",
					Password:    "$2a$10$hu89BgmHArw0/s2EKfCs1OdmZ9y2nmjXLXO59PGKS3DUzQn.8H6jC",
					PhoneNumber: "1234567890",
					Role:        "user",
				}, nil).Once()

			},
		},
		{
			name: "Login failure",
			args: args{
				ctx: context.TODO(),
				auth: Authentication{
					Email:    "test2@gmail.com",
					Password: "pass1",
				},
			},
			wantErr:     errors.New("unauthorized"),
			tokenString: "",
			prepare: func(a args, s *mocks.Storer) {
				s.On("GetUserByEmail", mock.Anything, a.auth.Email).Return(db.User{}, errors.New("user does not exist in db")).Once()

			},
		},
		{
			name: "Login failure (Password incorrect)",
			args: args{
				ctx: context.TODO(),
				auth: Authentication{
					Email:    "test2@gmail.com",
					Password: "pass1",
				},
			},
			wantErr:     errors.New("username or password is incorrect"),
			tokenString: "",
			prepare: func(a args, s *mocks.Storer) {
				s.On("GetUserByEmail", mock.Anything, a.auth.Email).Return(db.User{
					Password: "pass2",
				}, nil).Once()

			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.prepare(tt.args, suite.storer)
			token, err := suite.service.Login(tt.args.ctx, tt.args.auth)
			if tt.wantErr != nil {
				suite.ErrorContains(err, "unauthorized")
			} else {
				suite.ErrorIs(err, nil)
				if token == "" {
					log.Fatalf("Token empty")
				}
			}

		})

	}
}

func (suite *BookingServiceTestSuite) TestGetAllMultiplexesByCity() {
	t := suite.T()
	type args struct {
		ctx         context.Context
		city        string
		location_id int
	}

	tests := []struct {
		name    string
		args    args
		wantErr error
		m       []NewMultiplex
		prepare func(args, *mocks.Storer)
	}{
		{
			name: "GetAllMultiplexesByCity success",
			args: args{
				ctx:         context.TODO(),
				city:        "Mumbai",
				location_id: 1,
			},
			wantErr: nil,
			m:       []NewMultiplex{{}},
			prepare: func(a args, s *mocks.Storer) {
				s.On("GetLocationIdByCity", a.ctx, a.city).Return(db.Location{Location_id: 1}, nil).Once()
				s.On("GetAllMultiplexesByLocationID", a.ctx, a.location_id).Return([]db.Multiplexe{{}}, nil).Once()

			},
		},
		{
			name: "GetAllMultiplexesByCity failure (Location err)",
			args: args{
				ctx:         context.TODO(),
				city:        "Banglore",
				location_id: 0,
			},
			wantErr: errors.New("failed to get multiplexes"),
			m:       []NewMultiplex{{}},
			prepare: func(a args, s *mocks.Storer) {
				s.On("GetLocationIdByCity", a.ctx, a.city).Return(db.Location{}, errors.New("Location doesn't exist")).Once()

			},
		},
		{

			name: "GetAllMultiplexesByCity failure (Multiplex err)",
			args: args{
				ctx:         context.TODO(),
				city:        "Banglore",
				location_id: 2,
			},
			wantErr: errors.New("failed to get multiplexes"),
			m:       []NewMultiplex{{}},
			prepare: func(a args, s *mocks.Storer) {
				s.On("GetLocationIdByCity", a.ctx, a.city).Return(db.Location{Location_id: 2}, nil).Once()
				s.On("GetAllMultiplexesByLocationID", a.ctx, a.location_id).Return([]db.Multiplexe{{}}, errors.New("No multiplexes found.")).Once()

			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.prepare(tt.args, suite.storer)
			multiplexes, err := suite.service.GetAllMultiplexesByCity(tt.args.ctx, tt.args.city)
			if tt.wantErr != nil {
				suite.ErrorContains(err, "failed to get multiplexes")
			} else {
				suite.ErrorIs(err, nil)
			}

			suite.IsType(tt.m, multiplexes)
		})

	}
}

func (suite *BookingServiceTestSuite) TestGetAllShowsByMovieAndDate() {

	t := suite.T()
	cdate, _ := time.Parse(DateOnly, "2023-02-09")
	type args struct {
		ctx   context.Context
		title string
		city  string
		date  time.Time
	}

	tests := []struct {
		name    string
		args    args
		wantErr error
		date    string
		prepare func(args, *mocks.Storer)
	}{
		{
			name: "success",
			args: args{
				ctx:   context.TODO(),
				title: "Intern",
				city:  "Mumbai",
				date:  cdate,
			},
			wantErr: nil,
			date:    "2023-02-09",
			prepare: func(a args, s *mocks.Storer) {
				s.On("GetAllShowsByMovieAndDate", a.ctx, a.title, a.city, a.date).Return([]db.MultiplexShow{{}}, nil).Once()
			},
		},
		{
			name: "Failure invalid date",
			args: args{
				ctx:   context.TODO(),
				title: "Intern",
				city:  "Mumbai",
				date:  cdate,
			},
			wantErr: errors.New("invalid date format"),
			date:    "23-02-09",
			prepare: func(a args, s *mocks.Storer) {
			},
		},
		{
			name: "Failure Empty rows",
			args: args{
				ctx:   context.TODO(),
				title: "Terminator",
				city:  "Mumbai",
				date:  cdate,
			},
			wantErr: errors.New("No shows found."),
			date:    "2023-02-09",
			prepare: func(a args, s *mocks.Storer) {
				s.On("GetAllShowsByMovieAndDate", a.ctx, a.title, a.city, a.date).Return([]db.MultiplexShow{{}}, errors.New("No shows found.")).Once()

			},
		},
	}

	for _, tt := range tests {

		suite.Run(tt.name, func() {
			tt.prepare(tt.args, suite.storer)
			_, err := suite.service.GetAllShowsByMovieAndDate(tt.args.ctx, tt.date, tt.args.title, tt.args.city)
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				require.ErrorIs(t, err, tt.wantErr)
			}

		})
	}

}

func (suite *BookingServiceTestSuite) TestGetAllSeatsByShowId() {

	t := suite.T()
	type args struct {
		ctx     context.Context
		show_id int
	}

	tests := []struct {
		name    string
		args    args
		wantErr error
		prepare func(args, *mocks.Storer)
	}{
		{
			name: "Success",
			args: args{
				ctx:     context.TODO(),
				show_id: 1,
			},
			wantErr: nil,
			prepare: func(a args, s *mocks.Storer) {
				s.On("GetSeatsByShowID", a.ctx, a.show_id).Return([]db.Seats{{}}, nil).Once()
			},
		},
		{
			name: "Failure",
			args: args{
				ctx:     context.TODO(),
				show_id: 2,
			},
			wantErr: errors.New("No seats found."),
			prepare: func(a args, s *mocks.Storer) {
				s.On("GetSeatsByShowID", a.ctx, a.show_id).Return([]db.Seats{}, errors.New("No seats found.")).Once()
			},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			tt.prepare(tt.args, suite.storer)
			_, err := suite.service.GetAllSeatsByShowID(tt.args.ctx, tt.args.show_id)
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				require.ErrorIs(t, err, tt.wantErr)
			}
		})
	}
}

func (suite *BookingServiceTestSuite) TestAddBookingsBySeatId() {

	t := suite.T()

	type args struct {
		ctx      context.Context
		seats    []int
		email    string
		show_id  int
		seat_num []int
	}

	tests := []struct {
		name    string
		args    args
		wantErr error
		invoice Invoice
		prepare func(args, *mocks.Storer)
	}{
		{
			name: "success",
			args: args{
				ctx:      context.TODO(),
				seats:    []int{},
				email:    "test@test.com",
				show_id:  0,
				seat_num: []int{0},
			},
			wantErr: nil,
			invoice: Invoice{},
			prepare: func(a args, s *mocks.Storer) {
				s.On("CheckAvailability", a.ctx, a.seats).Return(true, nil).Once()
				s.On("GetSeatsByID", a.ctx, a.seats).Return([]db.Seats{{}}, nil).Once()
				s.On("AddBookingsBySeatId", a.ctx, a.seats, a.email, a.show_id, a.seat_num).Return(nil).Once()
				s.On("GetInvoiceDetails", a.ctx, a.show_id).Return(db.Invoice{}, nil).Once()
			},
		},
		{
			name: "failure (Seats not available)",
			args: args{
				ctx:      context.TODO(),
				seats:    []int{180, 181},
				email:    "test@test.com",
				show_id:  0,
				seat_num: []int{},
			},
			wantErr: errors.New("Seats not available"),
			invoice: Invoice{},
			prepare: func(a args, s *mocks.Storer) {
				s.On("CheckAvailability", a.ctx, a.seats).Return(false, sql.ErrNoRows).Once()
			},
		},
		{
			name: "failure (GetSeats)",
			args: args{
				ctx:      context.TODO(),
				seats:    []int{180, 181},
				email:    "test@test.com",
				show_id:  0,
				seat_num: []int{},
			},
			wantErr: errors.New("Err fetching seats"),
			invoice: Invoice{},
			prepare: func(a args, s *mocks.Storer) {
				s.On("CheckAvailability", a.ctx, a.seats).Return(true, nil).Once()
				s.On("GetSeatsByID", a.ctx, a.seats).Return([]db.Seats{{}}, errors.New("Err fetching seats")).Once()
			},
		},
		{
			name: "failure AddBooking",
			args: args{
				ctx:      context.TODO(),
				seats:    []int{},
				email:    "test@test.com",
				show_id:  0,
				seat_num: []int{0},
			},
			wantErr: errors.New("err: cannot generate invoice"),
			invoice: Invoice{},
			prepare: func(a args, s *mocks.Storer) {
				s.On("CheckAvailability", a.ctx, a.seats).Return(true, nil).Once()
				s.On("GetSeatsByID", a.ctx, a.seats).Return([]db.Seats{{}}, nil).Once()
				s.On("AddBookingsBySeatId", a.ctx, a.seats, a.email, a.show_id, a.seat_num).Return(nil).Once()
				s.On("GetInvoiceDetails", a.ctx, a.show_id).Return(db.Invoice{}, errors.New("Err fetching invoice details")).Once()
			},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			tt.prepare(tt.args, suite.storer)
			invoice, err := suite.service.AddBookingsBySeatId(tt.args.ctx, tt.args.seats, tt.args.email)
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				require.ErrorIs(t, err, tt.wantErr)
			}
			suite.IsType(tt.invoice, invoice)
		})
	}

}

func (suite *BookingServiceTestSuite) TestGetUpcomingMovies() {

	t := suite.T()

	type args struct {
		ctx  context.Context
		date string
	}

	tests := []struct {
		name    string
		args    args
		wantErr error
		m       []NewMovie
		prepare func(args, *mocks.Storer)
	}{
		{
			name: "Success",
			args: args{
				ctx:  context.TODO(),
				date: "2023-02-12",
			},
			wantErr: nil,
			m:       []NewMovie{{}},
			prepare: func(a args, s *mocks.Storer) {
				s.On("GetUpcomingMovies", a.ctx, a.date).Return([]db.Movie{{}}, nil).Once()
			},
		},
		{
			name: "Failure",
			args: args{
				ctx:  context.TODO(),
				date: "2023-02-12",
			},
			wantErr: errors.New("No movies available"),
			m:       []NewMovie{{}},
			prepare: func(a args, s *mocks.Storer) {
				s.On("GetUpcomingMovies", a.ctx, a.date).Return([]db.Movie{{}}, errors.New("No movies available")).Once()
			},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			tt.prepare(tt.args, suite.storer)
			movies, err := suite.service.GetUpcomingMovies(tt.args.ctx, tt.args.date)
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				require.ErrorIs(t, err, tt.wantErr)
			}
			suite.IsType(tt.m, movies)
		})
	}
}

func (suite *BookingServiceTestSuite) TestGetMovieByTitle() {

	t := suite.T()
	type args struct {
		ctx   context.Context
		title string
	}

	tests := []struct {
		name    string
		args    args
		wantErr error
		m       NewMovie
		prepare func(a args, s *mocks.Storer)
	}{
		{
			name: "Success",
			args: args{
				ctx:   context.TODO(),
				title: "Terminator",
			},
			wantErr: nil,
			m:       NewMovie{},
			prepare: func(a args, s *mocks.Storer) {
				s.On("GetMovieByTitle", a.ctx, a.title).Return(db.Movie{}, nil).Once()
			},
		},
		{
			name: "Failure",
			args: args{
				ctx:   context.TODO(),
				title: "Terminator",
			},
			wantErr: errors.New("movie not found"),
			m:       NewMovie{},
			prepare: func(a args, s *mocks.Storer) {
				s.On("GetMovieByTitle", a.ctx, a.title).Return(db.Movie{}, errors.New("movie doesn't exist")).Once()
			},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			tt.prepare(tt.args, suite.storer)
			movie, err := suite.service.GetMovieByTitle(tt.args.ctx, tt.args.title)
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				require.ErrorIs(t, err, tt.wantErr)
			}
			suite.IsType(tt.m, movie)
		})
	}

}

func (suite *BookingServiceTestSuite) TestCancelBooking() {

	t := suite.T()
	type args struct {
		ctx context.Context
		id  int
	}

	tests := []struct {
		name    string
		args    args
		wantErr error
		prepare func(a args, s *mocks.Storer)
	}{
		{
			name: "Success",
			args: args{
				ctx: context.TODO(),
				id:  1,
			},
			wantErr: nil,
			prepare: func(a args, s *mocks.Storer) {
				s.On("DeleteByBookingByID", a.ctx, a.id).Return(nil).Once()
			},
		},
		{
			name: "Failure",
			args: args{
				ctx: context.TODO(),
				id:  2,
			},
			wantErr: errors.New("err: cannot cancel booking"),
			prepare: func(a args, s *mocks.Storer) {
				s.On("DeleteByBookingByID", a.ctx, a.id).Return(sql.ErrNoRows).Once()
			},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			tt.prepare(tt.args, suite.storer)
			err := suite.service.CancelBooking(tt.args.ctx, tt.args.id)
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				require.ErrorIs(t, err, tt.wantErr)
			}
		})
	}

}

func (suite *BookingServiceTestSuite) TestAddMovie() {

	t := suite.T()
	type args struct {
		ctx     context.Context
		dbMovie db.Movie
	}

	tests := []struct {
		name     string
		args     args
		m        NewMovie
		wantErr  error
		movie_id uint
		prepare  func(a args, s *mocks.Storer)
	}{
		{
			name: "Success",
			args: args{
				ctx:     context.TODO(),
				dbMovie: db.Movie{},
			},
			m:        NewMovie{},
			wantErr:  nil,
			movie_id: 1,
			prepare: func(a args, s *mocks.Storer) {
				s.On("AddMovie", a.ctx, a.dbMovie).Return(uint(1), nil).Once()
			},
		},
		{
			name: "Failure",
			args: args{
				ctx:     context.TODO(),
				dbMovie: db.Movie{},
			},
			m:        NewMovie{},
			wantErr:  errors.New("failed to add movie"),
			movie_id: 0,
			prepare: func(a args, s *mocks.Storer) {
				s.On("AddMovie", a.ctx, a.dbMovie).Return(uint(0), errors.New("failure")).Once()
			},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			tt.prepare(tt.args, suite.storer)
			movie, err := suite.service.AddMovie(tt.args.ctx, tt.m)
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				require.ErrorIs(t, err, tt.wantErr)
			}
			assert.Equal(t, tt.movie_id, movie)
		})
	}
}

func (suite *BookingServiceTestSuite) TestAddMultiplex() {

	t := suite.T()
	type args struct {
		ctx         context.Context
		dbMultiplex db.Multiplexe
		dbLocation  db.Location
		city        string
	}

	tests := []struct {
		name         string
		args         args
		wantErr      error
		m            NewMultiplex
		multiplex_id uint
		prepare      func(a args, s *mocks.Storer)
	}{
		{
			name: "Success",
			args: args{
				ctx:         context.TODO(),
				dbMultiplex: db.Multiplexe{},
				dbLocation: db.Location{
					City:    "Mumbai",
					State:   "Maharashtra",
					Pincode: 400078,
				},
				city: "Mumbai",
			},
			wantErr: nil,
			m: NewMultiplex{
				City:    "Mumbai",
				State:   "Maharashtra",
				Pincode: 400078,
			},
			multiplex_id: 1,
			prepare: func(a args, s *mocks.Storer) {
				s.On("GetLocationIdByCity", a.ctx, a.city).Return(db.Location{}, nil).Once()
				// s.On("AddLocation", a.ctx, a.dbLocation).Return(1, nil).Once()
				s.On("AddMultiplex", a.ctx, a.dbMultiplex).Return(uint(1), nil).Once()

			},
		},
		{
			name: "Success AddLocation",
			args: args{
				ctx: context.TODO(),
				dbMultiplex: db.Multiplexe{
					Location_id: 1,
				},
				dbLocation: db.Location{
					City:    "Mumbai",
					State:   "Maharashtra",
					Pincode: 400078},
				city: "Mumbai",
			},
			wantErr: nil,
			m: NewMultiplex{
				City:    "Mumbai",
				State:   "Maharashtra",
				Pincode: 400078,
			},
			multiplex_id: 1,
			prepare: func(a args, s *mocks.Storer) {
				s.On("GetLocationIdByCity", a.ctx, a.city).Return(db.Location{}, errors.New("location doesn't exist")).Once()
				s.On("AddLocation", a.ctx, a.dbLocation).Return(1, nil).Once()
				s.On("AddMultiplex", a.ctx, a.dbMultiplex).Return(uint(1), nil).Once()

			},
		},
		{
			name: "Failure AddLocation",
			args: args{
				ctx: context.TODO(),
				dbMultiplex: db.Multiplexe{
					Location_id: 1,
				},
				dbLocation: db.Location{
					City:    "Mumbai",
					State:   "Maharashtra",
					Pincode: 400078},
				city: "Mumbai",
			},
			wantErr: errors.New("cannot add multiplex"),
			m: NewMultiplex{
				City:    "Mumbai",
				State:   "Maharashtra",
				Pincode: 400078,
			},
			multiplex_id: 0,
			prepare: func(a args, s *mocks.Storer) {
				s.On("GetLocationIdByCity", a.ctx, a.city).Return(db.Location{}, errors.New("location doesn't exist")).Once()
				s.On("AddLocation", a.ctx, a.dbLocation).Return(0, errors.New("failed")).Once()
				// s.On("AddMultiplex", a.ctx, a.dbMultiplex).Return(uint(1), nil).Once()

			},
		},
		{
			name: "Failure AddMultiplex",
			args: args{
				ctx: context.TODO(),
				dbMultiplex: db.Multiplexe{
					Location_id: 1,
				},
				dbLocation: db.Location{
					City:    "Mumbai",
					State:   "Maharashtra",
					Pincode: 400078},
				city: "Mumbai",
			},
			wantErr: errors.New("cannot add multiplex"),
			m: NewMultiplex{
				City:    "Mumbai",
				State:   "Maharashtra",
				Pincode: 400078,
			},
			multiplex_id: 0,
			prepare: func(a args, s *mocks.Storer) {
				s.On("GetLocationIdByCity", a.ctx, a.city).Return(db.Location{}, errors.New("location doesn't exist")).Once()
				s.On("AddLocation", a.ctx, a.dbLocation).Return(1, nil).Once()
				s.On("AddMultiplex", a.ctx, a.dbMultiplex).Return(uint(0), errors.New("failed")).Once()

			},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			tt.prepare(tt.args, suite.storer)
			multiplex, err := suite.service.AddMultiplex(tt.args.ctx, tt.m)
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				require.ErrorIs(t, err, tt.wantErr)
			}
			assert.Equal(t, tt.multiplex_id, multiplex)
		})
	}

}

func (suite *BookingServiceTestSuite) TestAddShow() {

	t := suite.T()

	type args struct {
		ctx    context.Context
		s      NewShow
		dbShow db.Show
	}
	rDate, _ := time.Parse(DateOnly, "2023-02-09")
	tests := []struct {
		name    string
		args    args
		wantErr error
		show_id uint
		prepare func(a args, s *mocks.Storer)
	}{
		{
			name: "Success",
			args: args{
				ctx: context.TODO(),
				s: NewShow{
					Date:    "2023-02-09",
					Show_id: 0,
				},
				dbShow: db.Show{
					Show_date: rDate,
					Show_id:   0,
				},
			},
			wantErr: nil,
			show_id: 0,
			prepare: func(a args, s *mocks.Storer) {
				s.On("GetMultiplexeByID", a.ctx, a.s.Multiplex_id).Return(uint(1), nil).Once()
				s.On("GetScreenByNumberAndMultiplexID", a.ctx, a.s.Screen, a.s.Multiplex_id).Return(db.Screen{}, nil).Once()
				s.On("GetMovieByTitle", a.ctx, a.s.Movie).Return(db.Movie{}, nil).Once()
				s.On("AddShow", a.ctx, a.dbShow).Return(uint(0), nil).Once()
				s.On("AddSeats", a.ctx, 0, a.s.Show_id).Return(nil).Once()
			},
		},
		{
			name: "Failure to add seats",
			args: args{
				ctx: context.TODO(),
				s: NewShow{
					Date:    "2023-02-09",
					Show_id: 0,
				},
				dbShow: db.Show{
					Show_date: rDate,
					Show_id:   0,
				},
			},
			wantErr: errors.New("err : adding seats"),
			show_id: 0,
			prepare: func(a args, s *mocks.Storer) {
				s.On("GetMultiplexeByID", a.ctx, a.s.Multiplex_id).Return(uint(1), nil).Once()
				s.On("GetScreenByNumberAndMultiplexID", a.ctx, a.s.Screen, a.s.Multiplex_id).Return(db.Screen{}, nil).Once()
				s.On("GetMovieByTitle", a.ctx, a.s.Movie).Return(db.Movie{}, nil).Once()
				s.On("AddShow", a.ctx, a.dbShow).Return(uint(0), nil).Once()
				s.On("AddSeats", a.ctx, 0, a.s.Show_id).Return(errors.New("failed")).Once()
			},
		},
		{
			name: "Failure to add show",
			args: args{
				ctx: context.TODO(),
				s: NewShow{
					Date:    "2023-02-09",
					Show_id: 0,
				},
				dbShow: db.Show{
					Show_date: rDate,
					Show_id:   0,
				},
			},
			wantErr: errors.New("err : overlapping show times"),
			show_id: 0,
			prepare: func(a args, s *mocks.Storer) {
				s.On("GetMultiplexeByID", a.ctx, a.s.Multiplex_id).Return(uint(1), nil).Once()
				s.On("GetScreenByNumberAndMultiplexID", a.ctx, a.s.Screen, a.s.Multiplex_id).Return(db.Screen{}, nil).Once()
				s.On("GetMovieByTitle", a.ctx, a.s.Movie).Return(db.Movie{}, nil).Once()
				s.On("AddShow", a.ctx, a.dbShow).Return(uint(0), sql.ErrNoRows).Once()
				// s.On("AddSeats", a.ctx, 0, a.s.Show_id).Return(errors.New("failed")).Once()
			},
		},
		{
			name: "Failure (Movie doesn't exist)",
			args: args{
				ctx: context.TODO(),
				s: NewShow{
					Date:    "2023-02-09",
					Show_id: 0,
				},
				dbShow: db.Show{
					Show_date: rDate,
					Show_id:   0,
				},
			},
			wantErr: errors.New("err: Movie doesn't exist"),
			show_id: 0,
			prepare: func(a args, s *mocks.Storer) {
				s.On("GetMultiplexeByID", a.ctx, a.s.Multiplex_id).Return(uint(1), nil).Once()
				s.On("GetScreenByNumberAndMultiplexID", a.ctx, a.s.Screen, a.s.Multiplex_id).Return(db.Screen{}, nil).Once()
				s.On("GetMovieByTitle", a.ctx, a.s.Movie).Return(db.Movie{}, errors.New("movie doesn't exist")).Once()
				// s.On("AddShow", a.ctx, a.dbShow).Return(uint(0), sql.ErrNoRows).Once()
				// s.On("AddSeats", a.ctx, 0, a.s.Show_id).Return(errors.New("failed")).Once()
			},
		},
		{
			name: "Failure (Movie doesn't exist)",
			args: args{
				ctx: context.TODO(),
				s: NewShow{
					Date:    "2023-02-09",
					Show_id: 0,
				},
				dbShow: db.Show{
					Show_date: rDate,
					Show_id:   0,
				},
			},
			wantErr: errors.New("err: invalid screen number"),
			show_id: 0,
			prepare: func(a args, s *mocks.Storer) {
				s.On("GetMultiplexeByID", a.ctx, a.s.Multiplex_id).Return(uint(1), nil).Once()
				s.On("GetScreenByNumberAndMultiplexID", a.ctx, a.s.Screen, a.s.Multiplex_id).Return(db.Screen{}, errors.New("failed")).Once()
				// s.On("GetMovieByTitle", a.ctx, a.s.Movie).Return(db.Movie{}, errors.New("movie doesn't exist")).Once()
				// s.On("AddShow", a.ctx, a.dbShow).Return(uint(0), sql.ErrNoRows).Once()
				// s.On("AddSeats", a.ctx, 0, a.s.Show_id).Return(errors.New("failed")).Once()
			},
		},
		{
			name: "Failure (Movie doesn't exist)",
			args: args{
				ctx: context.TODO(),
				s: NewShow{
					Date:    "2023-02-09",
					Show_id: 0,
				},
				dbShow: db.Show{
					Show_date: rDate,
					Show_id:   0,
				},
			},
			wantErr: errors.New("err: invalid Multiplex id"),
			show_id: 0,
			prepare: func(a args, s *mocks.Storer) {
				s.On("GetMultiplexeByID", a.ctx, a.s.Multiplex_id).Return(uint(1), errors.New("failed")).Once()
				// s.On("GetScreenByNumberAndMultiplexID", a.ctx, a.s.Screen, a.s.Multiplex_id).Return(db.Screen{}, errors.New("failed")).Once()
				// s.On("GetMovieByTitle", a.ctx, a.s.Movie).Return(db.Movie{}, errors.New("movie doesn't exist")).Once()
				// s.On("AddShow", a.ctx, a.dbShow).Return(uint(0), sql.ErrNoRows).Once()
				// s.On("AddSeats", a.ctx, 0, a.s.Show_id).Return(errors.New("failed")).Once()
			},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			tt.prepare(tt.args, suite.storer)
			show, err := suite.service.AddShow(tt.args.ctx, tt.args.s)
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				require.ErrorIs(t, err, tt.wantErr)
			}
			assert.Equal(t, tt.show_id, show)
		})
	}
}

func (suite *BookingServiceTestSuite) TestAddScreen() {

	t := suite.T()

	type args struct {
		ctx      context.Context
		s        *mocks.Storer
		newS     NewScreen
		dbScreen db.Screen
	}

	tests := []struct {
		name      string
		args      args
		wantErr   error
		screen_id uint
		prepare   func(a args, s *mocks.Storer)
	}{
		{
			name: "Success",
			args: args{
				ctx: context.TODO(),
				s:   suite.storer,
				newS: NewScreen{
					Multiplex_id: 1,
				},
				dbScreen: db.Screen{
					Multiplex_id: 1,
				},
			},
			wantErr:   nil,
			screen_id: 1,
			prepare: func(a args, s *mocks.Storer) {
				s.On("GetMultiplexeByID", a.ctx, a.newS.Multiplex_id).Return(uint(1), nil).Once()
				s.On("AddScreen", a.ctx, a.dbScreen).Return(uint(1), nil).Once()
			},
		},
		{
			name: "Failure - multiplexid",
			args: args{
				ctx: context.TODO(),
				s:   suite.storer,
				newS: NewScreen{
					Multiplex_id: 2,
				},
				dbScreen: db.Screen{
					Multiplex_id: 2,
				},
			},
			wantErr:   errors.New("err: invalid Multiplex id"),
			screen_id: 0,
			prepare: func(a args, s *mocks.Storer) {
				s.On("GetMultiplexeByID", a.ctx, a.newS.Multiplex_id).Return(uint(0), errors.New("failed")).Once()
			},
		},
		{
			name: "Failure - Add screen",
			args: args{
				ctx: context.TODO(),
				s:   suite.storer,
				newS: NewScreen{
					Multiplex_id: 2,
				},
				dbScreen: db.Screen{
					Multiplex_id: 2,
				},
			},
			wantErr:   errors.New("failed to add scren"),
			screen_id: 0,
			prepare: func(a args, s *mocks.Storer) {
				s.On("GetMultiplexeByID", a.ctx, a.newS.Multiplex_id).Return(uint(2), nil).Once()
				s.On("AddScreen", a.ctx, a.dbScreen).Return(uint(0), sql.ErrNoRows).Once()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.prepare(tt.args, suite.storer)
			screen_id, err := suite.service.AddScreen(tt.args.ctx, tt.args.newS)
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				require.ErrorIs(t, err, tt.wantErr)
			}
			assert.Equal(t, tt.screen_id, screen_id)
		})
	}

}

func (suite *BookingServiceTestSuite) TestGetAllShowsByDateAndMultiplexId() {

	t := suite.T()
	cDate, _ := time.Parse("2006-01-02", "2023-02-09")
	type args struct {
		ctx          context.Context
		s            *mocks.Storer
		cDate        time.Time
		multiplex_id int
	}

	tests := []struct {
		name    string
		args    args
		wantErr error
		date    string
		shows   map[string][]MultiplexShow
		prepare func(args, *mocks.Storer)
	}{
		{
			name: "Success",
			args: args{
				ctx:          context.TODO(),
				s:            suite.storer,
				cDate:        cDate,
				multiplex_id: 1,
			},
			wantErr: nil,
			date:    "2023-02-09",
			shows:   make(map[string][]MultiplexShow),
			prepare: func(a args, s *mocks.Storer) {
				s.On("GetAllShowsByDateAndMultiplexId", a.ctx, a.cDate, a.multiplex_id).Return([]db.MultiplexShow{{}}, nil).Once()
			},
		},
		{
			name: "Failure: Get all shows by date and multiplex id",

			args: args{
				ctx:          context.TODO(),
				s:            suite.storer,
				cDate:        cDate,
				multiplex_id: 2,
			},
			wantErr: errors.New("No shows found."),
			date:    "2023-02-09",
			shows:   make(map[string][]MultiplexShow),
			prepare: func(a args, s *mocks.Storer) {
				s.On("GetAllShowsByDateAndMultiplexId", a.ctx, a.cDate, a.multiplex_id).Return([]db.MultiplexShow{{}}, errors.New("No shows found.")).Once()
			},
		},
		{
			name: "Failure: invalid time format",

			args: args{
				ctx:          context.TODO(),
				s:            suite.storer,
				cDate:        cDate,
				multiplex_id: 2,
			},
			wantErr: errors.New("invalid date format"),
			date:    "23-02-09",
			shows:   make(map[string][]MultiplexShow),
			prepare: func(a args, s *mocks.Storer) {
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.prepare(tt.args, suite.storer)
			shows, err := suite.service.GetAllShowsByDateAndMultiplexId(tt.args.ctx, tt.date, tt.args.multiplex_id)
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				require.ErrorIs(t, err, tt.wantErr)
			}
			suite.IsType(tt.shows, shows)
		})
	}
}
