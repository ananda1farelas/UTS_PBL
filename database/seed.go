package database

import (
	"database/sql"
	"fmt"

	"github.com/ananda1farelas/latihan-fiber/UTS/helper"
)

var prodiList = []string{"Teknik Informatika", "Sistem Informasi", "Ilmu Komputer"}

// Seed populates the database with the minimum data required by the spec:
// 1 admin, 20 mahasiswa (with matching user accounts), and 10 mata kuliah.
// It is safe to run multiple times: every insert is skipped if the row
// already exists (checked by its unique key).
func Seed(db *sql.DB) error {
	if err := seedAdmin(db); err != nil {
		return fmt.Errorf("seed: admin: %w", err)
	}
	if err := seedStudents(db); err != nil {
		return fmt.Errorf("seed: students: %w", err)
	}
	if err := seedCourses(db); err != nil {
		return fmt.Errorf("seed: courses: %w", err)
	}
	return nil
}

func seedAdmin(db *sql.DB) error {
	const email = "admin@siakad.test"

	var exists bool
	if err := db.QueryRow(`SELECT EXISTS (SELECT 1 FROM users WHERE email = $1)`, email).Scan(&exists); err != nil {
		return err
	}
	if exists {
		return nil
	}

	hashed, err := helper.HashPassword("admin12345")
	if err != nil {
		return err
	}

	_, err = db.Exec(
		`INSERT INTO users (email, password, role) VALUES ($1, $2, $3)`,
		email, hashed, "admin",
	)
	return err
}

func seedStudents(db *sql.DB) error {
	for i := 1; i <= 20; i++ {
		nim := fmt.Sprintf("2022%08d", i) // 12 digit NIM
		email := fmt.Sprintf("mahasiswa%02d@siakad.test", i)
		nama := fmt.Sprintf("Mahasiswa %02d", i)
		prodi := prodiList[i%len(prodiList)]
		angkatan := 2020 + (i % 5)
		ipk := 2.00 + float64(i%21)/10.0 // spreads IPK across all three SKS brackets
		if ipk > 4.00 {
			ipk = 4.00
		}

		var exists bool
		if err := db.QueryRow(`SELECT EXISTS (SELECT 1 FROM students WHERE nim = $1)`, nim).Scan(&exists); err != nil {
			return err
		}
		if exists {
			continue
		}

		// Initial password is the NIM itself, hashed, per spec.
		hashed, err := helper.HashPassword(nim)
		if err != nil {
			return err
		}

		tx, err := db.Begin()
		if err != nil {
			return err
		}

		var userID int
		err = tx.QueryRow(
			`INSERT INTO users (email, password, role) VALUES ($1, $2, $3) RETURNING id`,
			email, hashed, "mahasiswa",
		).Scan(&userID)
		if err != nil {
			tx.Rollback()
			return err
		}

		_, err = tx.Exec(
			`INSERT INTO students (user_id, nim, nama, prodi, angkatan, ipk_terakhir)
			 VALUES ($1, $2, $3, $4, $5, $6)`,
			userID, nim, nama, prodi, angkatan, ipk,
		)
		if err != nil {
			tx.Rollback()
			return err
		}

		if err := tx.Commit(); err != nil {
			return err
		}
	}

	return nil
}

func seedCourses(db *sql.DB) error {
	courses := []struct {
		kodeMK   string
		namaMK   string
		sks      int
		semester int
		kuota    int
	}{
		{"IF101", "Algoritma dan Pemrograman", 3, 1, 40},
		{"IF102", "Matematika Diskrit", 3, 1, 40},
		{"IF201", "Struktur Data", 3, 2, 35},
		{"IF202", "Basis Data", 3, 2, 35},
		{"IF301", "Pemrograman Web", 3, 3, 30},
		{"IF302", "Jaringan Komputer", 3, 3, 30},
		{"IF401", "Rekayasa Perangkat Lunak", 3, 4, 25},
		{"IF402", "Kecerdasan Buatan", 3, 4, 25},
		{"IF501", "Keamanan Siber", 3, 5, 20},
		{"IF502", "Pengembangan Aplikasi Mobile", 3, 5, 20},
	}

	for _, c := range courses {
		var exists bool
		if err := db.QueryRow(`SELECT EXISTS (SELECT 1 FROM courses WHERE kode_mk = $1)`, c.kodeMK).Scan(&exists); err != nil {
			return err
		}
		if exists {
			continue
		}

		_, err := db.Exec(
			`INSERT INTO courses (kode_mk, nama_mk, sks, semester, kuota) VALUES ($1, $2, $3, $4, $5)`,
			c.kodeMK, c.namaMK, c.sks, c.semester, c.kuota,
		)
		if err != nil {
			return err
		}
	}

	return nil
}
