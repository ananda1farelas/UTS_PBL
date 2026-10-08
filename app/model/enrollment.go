package model

import "time"

// Enrollment represents a row in the enrollments table (one KRS entry).
type Enrollment struct {
	ID            int       `json:"id"`
	StudentID     int       `json:"student_id"`
	CourseID      int       `json:"course_id"`
	TahunAkademik string    `json:"tahun_akademik"`
	CreatedAt     time.Time `json:"created_at"`

	// Course is populated on reads that join enrollments with courses.
	Course *Course `json:"course,omitempty"`
}
