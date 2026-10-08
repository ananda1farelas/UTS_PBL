package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/ananda1farelas/latihan-fiber/UTS/app/model"
	"github.com/ananda1farelas/latihan-fiber/UTS/helper"
)

type StudentRepository struct {
	db *sql.DB
}

func NewStudentRepository(db *sql.DB) *StudentRepository {
	return &StudentRepository{db: db}
}

// StudentFilter holds the optional query filters supported by GET /students.
type StudentFilter struct {
	Prodi    string
	Angkatan string
	Search   string
	Sort     string // "nama" or "-ipk_terakhir"
}

func (r *StudentRepository) FindByUserID(userID int) (*model.Student, error) {
	var s model.Student
	err := r.db.QueryRow(
		`SELECT id, user_id, nim, nama, prodi, angkatan, ipk_terakhir, deleted_at, created_at, updated_at
		 FROM students WHERE user_id = $1 AND deleted_at IS NULL`,
		userID,
	).Scan(&s.ID, &s.UserID, &s.NIM, &s.Nama, &s.Prodi, &s.Angkatan, &s.IPKTerakhir, &s.DeletedAt, &s.CreatedAt, &s.UpdatedAt)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *StudentRepository) FindByID(id int) (*model.Student, error) {
	var s model.Student
	err := r.db.QueryRow(
		`SELECT id, user_id, nim, nama, prodi, angkatan, ipk_terakhir, deleted_at, created_at, updated_at
		 FROM students WHERE id = $1 AND deleted_at IS NULL`,
		id,
	).Scan(&s.ID, &s.UserID, &s.NIM, &s.Nama, &s.Prodi, &s.Angkatan, &s.IPKTerakhir, &s.DeletedAt, &s.CreatedAt, &s.UpdatedAt)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *StudentRepository) NimExists(nim string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(`SELECT EXISTS (SELECT 1 FROM students WHERE nim = $1)`, nim).Scan(&exists)
	return exists, err
}

func (r *StudentRepository) EmailExists(email string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(`SELECT EXISTS (SELECT 1 FROM users WHERE email = $1)`, email).Scan(&exists)
	return exists, err
}

// List returns a page of students matching the filter, plus the total count
// (ignoring pagination) for building the response meta.
func (r *StudentRepository) List(filter StudentFilter, p helper.PaginationParams) ([]model.Student, int, error) {
	where := []string{"deleted_at IS NULL"}
	args := []interface{}{}
	argPos := 1

	if filter.Prodi != "" {
		where = append(where, fmt.Sprintf("prodi = $%d", argPos))
		args = append(args, filter.Prodi)
		argPos++
	}
	if filter.Angkatan != "" {
		where = append(where, fmt.Sprintf("angkatan = $%d", argPos))
		args = append(args, filter.Angkatan)
		argPos++
	}
	if filter.Search != "" {
		where = append(where, fmt.Sprintf("(nim ILIKE $%d OR nama ILIKE $%d)", argPos, argPos))
		args = append(args, "%"+filter.Search+"%")
		argPos++
	}

	whereClause := strings.Join(where, " AND ")

	var total int
	countQuery := "SELECT COUNT(*) FROM students WHERE " + whereClause
	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	orderBy := "nama ASC"
	switch filter.Sort {
	case "nama":
		orderBy = "nama ASC"
	case "-ipk_terakhir":
		orderBy = "ipk_terakhir DESC"
	}

	query := fmt.Sprintf(
		`SELECT id, user_id, nim, nama, prodi, angkatan, ipk_terakhir, deleted_at, created_at, updated_at
		 FROM students WHERE %s ORDER BY %s LIMIT $%d OFFSET $%d`,
		whereClause, orderBy, argPos, argPos+1,
	)
	args = append(args, p.PerPage, p.Offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var students []model.Student
	for rows.Next() {
		var s model.Student
		if err := rows.Scan(&s.ID, &s.UserID, &s.NIM, &s.Nama, &s.Prodi, &s.Angkatan, &s.IPKTerakhir, &s.DeletedAt, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, 0, err
		}
		students = append(students, s)
	}

	return students, total, rows.Err()
}

// CreateWithUser creates the users row and the students row in one
// transaction, as required by the spec.
func (r *StudentRepository) CreateWithUser(userRepo *UserRepository, email, hashedPassword, nim, nama, prodi string, angkatan int, ipk float64) (*model.Student, error) {
	tx, err := userRepo.BeginTx()
	if err != nil {
		return nil, err
	}

	userID, err := userRepo.Create(tx, email, hashedPassword, model.RoleMahasiswa)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	var s model.Student
	err = tx.QueryRow(
		`INSERT INTO students (user_id, nim, nama, prodi, angkatan, ipk_terakhir)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id, user_id, nim, nama, prodi, angkatan, ipk_terakhir, deleted_at, created_at, updated_at`,
		userID, nim, nama, prodi, angkatan, ipk,
	).Scan(&s.ID, &s.UserID, &s.NIM, &s.Nama, &s.Prodi, &s.Angkatan, &s.IPKTerakhir, &s.DeletedAt, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &s, nil
}

// Update updates the mutable student fields (nim cannot be changed, per spec).
func (r *StudentRepository) Update(id int, nama, prodi string, angkatan int, ipk float64) (*model.Student, error) {
	var s model.Student
	err := r.db.QueryRow(
		`UPDATE students
		 SET nama = $1, prodi = $2, angkatan = $3, ipk_terakhir = $4, updated_at = now()
		 WHERE id = $5 AND deleted_at IS NULL
		 RETURNING id, user_id, nim, nama, prodi, angkatan, ipk_terakhir, deleted_at, created_at, updated_at`,
		nama, prodi, angkatan, ipk, id,
	).Scan(&s.ID, &s.UserID, &s.NIM, &s.Nama, &s.Prodi, &s.Angkatan, &s.IPKTerakhir, &s.DeletedAt, &s.CreatedAt, &s.UpdatedAt)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// SoftDelete marks a student (and its user account) as deleted without
// removing the rows, so history (enrollments) is preserved.
func (r *StudentRepository) SoftDelete(id int) error {
	result, err := r.db.Exec(
		`UPDATE students SET deleted_at = now(), updated_at = now() WHERE id = $1 AND deleted_at IS NULL`,
		id,
	)
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

// TotalSKS sums the SKS of every course the student is currently enrolled
// in, across all academic years.
func (r *StudentRepository) TotalSKS(studentID int) (int, error) {
	var total sql.NullInt64
	err := r.db.QueryRow(
		`SELECT COALESCE(SUM(c.sks), 0)
		 FROM enrollments e
		 JOIN courses c ON c.id = e.course_id
		 WHERE e.student_id = $1`,
		studentID,
	).Scan(&total)
	if err != nil {
		return 0, err
	}
	return int(total.Int64), nil
}

// EnrolledCourses returns the courses a student is currently taking.
func (r *StudentRepository) EnrolledCourses(studentID int) ([]model.Course, error) {
	rows, err := r.db.Query(
		`SELECT c.id, c.kode_mk, c.nama_mk, c.sks, c.semester, c.kuota, c.created_at, c.updated_at
		 FROM enrollments e
		 JOIN courses c ON c.id = e.course_id
		 WHERE e.student_id = $1
		 ORDER BY c.kode_mk`,
		studentID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var courses []model.Course
	for rows.Next() {
		var c model.Course
		if err := rows.Scan(&c.ID, &c.KodeMK, &c.NamaMK, &c.SKS, &c.Semester, &c.Kuota, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		courses = append(courses, c)
	}
	return courses, rows.Err()
}
