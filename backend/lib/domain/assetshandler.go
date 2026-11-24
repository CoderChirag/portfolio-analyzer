package domain

type ImportJobStatus struct {
	UserID       ID
	JobID        ID
	ImportID     ID
	JobStatus    JobStatus
	ImportStatus ImportStatus
}

type GetImportStatusParams struct {
	UserID   ID
	JobID    ID
	ImportID ID
}
