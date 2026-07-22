package usecase

import (
	"context"
	"employee_management/internal/domain"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	cacheKeyEmployeePrefix = "employee:"
	cacheKeyEmployeeList   = "employees:all"
	defaultCacheTTL        = 5 * time.Minute
)

// employeeUsecase implements domain.EmployeeUsecase. It depends only on the
// domain.EmployeeRepository and domain.Cache ports, never on a concrete
// database driver or cache client - this is what keeps the business layer
// independent and unit-testable in isolation.
type employeeUsecase struct {
	repo  domain.EmployeeRepository
	cache domain.Cache
}

// NewEmployeeUsecase wires a concrete repository and cache into the usecase.
func NewEmployeeUsecase(repo domain.EmployeeRepository, cache domain.Cache) domain.EmployeeUsecase {
	return &employeeUsecase{repo: repo, cache: cache}
}

func employeeCacheKey(id int64) string {
	return cacheKeyEmployeePrefix + strconv.FormatInt(id, 10)
}

func (u *employeeUsecase) Create(ctx context.Context, input domain.CreateEmployeeInput) (*domain.Employee, error) {
	if err := validateEmployeeFields(input.Name, input.Position, input.Salary, input.HiredDate); err != nil {
		return nil, err
	}

	employee := &domain.Employee{
		Name:      strings.TrimSpace(input.Name),
		Position:  strings.TrimSpace(input.Position),
		Salary:    input.Salary,
		HiredDate: input.HiredDate,
	}

	created, err := u.repo.Create(ctx, employee)
	if err != nil {
		return nil, fmt.Errorf("usecase: create employee: %w", err)
	}

	// Invalidate the list cache since the collection has changed. We do not
	// fail the request if cache invalidation fails - the cache is a
	// performance optimization, not a source of truth.
	_ = u.cache.Delete(ctx, cacheKeyEmployeeList)

	return created, nil
}

func (u *employeeUsecase) GetByID(ctx context.Context, id int64) (*domain.Employee, error) {
	if id <= 0 {
		return nil, domain.ErrInvalidInput
	}

	key := employeeCacheKey(id)
	if cached, ok, err := u.cache.Get(ctx, key); err == nil && ok {
		var employee domain.Employee
		if jsonErr := json.Unmarshal(cached, &employee); jsonErr == nil {
			return &employee, nil
		}
	}

	employee, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if payload, err := json.Marshal(employee); err == nil {
		_ = u.cache.Set(ctx, key, payload, defaultCacheTTL)
	}

	return employee, nil
}

func (u *employeeUsecase) Update(ctx context.Context, id int64, input domain.UpdateEmployeeInput) (*domain.Employee, error) {
	if id <= 0 {
		return nil, domain.ErrInvalidInput
	}
	if err := validateEmployeeFields(input.Name, input.Position, input.Salary, input.HiredDate); err != nil {
		return nil, err
	}

	employee := &domain.Employee{
		Name:      strings.TrimSpace(input.Name),
		Position:  strings.TrimSpace(input.Position),
		Salary:    input.Salary,
		HiredDate: input.HiredDate,
	}

	updated, err := u.repo.Update(ctx, id, employee)
	if err != nil {
		return nil, err
	}

	_ = u.cache.Delete(ctx, employeeCacheKey(id), cacheKeyEmployeeList)

	return updated, nil
}

func (u *employeeUsecase) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return domain.ErrInvalidInput
	}

	if err := u.repo.Delete(ctx, id); err != nil {
		return err
	}

	_ = u.cache.Delete(ctx, employeeCacheKey(id), cacheKeyEmployeeList)

	return nil
}

func (u *employeeUsecase) List(ctx context.Context) ([]*domain.Employee, error) {
	if cached, ok, err := u.cache.Get(ctx, cacheKeyEmployeeList); err == nil && ok {
		var employees []*domain.Employee
		if jsonErr := json.Unmarshal(cached, &employees); jsonErr == nil {
			return employees, nil
		}
	}

	employees, err := u.repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("usecase: list employees: %w", err)
	}

	if payload, err := json.Marshal(employees); err == nil {
		_ = u.cache.Set(ctx, cacheKeyEmployeeList, payload, defaultCacheTTL)
	}

	return employees, nil
}

// validateEmployeeFields centralizes the business validation rules shared by
// Create and Update so that the two code paths cannot drift apart.
func validateEmployeeFields(name, position string, salary int64, hiredDate time.Time) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("%w: name is required", domain.ErrInvalidInput)
	}
	if strings.TrimSpace(position) == "" {
		return fmt.Errorf("%w: position is required", domain.ErrInvalidInput)
	}
	if salary < 1 {
		return fmt.Errorf("%w: salary must not be negative", domain.ErrInvalidInput)
	}
	if hiredDate.IsZero() {
		return fmt.Errorf("%w: provide a proper hired date", domain.ErrInvalidInput)
	}
	return nil
}
