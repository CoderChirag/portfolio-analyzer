package domain

type IDAndUserID struct {
	ID     ID
	UserID ID
}

type ImportIDUserAndStatus struct {
	UserID ID
	ID     ID
	Status ImportStatus
}

type JobIDUserAndStatus struct {
	UserID ID
	ID     ID
	Status JobStatus
}
