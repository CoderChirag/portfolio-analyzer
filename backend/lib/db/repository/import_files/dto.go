package importfilesrepo

import (
	"coderchirag/folio-tracker/lib/db/repository/internal/jetgen/model"
	"coderchirag/folio-tracker/lib/domain"
)

func toImportFileModel(d *domain.ImportFile) *model.ImportFiles {
	return &model.ImportFiles{
		ImportID:    d.ImportID,
		ContentType: d.ContentType,
		Content:     d.Content,
		CreatedAt:   d.CreatedAt,
	}
}
