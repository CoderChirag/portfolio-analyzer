package httpuploadshandler

import "coderchirag/folio-tracker/lib/domain"

type ImportJobStatusRes struct {
	RequestID    string `json:"requestId,omitempty"`
	JobID        string `json:"jobId"`
	ImportID     string `json:"importId"`
	JobStatus    string `json:"jobStatus"`
	ImportStatus string `json:"importStatus"`
}

func (j *ImportJobStatusRes) FromDomain(s *domain.ImportJobStatus, reqID string) *ImportJobStatusRes {
	j.RequestID = reqID
	j.JobID = s.JobID.String()
	j.ImportID = s.ImportID.String()
	j.JobStatus = string(s.JobStatus)
	j.ImportStatus = string(s.ImportStatus)
	return j
}
