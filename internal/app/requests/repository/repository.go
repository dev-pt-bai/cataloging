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
	VALUES (?, ?, ?, ?, ?)`

const CreateMaterialQuery = `
INSERT INTO materials (id, number, type, uom, plant_code, manufacturer, group_code, equipment_code, short_text, long_text, note, status, request_id)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

func (r *Repository) CreateRequest(ctx context.Context, request model.Request) *errors.Error {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelDefault})
	if err != nil {
		return errors.New(errors.StartingTransactionFailure).Wrap(err)
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, CreateRequestQuery, request.ID, request.Subject, request.IsNew, request.RequestedBy.ID, request.Status)
	if err != nil {
		return errors.New(errors.RunQueryFailure).Wrap(err)
	}

	stmt, err := tx.PrepareContext(ctx, CreateMaterialQuery)
	if err != nil {
		return errors.New(errors.PrepareStatementFailure).Wrap(err)
	}
	defer stmt.Close()

	for i := range request.Materials {
		_, err = stmt.ExecContext(ctx, request.Materials[i].ID)
		if err != nil {
			return errors.New(errors.RunQueryFailure).Wrap(err)
		}
	}

	if err = tx.Commit(); err != nil {
		return errors.New(errors.CommittingTransactionFailure).Wrap(err)
	}

	return nil
}
