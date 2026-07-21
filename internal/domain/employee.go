package domain

import "time"

// Employee is the core business entity for an employee record.
// It has no dependency on any framework, database driver, or transport
// mechanism, in accordance with the Dependency Rule of Clean Architecture.
type Employee struct {
	ID        int64
	Name      string
	Position  string
	Salary    int64
	HiredDate time.Time
	CreatedAt time.Time
	UpdatedAt *time.Time
}

// dateLayout is the canonical layout used to marshal/unmarshal dates
// that only carry day-precision (no time-of-day component).
const dateLayout = "2006-01-02"

// CreateEmployeeInput carries validated data required to create an employee.
// It is produced by the delivery layer and consumed by the usecase layer.
type CreateEmployeeInput struct {
	Name      string
	Position  string
	Salary    int64
	HiredDate time.Time
}

// UpdateEmployeeInput carries validated data required to update an employee.
type UpdateEmployeeInput struct {
	Name      string
	Position  string
	Salary    int64
	HiredDate time.Time
}

// EmployeeResponse is the JSON representation returned to API clients.
type EmployeeResponse struct {
	ID        int64   `json:"id"`
	Name      string  `json:"name"`
	Position  string  `json:"position"`
	Salary    int64   `json:"salary"`
	HiredDate string  `json:"hired_date"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt *string `json:"updated_at,omitempty"`
}

// ToResponse converts the domain entity into a transport-friendly DTO.
func (e *Employee) ToResponse() EmployeeResponse {
	resp := EmployeeResponse{
		ID:        e.ID,
		Name:      e.Name,
		Position:  e.Position,
		Salary:    e.Salary,
		HiredDate: e.HiredDate.Format(dateLayout),
		CreatedAt: e.CreatedAt.UTC().Format(time.RFC3339),
	}
	if e.UpdatedAt != nil {
		updated := e.UpdatedAt.UTC().Format(time.RFC3339)
		resp.UpdatedAt = &updated
	}
	return resp
}

// ParseDate parses a "YYYY-MM-DD" string into a time.Time (UTC, midnight).
func ParseDate(value string) (time.Time, error) {
	return time.Parse(dateLayout, value)
}
