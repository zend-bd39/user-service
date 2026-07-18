package repository

import (
	"context"
	"errors"
	"user-service/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)
var (
	ErrInternalServerError = errors.New("internal server error")
)

type userRepository struct {
	pool *pgxpool.Pool
}

var _ domain.UserRepository = (*userRepository)(nil)

func NewUserRepository(pool *pgxpool.Pool) *userRepository {
	return &userRepository{
		pool: pool,
	}
}

func (u *userRepository) Create(ctx context.Context, user domain.User) (domain.User, error) {
	query := "INSERT INTO users (username, email, password_hash) VALUES ($1,$2,$3) RETURNING id, username, email"
	var insertedUser domain.User
	if err := u.pool.QueryRow(ctx, query, user.Username, user.Email, user.PasswordHash).
	Scan(
		&insertedUser.ID,
		&insertedUser.Username,
		&insertedUser.Email,
	); err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok{
			//https://pkg.go.dev/github.com/jackc/pgerrcode#section-readme
			if pgErr.Code == "23505" {
				switch pgErr.ConstraintName {
				case "users_email_key":
					return domain.User{}, domain.ErrEmailTaken
				case "users_username_key":
					return domain.User{}, domain.ErrUsernameTaken
				default:
					return domain.User{}, ErrInternalServerError
				}
			}
		}
		return domain.User{}, ErrInternalServerError
	}
	return insertedUser, nil
}
func (u *userRepository) FindByUsername(ctx context.Context, username string) (domain.User, error) {
	query := "SELECT id, username, email, password_hash, role FROM users WHERE username = $1"
	var user domain.User
	err := u.pool.QueryRow(ctx, query, username).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Role,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, domain.ErrUserNotFound
		}
		return domain.User{}, ErrInternalServerError
	}
	return user, nil
}

func (u *userRepository) FindByID(ctx context.Context, id int)(domain.User, error) {
	query := "SELECT id, username, email, role FROM users WHERE id = $1"
	var user domain.User
	err  := u.pool.QueryRow(ctx, query, id).
		Scan(&user.ID, &user.Username, &user.Email, &user.Role)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, domain.ErrUserNotFound
		}
		return domain.User{}, ErrInternalServerError
	}
	return user, nil
}