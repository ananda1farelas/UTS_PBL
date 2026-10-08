package repository

import (
	"database/sql"
	"errors"

	"github.com/ananda1farelas/latihan-fiber/UTS/app/model"
)

var ErrNotFound = errors.New("record not found")

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) FindByEmail(email string) (*model.User, error) {
	var u model.User

	err := r.db.QueryRow(
		`SELECT id, email, password, role, created_at, updated_at
		 FROM users WHERE email = $1`,
		email,
	).Scan(&u.ID, &u.Email, &u.Password, &u.Role, &u.CreatedAt, &u.UpdatedAt)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return &u, nil
}

func (r *UserRepository) FindByID(id int) (*model.User, error) {
	var u model.User

	err := r.db.QueryRow(
		`SELECT id, email, password, role, created_at, updated_at
		 FROM users WHERE id = $1`,
		id,
	).Scan(&u.ID, &u.Email, &u.Password, &u.Role, &u.CreatedAt, &u.UpdatedAt)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return &u, nil
}

// Create inserts a new user row and returns its generated id.
func (r *UserRepository) Create(tx *sql.Tx, email, hashedPassword, role string) (int, error) {
	var id int
	err := tx.QueryRow(
		`INSERT INTO users (email, password, role) VALUES ($1, $2, $3) RETURNING id`,
		email, hashedPassword, role,
	).Scan(&id)
	return id, err
}

func (r *UserRepository) BeginTx() (*sql.Tx, error) {
	return r.db.Begin()
}
