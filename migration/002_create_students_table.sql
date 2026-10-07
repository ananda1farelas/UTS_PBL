CREATE TABLE IF NOT EXISTS students (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL UNIQUE REFERENCES users (id) ON DELETE CASCADE,
    nim VARCHAR(12) NOT NULL UNIQUE,
    nama VARCHAR(150) NOT NULL,
    prodi VARCHAR(100) NOT NULL,
    angkatan INTEGER NOT NULL,
    ipk_terakhir NUMERIC(3, 2) DEFAULT 0.00 CHECK (ipk_terakhir >= 0.00 AND ipk_terakhir <= 4.00),
    deleted_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_students_nim ON students (nim);
CREATE INDEX IF NOT EXISTS idx_students_nama ON students (nama);
CREATE INDEX IF NOT EXISTS idx_students_deleted_at ON students (deleted_at);
