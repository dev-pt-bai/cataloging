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

const CreateMaterialTypeQuery = `
INSERT INTO "cataloging".material_types (code, description, val_class)
	VALUES ($1, $2, $3)
	ON CONFLICT DO NOTHING`

const CreateMaterialUoMQuery = `
INSERT INTO "cataloging".material_uoms (code, description)
	VALUES ($1, $2)
	ON CONFLICT DO NOTHING`

const CreateMaterialGroupQuery = `
INSERT INTO "cataloging".material_groups (code, description)
	VALUES ($1, $2)
	ON CONFLICT DO NOTHING`

const ListMaterialTypeQuery = `
SELECT code, description, val_class, CAST (EXTRACT (EPOCH FROM created_at) AS integer), CAST (EXTRACT (EPOCH FROM updated_at) AS integer)
	FROM "cataloging".material_types `

const ListMaterialUoMQuery = `
SELECT code, description, CAST (EXTRACT (EPOCH FROM created_at) AS integer), CAST (EXTRACT (EPOCH FROM updated_at) AS integer)
	FROM "cataloging".material_uoms `

const ListMaterialGroupQuery = `
SELECT code, description, CAST (EXTRACT (EPOCH FROM created_at) AS integer), CAST (EXTRACT (EPOCH FROM updated_at) AS integer)
	FROM "cataloging".material_groups `

const GetMaterialTypeByCodeQuery = `
SELECT code, description, val_class, CAST (EXTRACT (EPOCH FROM created_at) AS integer), CAST (EXTRACT (EPOCH FROM updated_at) AS integer)
	FROM "cataloging".material_types
	WHERE code = $1 AND deleted_at IS NULL`

const GetMaterialUoMByCodeQuery = `
SELECT code, description, CAST (EXTRACT (EPOCH FROM created_at) AS integer), CAST (EXTRACT (EPOCH FROM updated_at) AS integer)
	FROM "cataloging".material_uoms
	WHERE code = $1 AND deleted_at IS NULL`

const GetMaterialGroupByCodeQuery = `
SELECT code, description, CAST (EXTRACT (EPOCH FROM created_at) AS integer), CAST (EXTRACT (EPOCH FROM updated_at) AS integer)
	FROM "cataloging".material_groups
	WHERE code = $1 AND deleted_at IS NULL`

const UpdateMaterialTypeQuery = `
UPDATE "cataloging".material_types SET (description, val_class, updated_at) = ($2, $3, NOW())
	WHERE code = $1 AND deleted_at IS NULL`

const UpdateMaterialUoMQuery = `
UPDATE "cataloging".material_uoms SET (description, updated_at) = ($2, NOW())
	WHERE code = $1 AND deleted_at IS NULL`

const UpdateMaterialGroupQuery = `
UPDATE "cataloging".material_groups SET (description, updated_at) = ($2, NOW())
	WHERE code = $1 AND deleted_at IS NULL`

const DeleteMaterialTypeQuery = `
UPDATE "cataloging".material_types SET deleted_at = NOW()
	WHERE code = $1`

const DeleteMaterialUoMQuery = `
UPDATE "cataloging".material_uoms SET deleted_at = NOW()
	WHERE code = $1`

const DeleteMaterialGroupQuery = `
UPDATE "cataloging".material_groups SET deleted_at = NOW()
	WHERE code = $1`

func (r *Repository) CreateMaterialType(ctx context.Context, mt model.MaterialType) *errors.Error {
	res, err := r.db.ExecContext(ctx, CreateMaterialTypeQuery, mt.Code, mt.Description, mt.ValuationClass)
	if err != nil {
		return errors.New(errors.RunQueryFailure).Wrap(err)
	}

	row, err := res.RowsAffected()
	if err != nil {
		return errors.New(errors.RowsAffectedFailure).Wrap(err)
	}

	if row < 1 {
		return errors.New(errors.MaterialTypeAlreadyExists)
	}

	return nil
}

func (r *Repository) CreateMaterialUoM(ctx context.Context, uom model.MaterialUoM) *errors.Error {
	res, err := r.db.ExecContext(ctx, CreateMaterialUoMQuery, uom.Code, uom.Description)
	if err != nil {
		return errors.New(errors.RunQueryFailure).Wrap(err)
	}

	row, err := res.RowsAffected()
	if err != nil {
		return errors.New(errors.RowsAffectedFailure).Wrap(err)
	}

	if row < 1 {
		return errors.New(errors.MaterialUoMAlreadyExists)
	}

	return nil
}

func (r *Repository) CreateMaterialGroup(ctx context.Context, mg model.MaterialGroup) *errors.Error {
	res, err := r.db.ExecContext(ctx, CreateMaterialGroupQuery, mg.Code, mg.Description)
	if err != nil {
		return errors.New(errors.RunQueryFailure).Wrap(err)
	}

	row, err := res.RowsAffected()
	if err != nil {
		return errors.New(errors.RowsAffectedFailure).Wrap(err)
	}

	if row < 1 {
		return errors.New(errors.MaterialGroupAlreadyExists)
	}

	return nil
}

func (r *Repository) ListMaterialTypes(ctx context.Context, criteria model.ListMaterialTypesCriteria) ([]*model.MaterialType, *errors.Error) {
	query, args, err := r.buildListMaterialTypesQuery(criteria)
	if err != nil {
		return nil, errors.New(errors.BuildQueryFailure).Wrap(err)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, errors.New(errors.RunQueryFailure).Wrap(err)
	}
	defer rows.Close()

	mts := make([]*model.MaterialType, 0, 10)
	for rows.Next() {
		mt := new(model.MaterialType)
		if err := rows.Scan(&mt.Code, &mt.Description, &mt.ValuationClass, &mt.CreatedAt, &mt.UpdatedAt); err != nil {
			return nil, errors.New(errors.ScanRowsFailure).Wrap(err)
		}
		mts = append(mts, mt)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.New(errors.ScanRowsFailure).Wrap(err)
	}

	return mts, nil
}

func (r *Repository) ListMaterialUoMs(ctx context.Context, criteria model.ListMaterialUoMsCriteria) ([]*model.MaterialUoM, *errors.Error) {
	query, args, err := r.buildListMaterialUoMsQuery(criteria)
	if err != nil {
		return nil, errors.New(errors.BuildQueryFailure).Wrap(err)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, errors.New(errors.RunQueryFailure).Wrap(err)
	}
	defer rows.Close()

	uoms := make([]*model.MaterialUoM, 0, 10)
	for rows.Next() {
		uom := new(model.MaterialUoM)
		if err := rows.Scan(&uom.Code, &uom.Description, &uom.CreatedAt, &uom.UpdatedAt); err != nil {
			return nil, errors.New(errors.ScanRowsFailure).Wrap(err)
		}
		uoms = append(uoms, uom)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.New(errors.ScanRowsFailure).Wrap(err)
	}

	return uoms, nil
}

func (r *Repository) ListMaterialGroups(ctx context.Context, criteria model.ListMaterialGroupsCriteria) ([]*model.MaterialGroup, *errors.Error) {
	query, args, err := r.buildListMaterialGroupsQuery(criteria)
	if err != nil {
		return nil, errors.New(errors.BuildQueryFailure).Wrap(err)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, errors.New(errors.RunQueryFailure).Wrap(err)
	}
	defer rows.Close()

	mgs := make([]*model.MaterialGroup, 0, 10)
	for rows.Next() {
		mg := new(model.MaterialGroup)
		if err := rows.Scan(&mg.Code, &mg.Description, &mg.CreatedAt, &mg.UpdatedAt); err != nil {
			return nil, errors.New(errors.ScanRowsFailure).Wrap(err)
		}
		mgs = append(mgs, mg)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.New(errors.ScanRowsFailure).Wrap(err)
	}

	return mgs, nil
}

type listParam struct {
	q           strings.Builder
	args        []any
	placeholder int
}

func (r *Repository) buildListMaterialTypesQuery(criteria model.ListMaterialTypesCriteria) (string, []any, error) {
	param := listParam{
		q:           strings.Builder{},
		args:        make([]any, 0, 5),
		placeholder: 1,
	}
	param.q.WriteString(ListMaterialTypeQuery)

	r.filterMaterialType(criteria.FilterMaterialType, &param)
	if err := r.sort(criteria.Sort, &param, model.IsAvailableToSortMaterialType); err != nil {
		return "", nil, err
	}
	if err := r.paginate(criteria.Page, &param); err != nil {
		return "", nil, err
	}

	return param.q.String(), param.args, nil
}

func (r *Repository) filterMaterialType(filter model.FilterMaterialType, param *listParam) {
	whereClauses := make([]string, 0, 5)
	whereClauses = append(whereClauses, "deleted_at IS NULL ")

	if len(filter.Description) != 0 {
		whereClauses = append(whereClauses, fmt.Sprintf("name ILIKE '%%' || $%d || '%%' ", param.placeholder))
		param.args = append(param.args, filter.Description)
		param.placeholder++
	}

	param.q.WriteString(fmt.Sprintf("WHERE %s ", strings.Join(whereClauses, "AND ")))
}

func (r *Repository) buildListMaterialUoMsQuery(criteria model.ListMaterialUoMsCriteria) (string, []any, error) {
	param := listParam{
		q:           strings.Builder{},
		args:        make([]any, 0, 5),
		placeholder: 1,
	}
	param.q.WriteString(ListMaterialUoMQuery)

	r.filterMaterialUoM(criteria.FilterMaterialUoM, &param)
	if err := r.sort(criteria.Sort, &param, model.IsAvailableToSortMaterialUoM); err != nil {
		return "", nil, err
	}
	if err := r.paginate(criteria.Page, &param); err != nil {
		return "", nil, err
	}

	return param.q.String(), param.args, nil
}

func (r *Repository) filterMaterialUoM(filter model.FilterMaterialUoM, param *listParam) {
	whereClauses := make([]string, 0, 5)
	whereClauses = append(whereClauses, "deleted_at IS NULL ")

	if len(filter.Description) != 0 {
		whereClauses = append(whereClauses, fmt.Sprintf("name ILIKE '%%' || $%d || '%%' ", param.placeholder))
		param.args = append(param.args, filter.Description)
		param.placeholder++
	}

	param.q.WriteString(fmt.Sprintf("WHERE %s ", strings.Join(whereClauses, "AND ")))
}

func (r *Repository) buildListMaterialGroupsQuery(criteria model.ListMaterialGroupsCriteria) (string, []any, error) {
	param := listParam{
		q:           strings.Builder{},
		args:        make([]any, 0, 5),
		placeholder: 1,
	}
	param.q.WriteString(ListMaterialGroupQuery)

	r.filterMaterialGroup(criteria.FilterMaterialGroup, &param)
	if err := r.sort(criteria.Sort, &param, model.IsAvailableToSortMaterialGroup); err != nil {
		return "", nil, err
	}
	if err := r.paginate(criteria.Page, &param); err != nil {
		return "", nil, err
	}

	return param.q.String(), param.args, nil
}

func (r *Repository) filterMaterialGroup(filter model.FilterMaterialGroup, param *listParam) {
	whereClauses := make([]string, 0, 5)
	whereClauses = append(whereClauses, "deleted_at IS NULL ")

	if len(filter.Description) != 0 {
		whereClauses = append(whereClauses, fmt.Sprintf("name ILIKE '%%' || $%d || '%%' ", param.placeholder))
		param.args = append(param.args, filter.Description)
		param.placeholder++
	}

	param.q.WriteString(fmt.Sprintf("WHERE %s ", strings.Join(whereClauses, "AND ")))
}

func (r *Repository) sort(sortCriteria model.Sort, param *listParam, isAvailable func(string) bool) *errors.Error {
	if len(sortCriteria.FieldName) == 0 {
		param.q.WriteString("ORDER BY created_at DESC ")
		return nil
	}

	if !isAvailable(sortCriteria.FieldName) {
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

func (r *Repository) paginate(page model.Page, param *listParam) *errors.Error {
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

func (r *Repository) GetMaterialTypeByCode(ctx context.Context, code string) (*model.MaterialType, *errors.Error) {
	mt := new(model.MaterialType)
	err := r.db.QueryRowContext(ctx, GetMaterialTypeByCodeQuery, code).Scan(&mt.Code, &mt.Description, &mt.ValuationClass, &mt.CreatedAt, &mt.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New(errors.MaterialTypeNotFound)
		}
		return nil, errors.New(errors.RunQueryFailure).Wrap(err)
	}

	return mt, nil
}

func (r *Repository) GetMaterialUoMByCode(ctx context.Context, code string) (*model.MaterialUoM, *errors.Error) {
	uom := new(model.MaterialUoM)
	err := r.db.QueryRowContext(ctx, GetMaterialUoMByCodeQuery, code).Scan(&uom.Code, &uom.Description, &uom.CreatedAt, &uom.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New(errors.MaterialUoMNotFound)
		}
		return nil, errors.New(errors.RunQueryFailure).Wrap(err)
	}

	return uom, nil
}

func (r *Repository) GetMaterialGroupByCode(ctx context.Context, code string) (*model.MaterialGroup, *errors.Error) {
	mg := new(model.MaterialGroup)
	err := r.db.QueryRowContext(ctx, GetMaterialGroupByCodeQuery, code).Scan(&mg.Code, &mg.Description, &mg.CreatedAt, &mg.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New(errors.MaterialGroupNotFound)
		}
		return nil, errors.New(errors.RunQueryFailure).Wrap(err)
	}

	return mg, nil
}

func (r *Repository) UpdateMaterialType(ctx context.Context, mt model.MaterialType) *errors.Error {
	res, err := r.db.ExecContext(ctx, UpdateMaterialTypeQuery, mt.Code, mt.Description, mt.ValuationClass)
	if err != nil {
		return errors.New(errors.RunQueryFailure).Wrap(err)
	}

	row, err := res.RowsAffected()
	if err != nil {
		return errors.New(errors.RowsAffectedFailure).Wrap(err)
	}

	if row < 1 {
		return errors.New(errors.MaterialTypeNotFound)
	}

	return nil
}

func (r *Repository) UpdateMaterialUoM(ctx context.Context, uom model.MaterialUoM) *errors.Error {
	res, err := r.db.ExecContext(ctx, UpdateMaterialUoMQuery, uom.Code, uom.Description)
	if err != nil {
		return errors.New(errors.RunQueryFailure).Wrap(err)
	}

	row, err := res.RowsAffected()
	if err != nil {
		return errors.New(errors.RowsAffectedFailure).Wrap(err)
	}

	if row < 1 {
		return errors.New(errors.MaterialUoMNotFound)
	}

	return nil
}

func (r *Repository) UpdateMaterialGroup(ctx context.Context, mg model.MaterialGroup) *errors.Error {
	res, err := r.db.ExecContext(ctx, UpdateMaterialGroupQuery, mg.Code, mg.Description)
	if err != nil {
		return errors.New(errors.RunQueryFailure).Wrap(err)
	}

	row, err := res.RowsAffected()
	if err != nil {
		return errors.New(errors.RowsAffectedFailure).Wrap(err)
	}

	if row < 1 {
		return errors.New(errors.MaterialGroupNotFound)
	}

	return nil
}

func (r *Repository) DeleteMaterialTypeByCode(ctx context.Context, code string) *errors.Error {
	_, err := r.db.ExecContext(ctx, DeleteMaterialTypeQuery, code)
	if err != nil {
		return errors.New(errors.RunQueryFailure).Wrap(err)
	}

	return nil
}

func (r *Repository) DeleteMaterialUoMByCode(ctx context.Context, code string) *errors.Error {
	_, err := r.db.ExecContext(ctx, DeleteMaterialUoMQuery, code)
	if err != nil {
		return errors.New(errors.RunQueryFailure).Wrap(err)
	}

	return nil
}

func (r *Repository) DeleteMaterialGroupByCode(ctx context.Context, code string) *errors.Error {
	_, err := r.db.ExecContext(ctx, DeleteMaterialGroupQuery, code)
	if err != nil {
		return errors.New(errors.RunQueryFailure).Wrap(err)
	}

	return nil
}
