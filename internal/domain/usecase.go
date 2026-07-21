package domain

import "context"

// EmployeeUsecase is the inbound port exposed to the delivery (HTTP) layer.
// It encapsulates all business rules for employee management.
type EmployeeUsecase interface {
	Create(ctx context.Context, input CreateEmployeeInput) (*Employee, error)
	GetByID(ctx context.Context, id int64) (*Employee, error)
	Update(ctx context.Context, id int64, input UpdateEmployeeInput) (*Employee, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context) ([]*Employee, error)
}
