package repository

import (
	"database/sql"
	"errors"

	"github.com/ananda1farelas/latihan-fiber/UTS/app/model"
)

type EnrollmentRepository struct {
	db *sql.DB
}

func NewEnrollmentRepository(db *sql.DB) *EnrollmentRepository {
	return &EnrollmentRepository{db: db}
}

func (r *EnrollmentRepository) BeginTx() (*sql.Tx, error) {
	return r.db.Begin()
}

// ExistsForStudentCourseYear checks the duplicate-enrollment rule: a student
// cannot take the same course twice in the same academic year.
func (r *EnrollmentRepository) ExistsForStudentCourseYear(tx *sql.Tx, studentID, courseID int, tahunAkademik string) (bool, error) {
	var exists bool
	err := tx.QueryRow(
		`SELECT EXISTS (
			SELECT 1 FROM enrollments
			WHERE student_id = $1 AND course_id = $2 AND tahun_akademik = $3
		)`,
		studentID, courseID, tahunAkademik,
	).Scan(&exists)
	return exists, err
}

// CountForCourse returns how many students are currently enrolled in a
// course. Call it after LockForUpdate on the course row so the count is
// consistent under concurrent requests.
func (r *EnrollmentRepository) CountForCourse(tx *sql.Tx, courseID int) (int, error) {
	var count int
	err := tx.QueryRow(`SELECT COUNT(*) FROM enrollments WHERE course_id = $1`, courseID).Scan(&count)
	return count, err
}

// SumSKSForStudentYear sums the SKS of all courses the student has already
// taken in the given academic year, used to enforce the max-SKS rule.
func (r *EnrollmentRepository) SumSKSForStudentYear(tx *sql.Tx, studentID int, tahunAkademik string) (int, error) {
	var total sql.NullInt64
	err := tx.QueryRow(
		`SELECT COALESCE(SUM(c.sks), 0)
		 FROM enrollments e
		 JOIN courses c ON c.id = e.course_id
		 WHERE e.student_id = $1 AND e.tahun_akademik = $2`,
		studentID, tahunAkademik,
	).Scan(&total)
	return int(total.Int64), err
}

func (r *EnrollmentRepository) Create(tx *sql.Tx, studentID, courseID int, tahunAkademik string) (*model.Enrollment, error) {
	var e model.Enrollment
	err := tx.QueryRow(
		`INSERT INTO enrollments (student_id, course_id, tahun_akademik)
		 VALUES ($1, $2, $3)
		 RETURNING id, student_id, course_id, tahun_akademik, created_at`,
		studentID, courseID, tahunAkademik,
	).Scan(&e.ID, &e.StudentID, &e.CourseID, &e.TahunAkademik, &e.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *EnrollmentRepository) FindByID(id int) (*model.Enrollment, error) {
	var e model.Enrollment
	err := r.db.QueryRow(
		`SELECT id, student_id, course_id, tahun_akademik, created_at
		 FROM enrollments WHERE id = $1`,
		id,
	).Scan(&e.ID, &e.StudentID, &e.CourseID, &e.TahunAkademik, &e.CreatedAt)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *EnrollmentRepository) Delete(id int) error {
	result, err := r.db.Exec(`DELETE FROM enrollments WHERE id = $1`, id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}
