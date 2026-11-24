package importsrepo

import (
	"coderchirag/folio-tracker/lib/db/repository/internal/jetgen/model"
	"coderchirag/folio-tracker/lib/domain"
)

func toImportModel(d *domain.Import) *model.Imports {
	imp := &model.Imports{
		ID:           d.ID,
		UserID:       d.UserID,
		DataSourceID: d.DataSourceID,
		Kind:         d.Kind,
		FileName:     d.FileName,
		FileSize:     d.FileSize,
		FileSha256:   d.FileSHA256,
		Status:       d.Status,

		RowsProcessed: d.RowsProcessed,
		RowsImported:  d.RowsImported,
		RowsDuplicate: d.RowsDuplicate,
		RowsFailed:    d.RowsFailed,

		CreatedAt: d.CreatedAt,
		UpdatedAt: d.UpdatedAt,
	}

	if d.RowsTotal != 0 {
		imp.RowsTotal = &d.RowsTotal
	}
	if d.ErrorCode != "" {
		imp.ErrorCode = &d.ErrorCode
	}
	if d.ErrorMsg != "" {
		imp.ErrorMsg = &d.ErrorMsg
	}
	if !d.StartedAt.IsZero() {
		imp.StartedAt = &d.StartedAt
	}
	if !d.FinishedAt.IsZero() {
		imp.FinishedAt = &d.FinishedAt
	}
	return imp
}

func toIDUserAndStatusDomain(m *model.Imports) *domain.ImportIDUserAndStatus {
	return &domain.ImportIDUserAndStatus{
		ID:     m.ID,
		UserID: m.UserID,
		Status: m.Status,
	}
}
