package stocksgroww

import (
	"coderchirag/folio-tracker/lib/domain"
	"coderchirag/folio-tracker/lib/service/instruments/stocks"
)

const (
	logPrefix                 = "provider:Groww"
	parseStockOrdersLogPrefix = "ParseStockOrders"
)

var _ stocks.Provider = (*Provider)(nil)

type Provider struct {
	excelParser ExcelParser
}

func New(excelParser ExcelParser) *Provider {
	return &Provider{
		excelParser: excelParser,
	}
}

func (s *Provider) DataSource() domain.DataSourceCode {
	return domain.SourceGroww
}

// func (s *Provider) ParseStockOrders(p *domain.ParseImportFileParams) ([]*stocks.StockOrder, error) {

// 	xlData, err := s.xlParser.Parse(p.File, p.Header)
// 	if err != nil {
// 		return nil, fmt.Errorf(
// 			"%s.%s: %w: %w",
// 			logPrefix,
// 			parseStockOrdersLogPrefix,
// 			ErrFileParse,
// 			err)
// 	}

// 	return s.parseStockOrdersV1(xlData)
// }

// func (s *Provider) parseStockOrdersV1(xlData *domain.Excel) ([]*stocks.StockOrder, error) {
// 	var errs apperr.ValidationError

// 	var sheet *domain.Sheet
// 	orders := newStockOrders()
// 	for _, s := range xlData.Sheets {
// 		if s.SheetName == orders.sheetName() {
// 			sheet = s
// 			break
// 		}
// 	}
// 	if sheet == nil {
// 		errs.Add("sheet name", fmt.Errorf("no sheet named %s present in the file", orders.sheetName()))
// 		return nil, fmt.Errorf("%s.%s: %w", logPrefix, parseStockOrdersLogPrefix, errs)
// 	}

// 	out := make([]*stocks.StockOrder, 0)
// 	headersFound := false
// 	for len(sheet.Sheet.Rows) > 0 {
// 		row := sheet.Sheet.Rows[0]
// 		sheet.Sheet.Rows = sheet.Sheet.Rows[1:]

// 		if found, err := orders.parseHeadersRow(row); err != nil {
// 			return nil, fmt.Errorf("%s.%s: %w", logPrefix, parseStockOrdersLogPrefix, err)
// 		} else if found {
// 			headersFound = true
// 			break
// 		}
// 	}
// 	if !headersFound {
// 		errs.Add("file", ErrHeadersNotFound)
// 		return nil, fmt.Errorf("%s.%s: %w", logPrefix, parseStockOrdersLogPrefix, errs)
// 	}

// 	for len(sheet.Sheet.Rows) > 0 {
// 		row := sheet.Sheet.Rows[0]
// 		sheet.Sheet.Rows = sheet.Sheet.Rows[1:]

// 		o, err := orders.parseRow(row)
// 		if err != nil {
// 			return nil, fmt.Errorf("%s.%s: %w", logPrefix, parseStockOrdersLogPrefix, err)
// 		}
// 		if o != nil {
// 			out = append(out, o)
// 		}
// 	}

// 	return out, nil
// }
