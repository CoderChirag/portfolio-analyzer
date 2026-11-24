package jobsrepo

import (
	"coderchirag/folio-tracker/lib/db/repository/internal/jetgen/model"
	"coderchirag/folio-tracker/lib/domain"
)

func toJobModel(d *domain.Job) *model.Jobs {
	job := &model.Jobs{
		ID:      d.ID,
		JobType: d.Type,
		Payload: d.Payload,

		Status:      d.Status,
		Attempts:    d.Attempts,
		MaxAttempts: d.MaxAttempts,
		AvailableAt: d.AvailableAt,

		CreatedAt: d.CreatedAt,
		UpdatedAt: d.UpdatedAt,
	}

	if d.DedupeKey != "" {
		job.DedupeKey = &d.DedupeKey
	}
	if !d.UserID.IsZero() {
		job.UserID = &d.UserID
	}
	if d.LockedBy != "" {
		job.LockedBy = &d.LockedBy
	}
	if d.LastError != "" {
		job.LastError = &d.LastError
	}
	if !d.FinishedAt.IsZero() {
		job.FinishedAt = &d.FinishedAt
	}

	return job
}

func toIDUserAndStatusDomain(m *model.Jobs) *domain.JobIDUserAndStatus {
	return &domain.JobIDUserAndStatus{
		ID:     m.ID,
		UserID: *m.UserID,
		Status: m.Status,
	}
}
