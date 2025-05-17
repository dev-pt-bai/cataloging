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
INSERT INTO material_types (code, description, val_class)
	VALUES (?, ?, ?)`

const CreateMaterialUoMQuery = `
INSERT INTO material_uoms (code, description)
	VALUES (?, ?)`

const CreateMaterialGroupQuery = `
INSERT INTO material_groups (code, description)
	VALUES (?, ?)`

const CreatePlantQuery = `
INSERT INTO plants (code, description)
	VALUES (?, ?)`

const ListMaterialTypeQuery = `
WITH
	cte1 AS (SELECT JSON_OBJECT('code', code, 'description', description, 'valuationClass', val_class, 'createdAt', created_at, 'updatedAt', updated_at) AS record FROM material_types `

const ListMaterialUoMQuery = `
WITH
	cte1 AS (SELECT JSON_OBJECT('code', code, 'description', description, 'createdAt', created_at, 'updatedAt', updated_at) AS record FROM material_uoms `

const ListMaterialGroupQuery = `
WITH
	cte1 AS (SELECT JSON_OBJECT('code', code, 'description', description, 'createdAt', created_at, 'updatedAt', updated_at) AS record FROM material_groups `

const ListPlantQuery = `
WITH
	cte1 AS (SELECT JSON_OBJECT('code', code, 'description', description, 'createdAt', created_at, 'updatedAt', updated_at) AS record FROM plants `

const GetMaterialTypeQuery = `
SELECT code, description, val_class, created_at, updated_at
	FROM material_types
	WHERE code = ? AND deleted_at = 0`

const GetMaterialUoMQuery = `
SELECT code, description, created_at, updated_at
	FROM material_uoms
	WHERE code = ? AND deleted_at = 0`

const GetMaterialGroupQuery = `
SELECT code, description, created_at, updated_at
	FROM material_groups
	WHERE code = ? AND deleted_at = 0`

const UpdateMaterialTypeQuery = `
UPDATE material_types SET description = ?, val_class = ?, updated_at = (UNIX_TIMESTAMP())
	WHERE code = ? AND deleted_at = 0`

const UpdateMaterialUoMQuery = `
UPDATE material_uoms SET description = ?, updated_at = (UNIX_TIMESTAMP())
	WHERE code = ? AND deleted_at = 0`

const UpdateMaterialGroupQuery = `
UPDATE material_groups SET description = ?, updated_at = (UNIX_TIMESTAMP())
	WHERE code = ? AND deleted_at = 0`

const DeleteMaterialTypeQuery = `
UPDATE material_types SET deleted_at = (UNIX_TIMESTAMP())
	WHERE code = ?`

const DeleteMaterialUoMQuery = `
UPDATE material_uoms SET deleted_at = (UNIX_TIMESTAMP())
	WHERE code = ?`

const DeleteMaterialGroupQuery = `
UPDATE material_groups SET deleted_at = (UNIX_TIMESTAMP())
	WHERE code = ?`

func (r *Repository) CreateMaterialType(ctx context.Context, mt model.MaterialType) *errors.Error {
	_, err := r.db.ExecContext(ctx, CreateMaterialTypeQuery, mt.Code, mt.Description, mt.ValuationClass)
	if err != nil {
		if errors.HasMySQLErrCode(err, 1062) {
			return errors.New(errors.MaterialTypeAlreadyExists).Wrap(err)
		}
		return errors.New(errors.RunQueryFailure).Wrap(err)
	}

	return nil
}

func (r *Repository) CreateMaterialUoM(ctx context.Context, uom model.MaterialUoM) *errors.Error {
	_, err := r.db.ExecContext(ctx, CreateMaterialUoMQuery, uom.Code, uom.Description)
	if err != nil {
		if errors.HasMySQLErrCode(err, 1062) {
			return errors.New(errors.MaterialUoMAlreadyExists).Wrap(err)
		}
		return errors.New(errors.RunQueryFailure).Wrap(err)
	}

	return nil
}

func (r *Repository) CreateMaterialGroup(ctx context.Context, mg model.MaterialGroup) *errors.Error {
	_, err := r.db.ExecContext(ctx, CreateMaterialGroupQuery, mg.Code, mg.Description)
	if err != nil {
		if errors.HasMySQLErrCode(err, 1062) {
			return errors.New(errors.MaterialGroupAlreadyExists).Wrap(err)
		}
		return errors.New(errors.RunQueryFailure).Wrap(err)
	}

	return nil
}

func (r *Repository) CreatePlant(ctx context.Context, p model.Plant) *errors.Error {
	_, err := r.db.ExecContext(ctx, CreatePlantQuery, p.Code, p.Description)
	if err != nil {
		if errors.HasMySQLErrCode(err, 1062) {
			return errors.New(errors.MaterialGroupAlreadyExists).Wrap(err)
		}
		return errors.New(errors.RunQueryFailure).Wrap(err)
	}

	return nil
}

func (r *Repository) ListMaterialTypes(ctx context.Context, criteria model.ListMaterialTypesCriteria) (*model.MaterialTypes, *errors.Error) {
	query, args, err := r.buildListMaterialTypesQuery(criteria)
	if err != nil {
		return nil, errors.New(errors.BuildQueryFailure).Wrap(err)
	}

	mts := new(model.MaterialTypes)
	err = r.db.QueryRowContext(ctx, query, args...).Scan(&mts)
	if err != nil && err != sql.ErrNoRows {
		return nil, errors.New(errors.RunQueryFailure).Wrap(err)
	}

	return mts, nil
}

func (r *Repository) ListMaterialUoMs(ctx context.Context, criteria model.ListMaterialUoMsCriteria) (*model.MaterialUoMs, *errors.Error) {
	query, args, err := r.buildListMaterialUoMsQuery(criteria)
	if err != nil {
		return nil, errors.New(errors.BuildQueryFailure).Wrap(err)
	}

	uoms := new(model.MaterialUoMs)
	err = r.db.QueryRowContext(ctx, query, args...).Scan(&uoms)
	if err != nil && err != sql.ErrNoRows {
		return nil, errors.New(errors.RunQueryFailure).Wrap(err)
	}

	return uoms, nil
}

func (r *Repository) ListMaterialGroups(ctx context.Context, criteria model.ListMaterialGroupsCriteria) (*model.MaterialGroups, *errors.Error) {
	query, args, err := r.buildListMaterialGroupsQuery(criteria)
	if err != nil {
		return nil, errors.New(errors.BuildQueryFailure).Wrap(err)
	}

	mgs := new(model.MaterialGroups)
	err = r.db.QueryRowContext(ctx, query, args...).Scan(&mgs)
	if err != nil && err != sql.ErrNoRows {
		return nil, errors.New(errors.RunQueryFailure).Wrap(err)
	}

	return mgs, nil
}

func (r *Repository) ListPlants(ctx context.Context, criteria model.ListPlantsCriteria) (*model.Plants, *errors.Error) {
	query, args, err := r.buildListPlantsQuery(criteria)
	if err != nil {
		return nil, errors.New(errors.BuildQueryFailure).Wrap(err)
	}

	p := new(model.Plants)
	err = r.db.QueryRowContext(ctx, query, args...).Scan(&p)
	if err != nil && err != sql.ErrNoRows {
		return nil, errors.New(errors.RunQueryFailure).Wrap(err)
	}

	return p, nil
}

type listParam struct {
	q    strings.Builder
	args []any
}

func (r *Repository) buildListMaterialTypesQuery(criteria model.ListMaterialTypesCriteria) (string, []any, error) {
	param := listParam{
		q:    strings.Builder{},
		args: make([]any, 0, 5),
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
	whereClauses = append(whereClauses, "deleted_at = 0 ")

	if len(filter.Description) != 0 {
		whereClauses = append(whereClauses, "description LIKE ? ")
		param.args = append(param.args, fmt.Sprintf("%%%s%%", filter.Description))
	}

	param.q.WriteString(fmt.Sprintf("WHERE %s ", strings.Join(whereClauses, "AND ")))
}

func (r *Repository) buildListMaterialUoMsQuery(criteria model.ListMaterialUoMsCriteria) (string, []any, error) {
	param := listParam{
		q:    strings.Builder{},
		args: make([]any, 0, 5),
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
	whereClauses = append(whereClauses, "deleted_at = 0 ")

	if len(filter.Description) != 0 {
		whereClauses = append(whereClauses, "description LIKE ? ")
		param.args = append(param.args, fmt.Sprintf("%%%s%%", filter.Description))
	}

	param.q.WriteString(fmt.Sprintf("WHERE %s ", strings.Join(whereClauses, "AND ")))
}

func (r *Repository) buildListMaterialGroupsQuery(criteria model.ListMaterialGroupsCriteria) (string, []any, error) {
	param := listParam{
		q:    strings.Builder{},
		args: make([]any, 0, 5),
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
	whereClauses = append(whereClauses, "deleted_at = 0 ")

	if len(filter.Description) != 0 {
		whereClauses = append(whereClauses, "description LIKE ? ")
		param.args = append(param.args, fmt.Sprintf("%%%s%%", filter.Description))
	}

	param.q.WriteString(fmt.Sprintf("WHERE %s ", strings.Join(whereClauses, "AND ")))
}

func (r *Repository) buildListPlantsQuery(criteria model.ListPlantsCriteria) (string, []any, error) {
	param := listParam{
		q:    strings.Builder{},
		args: make([]any, 0, 5),
	}
	param.q.WriteString(ListPlantQuery)

	r.filterPlant(criteria.FilterPlant, &param)
	if err := r.sort(criteria.Sort, &param, model.IsAvailableToSortPlant); err != nil {
		return "", nil, err
	}
	if err := r.paginate(criteria.Page, &param); err != nil {
		return "", nil, err
	}

	return param.q.String(), param.args, nil
}

func (r *Repository) filterPlant(filter model.FilterPlant, param *listParam) {
	whereClauses := make([]string, 0, 5)
	whereClauses = append(whereClauses, "deleted_at = 0 ")

	if len(filter.Description) != 0 {
		whereClauses = append(whereClauses, "description LIKE ? ")
		param.args = append(param.args, fmt.Sprintf("%%%s%%", filter.Description))
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

func (r *Repository) GetMaterialType(ctx context.Context, code string) (*model.MaterialType, *errors.Error) {
	mt := new(model.MaterialType)
	err := r.db.QueryRowContext(ctx, GetMaterialTypeQuery, code).Scan(&mt.Code, &mt.Description, &mt.ValuationClass, &mt.CreatedAt, &mt.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New(errors.MaterialTypeNotFound)
		}
		return nil, errors.New(errors.RunQueryFailure).Wrap(err)
	}

	return mt, nil
}

func (r *Repository) GetMaterialUoM(ctx context.Context, code string) (*model.MaterialUoM, *errors.Error) {
	uom := new(model.MaterialUoM)
	err := r.db.QueryRowContext(ctx, GetMaterialUoMQuery, code).Scan(&uom.Code, &uom.Description, &uom.CreatedAt, &uom.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New(errors.MaterialUoMNotFound)
		}
		return nil, errors.New(errors.RunQueryFailure).Wrap(err)
	}

	return uom, nil
}

func (r *Repository) GetMaterialGroup(ctx context.Context, code string) (*model.MaterialGroup, *errors.Error) {
	mg := new(model.MaterialGroup)
	err := r.db.QueryRowContext(ctx, GetMaterialGroupQuery, code).Scan(&mg.Code, &mg.Description, &mg.CreatedAt, &mg.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New(errors.MaterialGroupNotFound)
		}
		return nil, errors.New(errors.RunQueryFailure).Wrap(err)
	}

	return mg, nil
}

func (r *Repository) UpdateMaterialType(ctx context.Context, mt model.MaterialType) *errors.Error {
	res, err := r.db.ExecContext(ctx, UpdateMaterialTypeQuery, mt.Description, mt.ValuationClass, mt.Code)
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
	res, err := r.db.ExecContext(ctx, UpdateMaterialUoMQuery, uom.Description, uom.Code)
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
	res, err := r.db.ExecContext(ctx, UpdateMaterialGroupQuery, mg.Description, mg.Code)
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

func (r *Repository) DeleteMaterialType(ctx context.Context, code string) *errors.Error {
	_, err := r.db.ExecContext(ctx, DeleteMaterialTypeQuery, code)
	if err != nil {
		return errors.New(errors.RunQueryFailure).Wrap(err)
	}

	return nil
}

func (r *Repository) DeleteMaterialUoM(ctx context.Context, code string) *errors.Error {
	_, err := r.db.ExecContext(ctx, DeleteMaterialUoMQuery, code)
	if err != nil {
		return errors.New(errors.RunQueryFailure).Wrap(err)
	}

	return nil
}

func (r *Repository) DeleteMaterialGroup(ctx context.Context, code string) *errors.Error {
	_, err := r.db.ExecContext(ctx, DeleteMaterialGroupQuery, code)
	if err != nil {
		return errors.New(errors.RunQueryFailure).Wrap(err)
	}

	return nil
}
