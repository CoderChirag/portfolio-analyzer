package stocksgroww

// import (
// 	"coderchirag/folio-tracker/lib/apperr"
// 	"coderchirag/folio-tracker/lib/domain"
// 	"coderchirag/folio-tracker/lib/service/instruments/stocks"
// 	"coderchirag/folio-tracker/lib/utils"
// 	"fmt"
// 	"strings"

// 	"github.com/shopspring/decimal"
// )

// const (
// 	stockOrdersSheetName             = "Sheet1"
// 	stockOrdersColumnStockName       = "stock_name"
// 	stockOrdersColumnSymbol          = "symbol"
// 	stockOrdersColumnISIN            = "isin"
// 	stockOrdersColumnType            = "type"
// 	stockOrdersColumnQuantity        = "quantity"
// 	stockOrdersColumnAmount          = "value"
// 	stockOrdersColumnExchange        = "exchange"
// 	stockOrdersColumnExchangeOrderID = "exchange_order_id"
// 	stockOrdersColumnExecutedAt      = "execution_date_and_time"
// 	stockOrdersColumnOrderStatus     = "order_status"

// 	stockOrderStockNameNull  = "null"
// 	stockOrderStatusExecuted = "executed"
// 	// Go time layout
// 	// 02: signifies day, 01: signifies month
// 	// 2006: year, 15: hour, 04: minute, PM: AM/PM
// 	stockOrdersTimeLayout    = "02-01-2006 15:04 PM"
// 	stockOrdersTimeLayoutMsg = "DD-MM-YYYY HH:MM AM"
// )

// type stockOrders struct {
// 	headers map[string]string //header to column map
// 	colMap  map[string]string //column to header map
// }

// func newStockOrders() *stockOrders {
// 	return &stockOrders{
// 		headers: map[string]string{
// 			stockOrdersColumnStockName:       "",
// 			stockOrdersColumnSymbol:          "",
// 			stockOrdersColumnISIN:            "",
// 			stockOrdersColumnType:            "",
// 			stockOrdersColumnQuantity:        "",
// 			stockOrdersColumnAmount:          "",
// 			stockOrdersColumnExchange:        "",
// 			stockOrdersColumnExchangeOrderID: "",
// 			stockOrdersColumnExecutedAt:      "",
// 			stockOrdersColumnOrderStatus:     "",
// 		},
// 		colMap: map[string]string{},
// 	}
// }

// func (o *stockOrders) sheetName() string {
// 	return stockOrdersSheetName
// }

// func (o *stockOrders) parseHeadersRow(row *domain.CSVRow) (bool, error) {
// 	var errs apperr.ValidationError
// 	if row.Error != nil {
// 		errs.Add(fmt.Sprintf("row: %d", row.Index), fmt.Errorf("%w: %w", ErrRowError, row.Error))
// 		return false, errs
// 	}

// 	found := false
// 	for _, c := range row.Columns {
// 		colName := normalizeColumnName(c.Value)
// 		if _, ok := o.headers[colName]; ok {
// 			found = true
// 			o.headers[colName] = c.Column
// 			o.colMap[c.Column] = colName
// 		}
// 	}

// 	missingHeaders := o.missingHeaders()
// 	if found && len(missingHeaders) > 0 {
// 		errs.Add(strings.Join(missingHeaders, ", "), ErrMissingHeaders)
// 		return found, errs
// 	}

// 	return found, nil
// }

// func (o *stockOrders) parseRow(row *domain.CSVRow) (*stocks.StockOrder, error) {
// 	var errs apperr.ValidationError

// 	missingHeaders := o.missingHeaders()
// 	if len(missingHeaders) > 0 {
// 		errs.Add(strings.Join(missingHeaders, ", "), ErrMissingHeaders)
// 		return nil, errs
// 	}

// 	if row.Error != nil {
// 		errs.Add(fmt.Sprintf("row: %d", row.Index), fmt.Errorf("%w: %w", ErrRowError, row.Error))
// 		return nil, errs
// 	}

// 	m := map[string]*domain.CSVColumn{}
// 	for _, c := range row.Columns {
// 		if h, ok := o.colMap[c.Column]; ok {
// 			v := strings.TrimSpace(c.Value)

// 			skip, value, err := o.evaluatePerRowConditions(c, h, v)
// 			if err != nil {
// 				return nil, err
// 			}
// 			if skip {
// 				return nil, nil
// 			}

// 			c.Value = value
// 			m[h] = c
// 		}
// 	}

// 	missing := o.missingColumns(m)
// 	if len(missing) > 0 {
// 		errs.Add(strings.Join(missing, ", "), ErrMissingColumns)
// 		return nil, errs
// 	}

// 	return o.toStockOrder(m)
// }

// func (o *stockOrders) evaluatePerRowConditions(c *domain.CSVColumn, header string, value string) (bool, string, error) {
// 	var errs apperr.ValidationError
// 	skip := false
// 	// null values represents bonds and needs to be skipped
// 	if header == stockOrdersColumnStockName && strings.ToLower(value) == stockOrderStockNameNull {
// 		skip = true
// 		return skip, value, nil
// 	}

// 	// only processes executed orders
// 	if header == stockOrdersColumnOrderStatus && strings.ToLower(value) != stockOrderStatusExecuted {
// 		skip = true
// 		return skip, value, nil
// 	}

// 	if value == "" {
// 		errs.Add(fmt.Sprintf("cell: %s%d", c.Column, c.Row), ErrInvalidValue)
// 		return skip, value, errs
// 	}

// 	return skip, value, nil
// }

// func (o *stockOrders) toStockOrder(m map[string]*domain.CSVColumn) (*stocks.StockOrder, error) {
// 	var errs apperr.ValidationError
// 	order := &stocks.StockOrder{}
// 	for h, c := range m {
// 		switch h {
// 		case stockOrdersColumnStockName:
// 			order.StockName = c.Value
// 		case stockOrdersColumnSymbol:
// 			order.Symbol = c.Value
// 		case stockOrdersColumnISIN:
// 			order.ISIN = c.Value
// 		case stockOrdersColumnType:
// 			order.Type = c.Value
// 		case stockOrdersColumnQuantity:
// 			qty, err := decimal.NewFromString(c.Value)
// 			if err != nil {
// 				errs.Add(
// 					fmt.Sprintf("cell: %s%d", c.Column, c.Row),
// 					ErrDecimalParse,
// 				)
// 				return nil, errs
// 			}
// 			order.Quantity = qty
// 		case stockOrdersColumnAmount:
// 			amount, err := decimal.NewFromString(c.Value)
// 			if err != nil {
// 				errs.Add(
// 					fmt.Sprintf("cell: %s%d", c.Column, c.Row),
// 					ErrDecimalParse,
// 				)
// 				return nil, errs
// 			}
// 			order.Amount = amount
// 		case stockOrdersColumnExchange:
// 			order.Exchange = c.Value
// 		case stockOrdersColumnExchangeOrderID:
// 			order.ExchangeOrderID = c.Value
// 		case stockOrdersColumnExecutedAt:
// 			executedAt, err := utils.ParseTimeWithIST(stockOrdersTimeLayout, c.Value)
// 			if err != nil {
// 				errs.Add(
// 					fmt.Sprintf("cell: %s%d", c.Column, c.Row),
// 					fmt.Errorf("%w:  expected format: %s", ErrTimeParse, stockOrdersTimeLayoutMsg),
// 				)
// 				return nil, errs
// 			}
// 			order.ExecutedAt = executedAt
// 		case stockOrdersColumnOrderStatus:
// 			order.OrderStatus = c.Value
// 		}
// 	}

// 	return order, nil
// }

// func (o *stockOrders) missingHeaders() []string {
// 	missing := []string{}
// 	for k, v := range o.headers {
// 		if v == "" {
// 			missing = append(missing, k)
// 		}
// 	}
// 	return missing
// }

// func (o *stockOrders) missingColumns(m map[string]*domain.CSVColumn) []string {
// 	colMap := map[string]bool{}
// 	for h := range o.headers {
// 		colMap[h] = false
// 	}

// 	for h := range m {
// 		if _, ok := colMap[h]; ok {
// 			colMap[h] = true
// 		}
// 	}

// 	missing := []string{}
// 	for h, v := range colMap {
// 		if !v {
// 			missing = append(missing, h)
// 		}
// 	}

// 	return missing
// }
