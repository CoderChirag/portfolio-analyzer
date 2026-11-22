package stocks

import (
	"coderchirag/folio-tracker/lib/domain"
)

type Provider interface {
	DataSource() domain.DataSourceCode
}
