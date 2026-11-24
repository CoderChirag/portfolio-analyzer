package httpuploadshandler

import "coderchirag/folio-tracker/lib/domain"

type ImportJobStatusRes struct {
	JobID        string `json:"jobId"`
	ImportID     string `json:"importId"`
	JobStatus    string `json:"jobStatus"`
	ImportStatus string `json:"importStatus"`
}

func (j *ImportJobStatusRes) FromDomain(s *domain.ImportJobStatus) *ImportJobStatusRes {
	j.JobID = s.JobID.String()
	j.ImportID = s.ImportID.String()
	j.JobStatus = string(s.JobStatus)
	j.ImportStatus = string(s.ImportStatus)
	return j
}
