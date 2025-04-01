package repository

import (
	"context"
	"database/sql"

	"github.com/dev-pt-bai/cataloging/internal/model"
	"github.com/dev-pt-bai/cataloging/internal/pkg/errors"
)

type Repository struct {
	db *sql.DB
}

func New(db *sql.DB) *Repository {
	return &Repository{db: db}
}

const CreateUserQuery = `
INSERT INTO "cataloging".users (id, name, email, password)
	VALUES ($1, $2, $3, $4)
	ON CONFLICT DO NOTHING`

const GetUserByIDQuery = `
SELECT id, name, email, password, is_admin, CAST (EXTRACT (EPOCH FROM created_at) AS integer), CAST (EXTRACT (EPOCH FROM updated_at) AS integer)
	FROM "cataloging".users
	WHERE id = $1 AND deleted_at IS NULL`

const UpdateUserQuery = `
UPDATE "cataloging".users SET (name, email, password, updated_at) = ($2, $3, $4, NOW())
	WHERE id = $1 AND deleted_at IS NULL`

const DeleteUserQuery = `
UPDATE "cataloging".users SET deleted_at = NOW()
	WHERE id = $1`

func (r *Repository) CreateUser(ctx context.Context, user model.User) *errors.Error {
	res, err := r.db.ExecContext(ctx, CreateUserQuery, user.ID, user.Name, user.Email, user.Password)
	if err != nil {
		return errors.New(errors.RunQueryFailure).Wrap(err)
	}

	row, err := res.RowsAffected()
	if err != nil {
		return errors.New(errors.RowsAffectedFailure).Wrap(err)
	}

	if row < 1 {
		return errors.New(errors.UserAlreadyExists)
	}

	return nil
}

func (r *Repository) GetUserByID(ctx context.Context, ID string) (*model.User, *errors.Error) {
	user := new(model.User)
	err := r.db.QueryRowContext(ctx, GetUserByIDQuery, ID).Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.IsAdmin, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New(errors.UserNotFound)
		}
		return nil, errors.New(errors.RunQueryFailure).Wrap(err)
	}

	return user, nil
}

func (r *Repository) UpdateUser(ctx context.Context, user model.User) *errors.Error {
	res, err := r.db.ExecContext(ctx, UpdateUserQuery, user.ID, user.Name, user.Email, user.Password)
	if err != nil {
		return errors.New(errors.RunQueryFailure).Wrap(err)
	}

	row, err := res.RowsAffected()
	if err != nil {
		return errors.New(errors.RowsAffectedFailure).Wrap(err)
	}

	if row < 1 {
		return errors.New(errors.UserNotFound)
	}

	return nil
}

func (r *Repository) DeleteUserByID(ctx context.Context, ID string) *errors.Error {
	_, err := r.db.ExecContext(ctx, DeleteUserQuery, ID)
	if err != nil {
		return errors.New(errors.RunQueryFailure).Wrap(err)
	}

	return nil
}
