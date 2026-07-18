package repository

import (
	"context"
	"errors"
	"fmt"
	"user-service/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrViolation     = errors.New("email or username already exists in database")
	ErrInternal = errors.New("internal server error")
	ErrNotFound       = errors.New("user not found")
)
var _ domain.UserRepository = (*UserAdaptor)(nil)

type UserAdaptor struct {
	CreateFn func(ctx context.Context, user domain.User) (domain.User, error)
	FindByUsernameFn func(ctx context.Context, username string) (domain.User, error)
	FindByIDFn func(ctx context.Context, id int)(domain.User, error)
}
func (u *UserAdaptor) Create(ctx context.Context, user domain.User)(domain.User, error) {
	if u.CreateFn == nil {
		return domain.User{}, fmt.Errorf("CreateFn not implemented")
	}
	return u.CreateFn(ctx, user)
}
func (u *UserAdaptor) FindByUsername(ctx context.Context, username string) (domain.User, error) {
	if u.FindByUsernameFn == nil {
		return domain.User{}, fmt.Errorf("FindByUsernameFn not implemented")
	}
	return u.FindByUsernameFn(ctx, username)
}
func (u *UserAdaptor) FindByID(ctx context.Context, id int)(domain.User, error) {
	if u.FindByIDFn == nil {
		return domain.User{}, fmt.Errorf("FindByIDFn not implemented")
	}
	return u.FindByIDFn(ctx, id)
}

func UserRepository(db *pgxpool.Pool) *UserAdaptor{
	return &UserAdaptor{
		CreateFn: func(ctx context.Context, user domain.User) (domain.User, error) {
			query := "INSERT INTO users (username, email, password_hash) VALUES ($1,$2,$3) RETURNING id, username, email"
			var insertedUser domain.User
			if err := db.QueryRow(ctx, query, user.Username, user.Email, user.PasswordHash).
			Scan(
				&insertedUser.ID,
				&insertedUser.Username,
				&insertedUser.Email,
			); err != nil {
				if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok{
					//https://pkg.go.dev/github.com/jackc/pgerrcode#section-readme
					if pgErr.Code == "23505" {
						return domain.User{}, ErrViolation
					}
				}
				return domain.User{}, ErrInternal
			}
			return insertedUser, nil
		},
		FindByUsernameFn: func(ctx context.Context, username string) (domain.User, error) {
			query := "SELECT id, username, email, password_hash, role FROM users WHERE username = $1"
			var user domain.User
			err := db.QueryRow(ctx, query, username).Scan(
				&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Role,
			)
			if err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					return domain.User{}, ErrNotFound
				}
				return domain.User{}, ErrInternal
			}
			return user, nil
		},
		FindByIDFn: func(ctx context.Context, id int) (domain.User, error) {
			query := "SELECT id, username, email, role FROM users WHERE id = $1"
			var user domain.User
			err  := db.QueryRow(ctx, query, id).
				Scan(&user.ID, &user.Username, &user.Email, &user.Role)
			if err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					return domain.User{}, ErrNotFound
				}
				return domain.User{}, ErrInternal
			}
			return user, nil
		},
	}
}