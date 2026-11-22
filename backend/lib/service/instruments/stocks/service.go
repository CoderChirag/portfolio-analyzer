package stocks

import (
	"coderchirag/folio-tracker/lib/domain"
)

const (
	logPrefix                = "instrument:stocks"
	parseImportDataLogPrefix = "ParseImportData"
)

type Instrument struct {
	providers []Provider
}

func New(providers []Provider) *Instrument {
	return &Instrument{
		providers: providers,
	}
}

func (s *Instrument) AssetType() []domain.AssetTypeCode {
	return []domain.AssetTypeCode{
		domain.AssetStock,
		domain.AssetETF,
	}
}

func (s *Instrument) getProvider(ds domain.DataSourceCode) Provider {
	for _, pr := range s.providers {
		if pr.DataSource() == ds {
			return pr
		}
	}
	return nil
}
