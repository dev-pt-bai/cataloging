package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

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

const ListUserQuery = `
SELECT id, name, email, password, is_admin, CAST (EXTRACT (EPOCH FROM created_at) AS integer), CAST (EXTRACT (EPOCH FROM updated_at) AS integer)
	FROM "cataloging".users `

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

func (r *Repository) ListUsers(ctx context.Context, criteria model.ListUsersCriteria) ([]*model.User, *errors.Error) {
	query, args, err := r.buildListUsersQuery(criteria)
	if err != nil {
		return nil, errors.New(errors.BuildQueryFailure).Wrap(err)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, errors.New(errors.RunQueryFailure).Wrap(err)
	}
	defer rows.Close()

	users := make([]*model.User, 0, 10)
	for rows.Next() {
		user := new(model.User)
		if err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.IsAdmin, &user.CreatedAt, &user.UpdatedAt); err != nil {
			return nil, errors.New(errors.ScanRowsFailure).Wrap(err)
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.New(errors.ScanRowsFailure).Wrap(err)
	}

	return users, nil
}

type listUserParam struct {
	q           strings.Builder
	args        []any
	placeholder int
}

func (r *Repository) buildListUsersQuery(criteria model.ListUsersCriteria) (string, []any, error) {
	param := listUserParam{
		q:           strings.Builder{},
		args:        make([]any, 0, 5),
		placeholder: 1,
	}
	param.q.WriteString(ListUserQuery)

	r.filterUser(criteria.FilterUser, &param)
	if err := r.sortUser(criteria.Sort, &param); err != nil {
		return "", nil, err
	}
	if err := r.paginate(criteria.Page, &param); err != nil {
		return "", nil, err
	}

	return param.q.String(), param.args, nil
}

func (r *Repository) filterUser(filter model.FilterUser, param *listUserParam) {
	whereClauses := make([]string, 0, 5)
	whereClauses = append(whereClauses, "deleted_at IS NULL ")

	if len(filter.Name) != 0 {
		whereClauses = append(whereClauses, fmt.Sprintf("name ILIKE '%%' || $%d || '%%' ", param.placeholder))
		param.args = append(param.args, filter.Name)
		param.placeholder++
	}

	if filter.IsAdmin != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("is_admin IS %v ", *filter.IsAdmin))
	}

	param.q.WriteString(fmt.Sprintf("WHERE %s ", strings.Join(whereClauses, "AND ")))
}

func (r *Repository) sortUser(sortCriteria model.Sort, param *listUserParam) *errors.Error {
	if len(sortCriteria.FieldName) == 0 {
		param.q.WriteString("ORDER BY created_at DESC ")
		return nil
	}

	if !model.IsAvailableToSortUser(sortCriteria.FieldName) {
		return errors.New(errors.UnknownField)
	}
	param.q.WriteString(fmt.Sprintf("ORDER BY %s ", sortCriteria.FieldName))

	if sortCriteria.IsDescending {
		param.q.WriteString("DESC ")
		return nil
	}
	param.q.WriteString("ASC ")

	return nil
}

func (r *Repository) paginate(page model.Page, param *listUserParam) *errors.Error {
	if page.ItemPerPage < 1 || page.ItemPerPage > 20 {
		return errors.New(errors.InvalidItemNumberPerPage)
	}

	if page.Number < 1 {
		return errors.New(errors.InvalidItemNumberPerPage)
	}

	param.q.WriteString(fmt.Sprintf("LIMIT $%d ", param.placeholder))
	param.args = append(param.args, page.ItemPerPage)
	param.placeholder++

	param.q.WriteString(fmt.Sprintf("OFFSET $%d ", param.placeholder))
	param.args = append(param.args, (page.Number-1)*page.ItemPerPage)

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
