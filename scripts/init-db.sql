CREATE TABLE IF NOT EXISTS employees (
    id          BIGSERIAL PRIMARY KEY,
    name        VARCHAR(255)   NOT NULL,
    position    VARCHAR(255)   NOT NULL,
    salary      INTEGER        NOT NULL CHECK (salary >= 0),
    hired_date  DATE           NOT NULL,
    created_at  TIMESTAMPTZ    NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_employees_position ON employees (position);
