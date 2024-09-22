package db

import (
	"database/sql"
	"fmt"

	"github.com/silver-eva/auth_service/auth_service/models"
)

// PostgresInterface represents database actions.
type PostgresInterface interface {
	GetUser(name string, password string, auth bool) (models.User, error)
	CreateUser(name string, password string, email string) (models.User, error)
	IsLoggedIn(userID string) (bool, error)
	SetLoggedIn(userID string, loggedIn bool) error
}

// PostgresDB implements the PostgresInterface.
type PostgresDB struct {
	DB *sql.DB
}

func New(connStr string) (*PostgresDB, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	return &PostgresDB{DB: db}, nil
}

func (p *PostgresDB) GetUser(name, password string, auth bool) (models.User, error) {
	var user models.User
	var sql_password_check string

	if auth {
		sql_password_check = fmt.Sprintf("'%s'", password)
	} else {
		sql_password_check = fmt.Sprintf("crypt('%s', app.user.password)", password)
	}

	query := fmt.Sprintf("select app.user.id, app.user.name, app.user.password, app.user.role, app.loggedin.is_logged_in FROM app.user join app.loggedin on app.loggedin.user_id = app.user.id where app.user.name = '%s' and	app.user.password = %s", name, sql_password_check)
	err := p.DB.QueryRow(query).Scan(&user.Id, &user.Name, &user.Password, &user.Role, &user.IsLoggedIn)
	if err != nil {
		return models.User{}, err
	}
	return user, nil
}

func (p *PostgresDB) CreateUser(name, password, email string) (models.User, error) {
	var user models.User
	user_query := "INSERT INTO app.user (name, password, email) VALUES ($1, $2, $3) RETURNING id, name, password, role"
	err := p.DB.QueryRow(user_query, name, password, email).Scan(&user.Id, &user.Name, &user.Password, &user.Role)
	if err != nil {
		return user, err
	}
	loggedin_query := "insert into app.loggedin (user_id,is_logged_in) values ($1,$2) returning is_logged_in"
	err = p.DB.QueryRow(loggedin_query, user.Id, true).Scan(&user.IsLoggedIn)
	if err != nil {
		return user, err
	}
	return user, nil
}

func (p *PostgresDB) IsLoggedIn(userID string) (bool, error) {
	var loggedIn bool
	query := "select app.loggedin.is_logged_in from app.loggedin where app.loggedin.user_id = $1"
	err := p.DB.QueryRow(query, userID).Scan(&loggedIn)
	if err != nil {
		return false, err
	}
	return loggedIn, nil
}

func (p *PostgresDB) SetLoggedIn(userID string, loggedIn bool) error {
	query := "update app.loggedin set is_logged_in=$1 where app.loggedin.user_id = $2;"
	_, err := p.DB.Exec(query, loggedIn, userID)
	return err
}
