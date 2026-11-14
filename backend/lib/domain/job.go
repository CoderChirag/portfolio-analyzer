package domain

import (
	"coderchirag/folio-tracker/lib/db/models/types"
	"coderchirag/folio-tracker/lib/utils"
	"coderchirag/folio-tracker/lib/validation"
	"encoding/json"
	"errors"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

// Limits mirror the jobs table columns.
const (
	jobTypeMaxLen      = 64   // VARCHAR(64) ascii
	jobDedupeKeyMaxLen = 255  // VARCHAR(255) ascii
	jobLockedByMaxLen  = 128  // VARCHAR(128), counted in characters
	jobLastErrorMaxLen = 2048 // VARCHAR(2048), counted in characters
)

const DefaultJobMaxAttempts = 5

// JobType names a handler in the worker's registry.
type JobType string

const (
	JobImportProcess JobType = "import.Process"
)

func (j JobType) IsValid() bool {
	switch j {
	case JobImportProcess:
		return true
	}
	return false
}

type JobStatus string

const (
	JobQueued    JobStatus = "queued"
	JobRunning   JobStatus = "running"
	JobSucceeded JobStatus = "succeeded"
	JobFailed    JobStatus = "failed"
	JobCancelled JobStatus = "cancelled"
)

func (s JobStatus) IsValid() bool {
	switch s {
	case JobQueued, JobRunning, JobSucceeded, JobFailed, JobCancelled:
		return true
	}
	return false
}

// IsTerminal reports whether the job will never run again. The repository clears
// dedupe_key on these, so the same work can be queued again later.
func (s JobStatus) IsTerminal() bool {
	switch s {
	case JobCancelled, JobFailed, JobSucceeded:
		return true
	}
	return false
}

var (
	ErrJobTypeInvalid       = errors.New("job type must be dot-separated lowercase words, at most 64 characters")
	ErrJobPayloadInvalid    = errors.New("payload must be a JSON object")
	ErrJobDedupeKeyInvalid  = errors.New("dedupe key must be at most 255 printable ASCII characters without spaces")
	ErrJobUserInvalid       = errors.New("user id is invalid")
	ErrJobStatusInvalid     = errors.New("invalid job status")
	ErrJobAttemptsInvalid   = errors.New("attempts must be between 0 and max attempts, and max attempts at least 1")
	ErrJobAvailableRequired = errors.New("available time is required")
	ErrJobLockedByTooLong   = errors.New("locked by must be at most 128 characters")
	ErrJobLastErrorTooLong  = errors.New("last error must be at most 2048 characters")
	ErrJobImportRequired    = errors.New("import is required")
)

// Job is one unit of background work. Its business outcome lives elsewhere (e.g. imports).
type Job struct {
	ID          types.ID
	Type        JobType
	Payload     types.JSON
	DedupeKey   string
	UserID      types.ID
	Status      JobStatus
	Attempts    uint16
	MaxAttempts uint16
	AvailableAt time.Time
	LockedBy    string
	LastError   string
	FinishedAt  time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type JobParams struct {
	Type        JobType
	Payload     any    // marshaled to JSON; must encode as an object
	DedupeKey   string // empty means no dedup
	UserID      types.ID
	MaxAttempts uint16    // 0 means DefaultJobMaxAttempts
	AvailableAt time.Time // earliest run time; pass the clock's now to run immediately
}

var (
	jobTypePattern   = regexp.MustCompile(`^[a-z][a-z0-9_]*(\.[a-z][a-z0-9_]*)*$`)
	dedupeKeyPattern = regexp.MustCompile(`^[\x21-\x7E]+$`)
)

// NewJob validates input and returns a queued job with a fresh ID.
// On failure it returns validation.Errors listing every invalid field.
func NewJob(p JobParams) (*Job, error) {
	payload, err := json.Marshal(p.Payload)
	if err != nil {
		var errs validation.Errors
		errs.Add("payload", ErrJobPayloadInvalid)
		return nil, errs.Err()
	}

	j := &Job{
		ID:          types.NewID(),
		Type:        p.Type,
		Payload:     payload,
		DedupeKey:   strings.TrimSpace(p.DedupeKey),
		UserID:      p.UserID,
		Status:      JobQueued,
		MaxAttempts: p.MaxAttempts,
		// DATETIME(3) keeps milliseconds; truncate so the struct matches what is stored.
		AvailableAt: p.AvailableAt.UTC().Truncate(time.Millisecond),
	}

	if j.MaxAttempts == 0 {
		j.MaxAttempts = DefaultJobMaxAttempts
	}
	if err := j.Validate(); err != nil {
		return nil, err
	}

	return j, nil
}

// Validate mirrors the table's CHECK constraints and column limits.
func (j *Job) Validate() error {
	var errs validation.Errors

	if len(j.Type) > jobTypeMaxLen || !jobTypePattern.MatchString(string(j.Type)) {
		errs.Add("type", ErrJobTypeInvalid)
	}

	if !j.Payload.IsObject() {
		errs.Add("payload", ErrJobPayloadInvalid)
	}

	if j.DedupeKey != "" &&
		(len(j.DedupeKey) > jobDedupeKeyMaxLen || !dedupeKeyPattern.MatchString(j.DedupeKey)) {
		errs.Add("dedupeKey", ErrJobDedupeKeyInvalid)
	}

	if j.UserID.IsZero() {
		errs.Add("user", ErrJobUserInvalid)
	}

	if !j.Status.IsValid() {
		errs.Add("status", ErrJobStatusInvalid)
	}

	if j.MaxAttempts == 0 || j.Attempts > j.MaxAttempts {
		errs.Add("attempts", ErrJobAttemptsInvalid)
	}

	if j.AvailableAt.IsZero() {
		errs.Add("availableAt", ErrJobAvailableRequired)
	}

	if j.LockedBy != "" && utf8.RuneCountInString(j.LockedBy) > jobLockedByMaxLen {
		errs.Add("lockedBy", ErrJobLockedByTooLong)
	}

	if j.LastError != "" && utf8.RuneCountInString(j.LastError) > jobLastErrorMaxLen {
		errs.Add("lastError", ErrJobLastErrorTooLong)
	}

	return errs.Err()
}

// JobErrorText fits a handler's error into last_error. Error text is unbounded
// (wrapped driver errors, file content), so it is cut rather than rejected.
func JobErrorText(err error) string {
	return utils.TruncateRunes(err.Error(), jobLastErrorMaxLen)
}

// ImportJobPayload is the payload of an import.process job.
type ImportJobPayload struct {
	ImportID types.ID `json:"importId"`
}

// NewImportJob queues processing of imp. The dedupe key allows one active job per import.
func NewImportJob(imp *Import, now time.Time) (*Job, error) {
	if imp == nil {
		var errs validation.Errors
		errs.Add("import", ErrJobImportRequired)
		return nil, errs.Err()
	}
	userID := imp.UserID
	return NewJob(JobParams{
		Type:        JobImportProcess,
		Payload:     ImportJobPayload{ImportID: imp.ID},
		DedupeKey:   string(JobImportProcess) + ":" + imp.ID.String(),
		UserID:      userID,
		AvailableAt: now,
	})
}
