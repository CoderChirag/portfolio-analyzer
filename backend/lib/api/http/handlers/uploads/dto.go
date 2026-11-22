package httpuploadshandler

import "coderchirag/folio-tracker/lib/domain"

type JobStatusResponse struct {
	ID     string `json:"jobId"`
	Status string `json:"status"`
}

func (j *JobStatusResponse) FromDomain(s *domain.Job) *JobStatusResponse {
	j.ID = s.ID.String()
	j.Status = string(s.Status)
	return j
}
