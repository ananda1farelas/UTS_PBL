package model

import "time"

// Student represents a row in the students table.
type Student struct {
	ID           int        `json:"id"`
	UserID       int        `json:"user_id"`
	NIM          string     `json:"nim"`
	Nama         string     `json:"nama"`
	Prodi        string     `json:"prodi"`
	Angkatan     int        `json:"angkatan"`
	IPKTerakhir  float64    `json:"ipk_terakhir"`
	DeletedAt    *time.Time `json:"-"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`

	// Email is pulled in via a join with users for convenience in responses;
	// it is not a column on students itself.
	Email string `json:"email,omitempty"`
}

// StudentDetail is the richer payload returned by GET /students/{id}:
// the student plus their enrolled courses and SKS usage.
type StudentDetail struct {
	Student
	Courses  []Course `json:"courses"`
	TotalSKS int      `json:"total_sks"`
	BatasSKS int      `json:"batas_sks"`
}

// MaxSKS implements the business rule that caps how many SKS a student may
// take per semester, based on their last known IPK.
func (s Student) MaxSKS() int {
	switch {
	case s.IPKTerakhir >= 3.00:
		return 24
	case s.IPKTerakhir >= 2.50:
		return 21
	default:
		return 18
	}
}
