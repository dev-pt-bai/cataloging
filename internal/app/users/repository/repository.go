package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

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
INSERT INTO users (id, name, email, password)
	VALUES (?, ?, ?, ?)`

const CreateOTPQuery = `
INSERT INTO user_otps (user_id, user_email, otp, expired_at)
	SELECT ?, ?, ?, ?
	WHERE NOT EXISTS(SELECT 1 FROM user_otps WHERE user_id = ? AND expired_at > (UNIX_TIMESTAMP()))`

const GetOTPQuery = `
SELECT user_id, user_email, otp, created_at, expired_at
	FROM user_otps
	WHERE user_id = ? AND otp = ?`

const VerifyUserQuery = `
UPDATE users SET is_verified = 1, updated_at = ?
	WHERE id = ? AND deleted_at = 0`

const ListUserQuery = `
WITH
	cte1 AS (SELECT JSON_OBJECT('id', id, 'name', name, 'email', email, 'password', password, 'isAdmin', is_admin, 'isVerified', is_verified, 'createdAt', created_at, 'updatedAt', updated_at) AS record FROM users `

const GetUserQuery = `
SELECT id, name, email, password, is_admin, is_verified, created_at, updated_at
	FROM users
	WHERE id = ? AND deleted_at = 0`

const UpdateUserQuery = `
UPDATE users SET name = ?, email = ?, password = ?, is_verified = ?, updated_at = ?
	WHERE id = ? AND deleted_at = 0`

const DeleteUserQuery = `
UPDATE users SET deleted_at = ?
	WHERE id = ?`

func (r *Repository) CreateUser(ctx context.Context, user model.User) *errors.Error {
	_, err := r.db.ExecContext(ctx, CreateUserQuery, user.ID, user.Name, user.Email, user.Password)
	if err != nil {
		if errors.HasMySQLErrCode(err, 1062) {
			return errors.New(errors.UserAlreadyExists).Wrap(err)
		}
		return errors.New(errors.RunQueryFailure).Wrap(err)
	}

	return nil
}

func (r *Repository) CreateOTP(ctx context.Context, otp model.UserOTP) *errors.Error {
	res, err := r.db.ExecContext(ctx, CreateOTPQuery, otp.UserID, otp.UserEmail, otp.OTP, otp.ExpiredAt, otp.UserID)
	if err != nil {
		if errors.HasMySQLErrCode(err, 1062) {
			return errors.New(errors.UserOTPAlreadyExists).Wrap(err)
		}
		return errors.New(errors.RunQueryFailure).Wrap(err)
	}

	row, err := res.RowsAffected()
	if err != nil {
		return errors.New(errors.RowsAffectedFailure).Wrap(err)
	}

	if row < 1 {
		return errors.New(errors.UserOTPAlreadyExists)
	}

	return nil
}

func (r *Repository) GetOTP(ctx context.Context, userID string, code string) (*model.UserOTP, *errors.Error) {
	otp := new(model.UserOTP)
	err := r.db.QueryRowContext(ctx, GetOTPQuery, userID, code).Scan(&otp.UserID, &otp.UserEmail, &otp.OTP, &otp.CreatedAt, &otp.ExpiredAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New(errors.UserOTPNotFound)
		}
		return nil, errors.New(errors.RunQueryFailure).Wrap(err)
	}

	return otp, nil
}

func (r *Repository) VerifyUser(ctx context.Context, ID string) *errors.Error {
	res, err := r.db.ExecContext(ctx, VerifyUserQuery, time.Now().Unix(), ID)
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

func (r *Repository) ListUsers(ctx context.Context, criteria model.ListUsersCriteria) (*model.Users, *errors.Error) {
	query, args, err := r.buildListUsersQuery(criteria)
	if err != nil {
		return nil, errors.New(errors.BuildQueryFailure).Wrap(err)
	}

	users := new(model.Users)
	err = r.db.QueryRowContext(ctx, query, args...).Scan(&users)
	if err != nil && err != sql.ErrNoRows {
		return nil, errors.New(errors.RunQueryFailure).Wrap(err)
	}

	return users, nil
}

type listParam struct {
	q    strings.Builder
	args []any
}

func (r *Repository) buildListUsersQuery(criteria model.ListUsersCriteria) (string, []any, error) {
	param := listParam{
		q:    strings.Builder{},
		args: make([]any, 0, 5),
	}
	param.q.WriteString(ListUserQuery)

	r.filterUser(criteria.FilterUser, &param)
	if err := r.sort(criteria.Sort, &param, model.IsAvailableToSortUser); err != nil {
		return "", nil, err
	}
	if err := r.paginate(criteria.Page, &param); err != nil {
		return "", nil, err
	}

	return param.q.String(), param.args, nil
}

func (r *Repository) filterUser(filter model.FilterUser, param *listParam) {
	whereClauses := make([]string, 0, 5)
	whereClauses = append(whereClauses, "deleted_at = 0 ")

	if len(filter.Name) != 0 {
		whereClauses = append(whereClauses, "name LIKE ? ")
		param.args = append(param.args, fmt.Sprintf("%%%s%%", filter.Name))
	}

	if filter.IsAdmin != nil {
		whereClauses = append(whereClauses, "is_admin = ? ")
		param.args = append(param.args, *filter.IsAdmin)
	}

	if filter.IsVerified != nil {
		whereClauses = append(whereClauses, "is_verified = ? ")
		param.args = append(param.args, *filter.IsVerified)
	}

	param.q.WriteString(fmt.Sprintf("WHERE %s ", strings.Join(whereClauses, "AND ")))
}

func (r *Repository) sort(sortCriteria model.Sort, param *listParam, isAvailable func(string) bool) *errors.Error {
	if len(sortCriteria.FieldName) == 0 {
		param.q.WriteString("ORDER BY created_at DESC), ")
		return nil
	}

	if !isAvailable(sortCriteria.FieldName) {
		return errors.New(errors.UnknownField)
	}
	param.q.WriteString(fmt.Sprintf("ORDER BY %s ", sortCriteria.FieldName))

	if sortCriteria.IsDescending {
		param.q.WriteString("DESC), ")
		return nil
	}
	param.q.WriteString("ASC), ")

	return nil
}

func (r *Repository) paginate(page model.Page, param *listParam) *errors.Error {
	param.q.WriteString("cte2 AS (SELECT record FROM cte1 ")

	if page.ItemPerPage < 1 || page.ItemPerPage > 20 {
		return errors.New(errors.InvalidItemNumberPerPage)
	}

	if page.Number < 1 {
		return errors.New(errors.InvalidItemNumberPerPage)
	}

	param.q.WriteString("LIMIT ? ")
	param.args = append(param.args, page.ItemPerPage)

	param.q.WriteString("OFFSET ?), ")
	param.args = append(param.args, (page.Number-1)*page.ItemPerPage)

	param.q.WriteString(`
	cte3 AS (SELECT JSON_ARRAYAGG(record) AS data FROM cte2),
	cte4 AS (SELECT COUNT(*) AS count FROM cte1)
	SELECT JSON_OBJECT('data', COALESCE(data, CAST('[]' AS JSON)), 'count', count) FROM cte3 JOIN cte4`)

	return nil
}

func (r *Repository) GetUser(ctx context.Context, ID string) (*model.User, *errors.Error) {
	user := new(model.User)
	err := r.db.QueryRowContext(ctx, GetUserQuery, ID).Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.IsAdmin, &user.IsVerified, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New(errors.UserNotFound)
		}
		return nil, errors.New(errors.RunQueryFailure).Wrap(err)
	}

	return user, nil
}

func (r *Repository) UpdateUser(ctx context.Context, user model.User) *errors.Error {
	res, err := r.db.ExecContext(ctx, UpdateUserQuery, user.Name, user.Email, user.Password, user.IsVerified, time.Now().Unix(), user.ID)
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

func (r *Repository) DeleteUser(ctx context.Context, ID string) *errors.Error {
	_, err := r.db.ExecContext(ctx, DeleteUserQuery, time.Now().Unix(), ID)
	if err != nil {
		return errors.New(errors.RunQueryFailure).Wrap(err)
	}

	return nil
}
