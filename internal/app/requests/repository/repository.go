package repository

import (
	"context"
	"database/sql"

	"github.com/dev-pt-bai/cataloging/internal/model"
	"github.com/dev-pt-bai/cataloging/internal/pkg/errors"
	"github.com/google/uuid"
)

type Repository struct {
	db *sql.DB
}

func New(db *sql.DB) *Repository {
	return &Repository{db: db}
}

const CreateRequestQuery = `
INSERT INTO requests (id, subject, is_new, requested_by, status)
	SELECT ?, ?, ?, ?, ?
	WHERE EXISTS(SELECT 1 FROM users WHERE id = ? AND deleted_at = 0)`

const CreateMaterialQuery = `
INSERT INTO materials (id, number, plant_code, type_code, uom_code, group_code, equipment_code, manufacturer, short_text, long_text, note, status, request_id)
	SELECT ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
	WHERE EXISTS(SELECT 1 FROM plants WHERE code = ? AND deleted_at = 0)
	AND EXISTS(SELECT 1 FROM material_types WHERE code = ? AND deleted_at = 0)
	AND EXISTS(SELECT 1 FROM material_uoms WHERE code = ? AND deleted_at = 0)
	AND EXISTS(SELECT 1 FROM material_groups WHERE code = ? AND deleted_at = 0)`

const UpdateMaterialAttachmentQuery = `
UPDATE assets SET material_id = ?, updated_at = (UNIX_TIMESTAMP())
	WHERE id = ? AND created_by = ? AND material_id IS NULL AND deleted_at = 0`

const GetRequestQuery = `
WITH
    cte1 AS (SELECT id, subject, is_new, requested_by, status, created_at, updated_at FROM requests WHERE id = ? AND deleted_at = 0),
	cte2 AS (SELECT JSON_OBJECT('id', id, 'name', name, 'email', email, 'isAdmin', is_admin, 'isVerified', is_verified, 'createdAt', created_at, 'updatedAt', updated_at) AS requester FROM users WHERE id = (SELECT requested_by FROM cte1) AND deleted_at = 0),
	cte3 AS (SELECT id, number, plant_code, type_code, uom_code, group_code, equipment_code, manufacturer, short_text, long_text, note, status, request_id, created_at, updated_at FROM materials WHERE request_id = (SELECT id FROM cte1) AND deleted_at = 0),
	cte4 AS (SELECT code, JSON_OBJECT('code', code, 'description', description, 'createdAt', created_at, 'updatedAt', updated_at) AS plant FROM plants WHERE code IN (SELECT plant_code FROM cte3) AND deleted_at = 0),
	cte5 AS (SELECT code, JSON_OBJECT('code', code, 'description', description, 'createdAt', created_at, 'updatedAt', updated_at) AS type FROM material_types WHERE code IN (SELECT type_code FROM cte3) AND deleted_at = 0),
	cte6 AS (SELECT code, JSON_OBJECT('code', code, 'description', description, 'createdAt', created_at, 'updatedAt', updated_at) AS uom FROM material_uoms WHERE code IN (SELECT uom_code FROM cte3) AND deleted_at = 0),
	cte7 AS (SELECT code, JSON_OBJECT('code', code, 'description', description, 'createdAt', created_at, 'updatedAt', updated_at) AS mgroup FROM material_groups WHERE code IN (SELECT group_code FROM cte3) AND deleted_at = 0),
	cte8 AS (SELECT material_id, JSON_ARRAYAGG(JSON_OBJECT('id', id, 'name', name, 'size', size, 'downloadURL', download_url, 'webURL', web_url, 'createdBy', created_by, 'materialID', material_id, "createdAt", created_at, 'updatedAt', updated_at)) AS attachments FROM assets WHERE material_id IN (SELECT id FROM cte3) AND deleted_at = 0 GROUP BY material_id),
	cte9 AS (SELECT JSON_ARRAYAGG(JSON_OBJECT('id', id, 'number', number, 'plant', plant, 'type', type, 'uom', uom, 'group', mgroup, 'equipmentCode', equipment_code, 'manufacturer', manufacturer, 'shortText', short_text, 'longText', long_text, 'note', note, 'status', status, 'requestID', request_id, 'createdAt', created_at, 'updatedAt', updated_at, 'attachments', attachments)) AS materials FROM cte3 JOIN cte4 ON cte3.plant_code = cte4.code JOIN cte5 ON cte3.type_code = cte5.code JOIN cte6 ON cte3.uom_code = cte6.code JOIN cte7 ON cte3.group_code = cte7.code JOIN cte8 ON cte3.id = cte8.material_id)
SELECT JSON_OBJECT('id', id, 'subject', subject, 'is_new', is_new, 'requestedBy', requester, 'status', status, 'createdAt', created_at, 'updatedAt', updated_at, 'materials', materials) AS request
	FROM cte1, cte2, cte9`

func (r *Repository) CreateRequest(ctx context.Context, request model.Request) *errors.Error {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelDefault})
	if err != nil {
		return errors.New(errors.StartingTransactionFailure).Wrap(err)
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx, CreateRequestQuery, request.ID, request.Subject, request.IsNew, request.RequestedBy.ID, request.Status, request.RequestedBy.ID)
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

	stmt1, err := tx.PrepareContext(ctx, CreateMaterialQuery)
	if err != nil {
		return errors.New(errors.PrepareStatementFailure).Wrap(err)
	}
	defer stmt1.Close()

	stmt2, err := tx.PrepareContext(ctx, UpdateMaterialAttachmentQuery)
	if err != nil {
		return errors.New(errors.PrepareStatementFailure).Wrap(err)
	}
	defer stmt2.Close()

	for i := range request.Materials {
		m := request.Materials[i]
		res, err = stmt1.ExecContext(ctx, m.ID, m.Number, m.Plant.Code, m.Type.Code, m.UoM.Code, m.Group.Code, m.EquipmentCode, m.Manufacturer, m.ShortText, m.LongText, m.Note, m.Status, request.ID, m.Plant.Code, m.Type.Code, m.UoM.Code, m.Group.Code)
		if err != nil {
			return errors.New(errors.RunQueryFailure).Wrap(err)
		}

		row, err = res.RowsAffected()
		if err != nil {
			return errors.New(errors.RowsAffectedFailure).Wrap(err)
		}

		if row < 1 {
			return errors.New(errors.MaterialPropertiesNotFound)
		}

		for j := range m.Attachments {
			a := m.Attachments[j]
			res, err = stmt2.ExecContext(ctx, m.ID, a.ID, request.RequestedBy.ID)
			if err != nil {
				return errors.New(errors.RunQueryFailure).Wrap(err)
			}

			row, err = res.RowsAffected()
			if err != nil {
				return errors.New(errors.RowsAffectedFailure).Wrap(err)
			}

			if row < 1 {
				return errors.New(errors.AssetNotFound)
			}
		}
	}

	if err = tx.Commit(); err != nil {
		return errors.New(errors.CommittingTransactionFailure).Wrap(err)
	}

	return nil
}

func (r *Repository) GetRequest(ctx context.Context, ID uuid.UUID) (*model.Request, *errors.Error) {
	request := new(model.Request)
	err := r.db.QueryRowContext(ctx, GetRequestQuery, ID).Scan(&request)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New(errors.RequestNotFound)
		}
		return nil, errors.New(errors.RunQueryFailure).Wrap(err)
	}

	return request, nil
}
