CREATE TABLE IF NOT EXISTS courses (
    id SERIAL PRIMARY KEY,
    kode_mk VARCHAR(20) NOT NULL UNIQUE,
    nama_mk VARCHAR(150) NOT NULL,
    sks INTEGER NOT NULL CHECK (sks > 0),
    semester INTEGER NOT NULL,
    kuota INTEGER NOT NULL CHECK (kuota >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_courses_kode_mk ON courses (kode_mk);
CREATE INDEX IF NOT EXISTS idx_courses_semester ON courses (semester);
