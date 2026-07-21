// Package postgres implements domain.EmployeeRepository against a real
// PostgreSQL database using the pgx driver directly. Per project
// requirements, no ORM is used - every query below is raw SQL.
package postgres

import (
	"context"
	"employee_management/internal/domain"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type employeeRepository struct {
	pool *pgxpool.Pool
}

// NewEmployeeRepository returns a domain.EmployeeRepository backed by Postgres.
func NewEmployeeRepository(pool *pgxpool.Pool) domain.EmployeeRepository {
	return &employeeRepository{pool: pool}
}

const createEmployeeQuery = `
	INSERT INTO employees (name, position, salary, hired_date, created_at)
	VALUES ($1, $2, $3, $4, now())
	RETURNING id, name, position, salary, hired_date, created_at, updated_at`

func (r *employeeRepository) Create(ctx context.Context, e *domain.Employee) (*domain.Employee, error) {
	row := r.pool.QueryRow(ctx, createEmployeeQuery, e.Name, e.Position, e.Salary, e.HiredDate)
	result, err := scanEmployee(row)
	if err != nil {
		return nil, fmt.Errorf("postgres: create employee: %w", err)
	}
	return result, nil
}

const getEmployeeByIDQuery = `
	SELECT id, name, position, salary, hired_date, created_at, updated_at
	FROM employees
	WHERE id = $1`

func (r *employeeRepository) GetByID(ctx context.Context, id int64) (*domain.Employee, error) {
	row := r.pool.QueryRow(ctx, getEmployeeByIDQuery, id)
	result, err := scanEmployee(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrEmployeeNotFound
		}
		return nil, fmt.Errorf("postgres: get employee by id: %w", err)
	}
	return result, nil
}

const updateEmployeeQuery = `
	UPDATE employees
	SET name = $1, position = $2, salary = $3, hired_date = $4, updated_at = now()
	WHERE id = $5
	RETURNING id, name, position, salary, hired_date, created_at, updated_at`

func (r *employeeRepository) Update(ctx context.Context, id int64, e *domain.Employee) (*domain.Employee, error) {
	row := r.pool.QueryRow(ctx, updateEmployeeQuery, e.Name, e.Position, e.Salary, e.HiredDate, id)
	result, err := scanEmployee(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrEmployeeNotFound
		}
		return nil, fmt.Errorf("postgres: update employee: %w", err)
	}
	return result, nil
}

const deleteEmployeeQuery = `DELETE FROM employees WHERE id = $1`

func (r *employeeRepository) Delete(ctx context.Context, id int64) error {
	tag, err := r.pool.Exec(ctx, deleteEmployeeQuery, id)
	if err != nil {
		return fmt.Errorf("postgres: delete employee: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrEmployeeNotFound
	}
	return nil
}

const listEmployeesQuery = `
	SELECT id, name, position, salary, hired_date, created_at, updated_at
	FROM employees
	ORDER BY id ASC`

func (r *employeeRepository) List(ctx context.Context) ([]*domain.Employee, error) {
	rows, err := r.pool.Query(ctx, listEmployeesQuery)
	if err != nil {
		return nil, fmt.Errorf("postgres: list employees: %w", err)
	}
	defer rows.Close()

	employees := make([]*domain.Employee, 0)
	for rows.Next() {
		e, err := scanEmployeeRows(rows)
		if err != nil {
			return nil, fmt.Errorf("postgres: scan employee row: %w", err)
		}
		employees = append(employees, e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("postgres: list employees rows: %w", err)
	}

	return employees, nil
}

// rowScanner abstracts over pgx.Row and pgx.Rows, both of which implement
// Scan(...) error, so scanEmployee can be shared by single-row and
// multi-row query paths without duplicating the field list.
type rowScanner interface {
	Scan(dest ...any) error
}

func scanEmployee(row rowScanner) (*domain.Employee, error) {
	return scanEmployeeRows(row)
}

func scanEmployeeRows(row rowScanner) (*domain.Employee, error) {
	var e domain.Employee
	if err := row.Scan(&e.ID, &e.Name, &e.Position, &e.Salary, &e.HiredDate, &e.CreatedAt, &e.UpdatedAt); err != nil {
		return nil, err
	}
	return &e, nil
}
