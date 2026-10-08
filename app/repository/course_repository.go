package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/ananda1farelas/latihan-fiber/UTS/app/model"
)

type CourseRepository struct {
	db *sql.DB
}

func NewCourseRepository(db *sql.DB) *CourseRepository {
	return &CourseRepository{db: db}
}

// CourseFilter holds the optional query filters supported by GET /courses.
type CourseFilter struct {
	Semester  string
	Search    string
	Available bool
}

// List returns courses with terisi/sisa_kuota computed from enrollments.
func (r *CourseRepository) List(filter CourseFilter) ([]model.Course, error) {
	where := []string{"1 = 1"}
	args := []interface{}{}
	argPos := 1

	if filter.Semester != "" {
		where = append(where, fmt.Sprintf("c.semester = $%d", argPos))
		args = append(args, filter.Semester)
		argPos++
	}
	if filter.Search != "" {
		where = append(where, fmt.Sprintf("(c.kode_mk ILIKE $%d OR c.nama_mk ILIKE $%d)", argPos, argPos))
		args = append(args, "%"+filter.Search+"%")
		argPos++
	}

	havingClause := ""
	if filter.Available {
		havingClause = "HAVING c.kuota - COUNT(e.id) > 0"
	}

	query := fmt.Sprintf(`
		SELECT c.id, c.kode_mk, c.nama_mk, c.sks, c.semester, c.kuota, c.created_at, c.updated_at,
		       COUNT(e.id) AS terisi
		FROM courses c
		LEFT JOIN enrollments e ON e.course_id = c.id
		WHERE %s
		GROUP BY c.id
		%s
		ORDER BY c.kode_mk
	`, strings.Join(where, " AND "), havingClause)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var courses []model.Course
	for rows.Next() {
		var c model.Course
		if err := rows.Scan(&c.ID, &c.KodeMK, &c.NamaMK, &c.SKS, &c.Semester, &c.Kuota, &c.CreatedAt, &c.UpdatedAt, &c.Terisi); err != nil {
			return nil, err
		}
		c.SisaKuota = c.Kuota - c.Terisi
		courses = append(courses, c)
	}

	return courses, rows.Err()
}

func (r *CourseRepository) FindByID(id int) (*model.Course, error) {
	var c model.Course
	err := r.db.QueryRow(
		`SELECT id, kode_mk, nama_mk, sks, semester, kuota, created_at, updated_at
		 FROM courses WHERE id = $1`,
		id,
	).Scan(&c.ID, &c.KodeMK, &c.NamaMK, &c.SKS, &c.Semester, &c.Kuota, &c.CreatedAt, &c.UpdatedAt)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// LockForUpdate locks the course row within tx so concurrent enrollment
// requests for the same course serialize on quota checks.
func (r *CourseRepository) LockForUpdate(tx *sql.Tx, id int) (*model.Course, error) {
	var c model.Course
	err := tx.QueryRow(
		`SELECT id, kode_mk, nama_mk, sks, semester, kuota, created_at, updated_at
		 FROM courses WHERE id = $1 FOR UPDATE`,
		id,
	).Scan(&c.ID, &c.KodeMK, &c.NamaMK, &c.SKS, &c.Semester, &c.Kuota, &c.CreatedAt, &c.UpdatedAt)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}
