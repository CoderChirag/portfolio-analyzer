package excel

type Parser struct{}

func New() *Parser {
	return &Parser{}
}

// func (p *Parser) Parse(file domain.File, header *domain.FileHeader) (*domain.Excel, error) {
// 	contentType := header.Header.Get("Content-Type")
// 	var errs apperr.ValidationError
// 	if contentType != string(domain.ContentTypeXLSX) {
// 		errs.Add("Content-Type", fmt.Errorf("invalid content-type: %s, required: %s", contentType, domain.ContentTypeXLSX))
// 		return nil, errs
// 	}
// 	if header.Size > domain.MaxImportFileSize {
// 		return nil, apperr.ErrFileTooLarge
// 	}

// 	zipReader, err := zip.NewReader(file, header.Size)
// 	if err != nil {
// 		return nil, apperr.ErrInvalidFile
// 	}

// 	xl, err := xlsxreader.NewReaderZip(zipReader)
// 	if err != nil {
// 		return nil, apperr.ErrInvalidFile
// 	}

// 	excelData := &domain.Excel{
// 		Sheets: make([]*domain.Sheet, 0),
// 	}

// 	for i, s := range xl.Sheets {
// 		sheet := &domain.Sheet{
// 			SheetName: s,
// 			Sheet: &domain.CSV{
// 				Rows: make([]*domain.CSVRow, 0),
// 			},
// 		}

// 		for r := range xl.ReadRows(xl.Sheets[i]) {
// 			row := &domain.CSVRow{
// 				Index:   r.Index,
// 				Error:   r.Error,
// 				Columns: make([]*domain.CSVColumn, 0),
// 			}

// 			for _, c := range r.Cells {
// 				column := &domain.CSVColumn{
// 					Column: c.Column,
// 					Row:    c.Row,
// 					Value:  c.Value,
// 					Type:   domain.ColumnType(string(c.Type)),
// 				}
// 				row.Columns = append(row.Columns, column)
// 			}

// 			sheet.Sheet.Rows = append(sheet.Sheet.Rows, row)
// 		}

// 		excelData.Sheets = append(excelData.Sheets, sheet)
// 	}

// 	return excelData, nil
// }
