package httpuploadshandler

import (
	httpauthmiddleware "coderchirag/folio-tracker/lib/api/http/middlewares/auth"
	httpresponse "coderchirag/folio-tracker/lib/api/http/response"
	"coderchirag/folio-tracker/lib/apperr"
	"coderchirag/folio-tracker/lib/domain"
	"coderchirag/folio-tracker/lib/lookup"
	"coderchirag/folio-tracker/lib/utils"
	"io"
	"net/http"
	"slices"
)

const (
	logPrefix               = "apiHandler:Uploads"
	uploadStockTxnLogPrefix = "UploadStockTransactions"
)

const maxFileSizeInMemoryBytes = 10 << 20 //10 MiB

var stockTxnDS = map[string]*domain.ImportKind{
	"groww": domain.ImportGrowwStockOrders,
}

var allowedFileFormats = map[string][]domain.ContentType{
	domain.ImportGrowwStockOrders.String(): []domain.ContentType{
		domain.ContentTypeXLSX,
	},
}

type HTTPHandler struct {
	svc   AssetsHandlerService
	dsReg *lookup.Registry[domain.DataSourceID, domain.DataSourceCode, domain.DataSource]
}

func New(svc AssetsHandlerService, dsReg *lookup.Registry[domain.DataSourceID, domain.DataSourceCode, domain.DataSource]) *HTTPHandler {
	return &HTTPHandler{
		svc:   svc,
		dsReg: dsReg,
	}
}

func (h *HTTPHandler) ImportStockTransactions(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	var errs apperr.ValidationError
	dataSource := query.Get("datasource")
	if dataSource == "" {
		errs.Add("datasource", ErrDataSourceRequired)
		httpresponse.JSONError(w, r, errs)
		return
	}
	importKind, ok := stockTxnDS[dataSource]
	if !ok {
		var errs apperr.ValidationError
		errs.Add("datasource", domain.ErrInvalidDataSource)
		httpresponse.JSONError(w, r, errs)
		return
	}
	ds, ok := h.dsReg.ByCode(importKind.DataSource())
	if !ok {
		httpresponse.JSONError(w, r, domain.ErrInvalidDataSource)
		return
	}

	uid := r.Context().Value(httpauthmiddleware.UserIDKey)
	userID, ok := uid.(domain.ID)
	if !ok {
		httpresponse.JSONError(w, r, httpresponse.ErrUnauthorized)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, domain.MaxImportFileSize+1<<20) // file + form overhead
	mr, err := r.MultipartReader()
	if err != nil {
		httpresponse.JSONError(w, r, httpresponse.ErrBadRequest)
		return
	}

	p := &domain.ImportParams{
		UserID:       userID,
		DataSourceID: ds.ID,
		Kind:         *importKind,
	}
	var contentType domain.ContentType
	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		} else if err != nil {
			httpresponse.JSONError(w, r, httpresponse.ErrBadRequest)
			return
		}

		switch part.FormName() {
		case "file":
			defer part.Close()
			contentType = domain.ContentType(part.Header.Get("Content-Type"))
			if !slices.Contains(
				allowedFileFormats[p.Kind.String()],
				contentType,
			) {
				errs.Add("Content-Type", domain.ErrImportFileContentTypeNotSupported)
				httpresponse.JSONError(w, r, errs)
				return
			}
			fileName := part.FileName()
			if fileName == "" {
				httpresponse.JSONError(w, r, apperr.ErrInvalidFileName)
				return
			}
			content, err := utils.ReadLimited(part, domain.MaxImportFileSize, r.ContentLength)
			if err != nil {
				httpresponse.JSONError(w, r, err)
				return
			}
			p.Content = content
			p.FileName = fileName
			break
		default:
			part.Close()
		}
	}
	if p.Content == nil {
		httpresponse.JSONError(w, r, httpresponse.ErrMultipartFormFile)
		return
	}

	status, err := h.svc.Import(r.Context(), p, contentType)
	if err != nil {
		httpresponse.JSONError(w, r, err)
		return
	}

	httpresponse.JSON(w, http.StatusAccepted, (&ImportJobStatusRes{}).FromDomain(status))
}

func (h *HTTPHandler) ImportStatus(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	var errs apperr.ValidationError
	j := query.Get("job_id")
	if j == "" {
		errs.Add("job_id", ErrJobIDRequired)
		httpresponse.JSONError(w, r, errs)
		return
	}
	jobID, err := domain.ParseID(j)
	if err != nil {
		errs.Add("job_id", ErrJobIDInvalid)
		httpresponse.JSONError(w, r, errs)
		return
	}

	i := query.Get("import_id")
	if i == "" {
		errs.Add("import_id", ErrImportIDRequired)
		httpresponse.JSONError(w, r, errs)
		return
	}
	importID, err := domain.ParseID(i)
	if err != nil {
		errs.Add("import_id", ErrImportIDInvalid)
		httpresponse.JSONError(w, r, errs)
		return
	}

	uid := r.Context().Value(httpauthmiddleware.UserIDKey)
	userID, ok := uid.(domain.ID)
	if !ok {
		httpresponse.JSONError(w, r, httpresponse.ErrUnauthorized)
		return
	}

	status, err := h.svc.GetImportStatus(r.Context(), &domain.GetImportStatusParams{
		UserID:   userID,
		JobID:    jobID,
		ImportID: importID,
	})
	if err != nil {
		httpresponse.JSONError(w, r, err)
		return
	}

	httpresponse.JSON(w, http.StatusOK, (&ImportJobStatusRes{}).FromDomain(status))
}
