CREATE TABLE IF NOT EXISTS enrollments (
    id SERIAL PRIMARY KEY,
    student_id INTEGER NOT NULL REFERENCES students (id) ON DELETE CASCADE,
    course_id INTEGER NOT NULL REFERENCES courses (id) ON DELETE CASCADE,
    tahun_akademik VARCHAR(20) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT uniq_enrollment UNIQUE (student_id, course_id, tahun_akademik)
);

CREATE INDEX IF NOT EXISTS idx_enrollments_student_id ON enrollments (student_id);
CREATE INDEX IF NOT EXISTS idx_enrollments_course_id ON enrollments (course_id);
