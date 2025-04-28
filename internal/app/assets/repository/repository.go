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

const CreateAssetQuery = `
INSERT INTO assets (id, name, size, download_url, web_url, created_by, material_id)
	VALUES (?, ?, ?, ?, ?, ?, ?)`

const GetAssetQuery = `
SELECT id, name, size, download_url, web_url, created_by, material_id, created_at, updated_at
	FROM assets
	WHERE id = ? AND deleted_at = 0`

const DeleteAssetQuery = `
UPDATE assets SET deleted_at = (UNIX_TIMESTAMP())
	WHERE id = ? AND created_by = ?`

const DeleteAssetByAdminQuery = `
UPDATE assets SET deleted_at = (UNIX_TIMESTAMP())
	WHERE id = ?`

func (r *Repository) CreateAsset(ctx context.Context, asset model.Asset) *errors.Error {
	_, err := r.db.ExecContext(ctx, CreateAssetQuery, asset.ID, asset.Name, asset.Size, asset.DownloadURL, asset.WebURL, asset.CreatedBy, asset.MaterialID)
	if err != nil {
		if errors.HasMySQLErrCode(err, 1062) {
			return errors.New(errors.AssetAlreadyExists).Wrap(err)
		}
		return errors.New(errors.RunQueryFailure).Wrap(err)
	}

	return nil
}

func (r *Repository) GetAsset(ctx context.Context, ID string) (*model.Asset, *errors.Error) {
	a := new(model.Asset)
	err := r.db.QueryRowContext(ctx, GetAssetQuery, ID).Scan(&a.ID, &a.Name, &a.Size, &a.DownloadURL, &a.WebURL, &a.CreatedBy, &a.MaterialID, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New(errors.AssetNotFound)
		}
		return nil, errors.New(errors.RunQueryFailure).Wrap(err)
	}

	return a, nil
}

func (r *Repository) DeleteAssetByCreator(ctx context.Context, ID string, deleted_by string) *errors.Error {
	_, err := r.db.ExecContext(ctx, DeleteAssetQuery, ID, deleted_by)
	if err != nil {
		return errors.New(errors.RunQueryFailure).Wrap(err)
	}

	return nil
}

func (r *Repository) DeleteAssetByAdmin(ctx context.Context, ID string) *errors.Error {
	_, err := r.db.ExecContext(ctx, DeleteAssetByAdminQuery, ID)
	if err != nil {
		return errors.New(errors.RunQueryFailure).Wrap(err)
	}

	return nil
}
