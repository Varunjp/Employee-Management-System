package domain

import "context"

// EmployeeRepository is the outbound port the usecase layer depends on to
// persist and retrieve employees. Concrete implementations (e.g. Postgres)
// live in internal/repository and are wired in at startup via main.go,
// so the usecase layer never depends on a concrete database driver.
type EmployeeRepository interface {
	Create(ctx context.Context, e *Employee) (*Employee, error)
	GetByID(ctx context.Context, id int64) (*Employee, error)
	Update(ctx context.Context, id int64, e *Employee) (*Employee, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context) ([]*Employee, error)
}
