package handler

import (
	"employee_management/internal/domain"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

// EmployeeHandler adapts HTTP requests/responses to the domain.EmployeeUsecase
// port. It contains no business logic - only request parsing, validation of
// wire-format concerns (e.g. date parsing), and response shaping.
type EmployeeHandler struct {
	usecase domain.EmployeeUsecase
}

func NewEmployeeHandler(usecase domain.EmployeeUsecase) *EmployeeHandler {
	return &EmployeeHandler{usecase: usecase}
}

// CreateEmployeeRequest is the wire-format request body for creating an employee.
type CreateEmployeeRequest struct {
	Name      string `json:"name" example:"John Doe"`
	Position  string `json:"position" example:"Software Engineer"`
	Salary    int64  `json:"salary" example:"60000"`
	HiredDate string `json:"hired_date" example:"2024-06-01"`
} //@name CreateEmployeeRequest

// UpdateEmployeeRequest is the wire-format request body for updating an employee.
type UpdateEmployeeRequest struct {
	Name      string `json:"name" example:"John Doe"`
	Position  string `json:"position" example:"Senior Software Engineer"`
	Salary    int64  `json:"salary" example:"80000"`
	HiredDate string `json:"hired_date" example:"2024-06-01"`
} //@name UpdateEmployeeRequest

func (h *EmployeeHandler) Create(c echo.Context) error {
	var req CreateEmployeeRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	hiredDate, err := domain.ParseDate(req.HiredDate)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "hired_date must be in YYYY-MM-DD format")
	}

	employee, err := h.usecase.Create(c.Request().Context(), domain.CreateEmployeeInput{
		Name:      req.Name,
		Position:  req.Position,
		Salary:    req.Salary,
		HiredDate: hiredDate,
	})
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, employee.ToResponse())
}

func (h *EmployeeHandler) GetByID(c echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}

	employee, err := h.usecase.GetByID(c.Request().Context(), id)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, employee.ToResponse())
}

func (h *EmployeeHandler) Update(c echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}

	var req UpdateEmployeeRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	hiredDate, err := domain.ParseDate(req.HiredDate)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "hired_date must be in YYYY-MM-DD format")
	}

	employee, err := h.usecase.Update(c.Request().Context(), id, domain.UpdateEmployeeInput{
		Name:      req.Name,
		Position:  req.Position,
		Salary:    req.Salary,
		HiredDate: hiredDate,
	})
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, employee.ToResponse())
}

func (h *EmployeeHandler) Delete(c echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}

	if err := h.usecase.Delete(c.Request().Context(), id); err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *EmployeeHandler) List(c echo.Context) error {
	employees, err := h.usecase.List(c.Request().Context())
	if err != nil {
		return err
	}

	responses := make([]domain.EmployeeResponse, 0, len(employees))
	for _, e := range employees {
		responses = append(responses, e.ToResponse())
	}

	return c.JSON(http.StatusOK, responses)
}

func parseID(c echo.Context) (int64, error) {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil || id <= 0 {
		return 0, echo.NewHTTPError(http.StatusBadRequest, "id must be a positive integer")
	}
	return id, nil
}
