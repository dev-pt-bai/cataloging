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

const CreateRequestQuery = `
INSERT INTO requests (id, subject, is_new, requested_by, status)
	SELECT ?, ?, ?, ?, ?
	WHERE EXISTS(SELECT 1 FROM users WHERE id = ? AND deleted_at = 0)`

const CreateMaterialQuery = `
INSERT INTO materials (id, plant_code, number, type, uom, manufacturer, group_code, equipment_code, short_text, long_text, note, status, request_id)
	SELECT ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
	WHERE EXISTS(SELECT 1 FROM plants WHERE code = ? AND deleted_at = 0)
	AND EXISTS(SELECT 1 FROM material_types WHERE code = ? AND deleted_at = 0)
	AND EXISTS(SELECT 1 FROM material_uoms WHERE code = ? AND deleted_at = 0)
	AND EXISTS(SELECT 1 FROM material_groups WHERE code = ? AND deleted_at = 0)`

const UpdateMaterialAttachmentQuery = `
UPDATE assets SET material_id = ?, updated_at = (UNIX_TIMESTAMP())
	WHERE id = ? AND created_by = ? AND material_id IS NULL AND deleted_at = 0`

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
		res, err = stmt1.ExecContext(ctx, m.ID, m.Plant.Code, m.Number, m.Type.Code, m.UoM.Code, m.Manufacturer, m.Group.Code, m.EquipmentCode, m.ShortText, m.LongText, m.Note, m.Status, request.ID, m.Plant.Code, m.Type.Code, m.UoM.Code, m.Group.Code)
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
