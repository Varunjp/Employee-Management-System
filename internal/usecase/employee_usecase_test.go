package usecase_test

import (
	"context"
	"employee_management/internal/domain"
	"employee_management/internal/usecase"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeRepository is an in-memory stand-in for domain.EmployeeRepository.
// Using a hand-rolled fake (rather than a mocking framework) keeps these
// tests fast, dependency-free, and focused purely on business behaviour.
type fakeRepository struct {
	employees map[int64]*domain.Employee
	nextID    int64
	failWith  error
}

func newFakeRepository() *fakeRepository {
	return &fakeRepository{employees: make(map[int64]*domain.Employee), nextID: 1}
}

func (f *fakeRepository) Create(_ context.Context, e *domain.Employee) (*domain.Employee, error) {
	if f.failWith != nil {
		return nil, f.failWith
	}
	e.ID = f.nextID
	e.CreatedAt = time.Now().UTC()
	f.employees[e.ID] = e
	f.nextID++
	return e, nil
}

func (f *fakeRepository) GetByID(_ context.Context, id int64) (*domain.Employee, error) {
	if e, ok := f.employees[id]; ok {
		return e, nil
	}
	return nil, domain.ErrEmployeeNotFound
}

func (f *fakeRepository) Update(_ context.Context, id int64, e *domain.Employee) (*domain.Employee, error) {
	existing, ok := f.employees[id]
	if !ok {
		return nil, domain.ErrEmployeeNotFound
	}
	existing.Name, existing.Position, existing.Salary, existing.HiredDate = e.Name, e.Position, e.Salary, e.HiredDate
	now := time.Now().UTC()
	existing.UpdatedAt = &now
	return existing, nil
}

func (f *fakeRepository) Delete(_ context.Context, id int64) error {
	if _, ok := f.employees[id]; !ok {
		return domain.ErrEmployeeNotFound
	}
	delete(f.employees, id)
	return nil
}

func (f *fakeRepository) List(_ context.Context) ([]*domain.Employee, error) {
	result := make([]*domain.Employee, 0, len(f.employees))
	for _, e := range f.employees {
		result = append(result, e)
	}
	return result, nil
}

// fakeCache is a no-op cache that always misses; it exercises the
// cache-aside code paths in the usecase without needing a real Redis.
type fakeCache struct {
	store map[string][]byte
}

func newFakeCache() *fakeCache {
	return &fakeCache{store: make(map[string][]byte)}
}

func (c *fakeCache) Get(_ context.Context, key string) ([]byte, bool, error) {
	v, ok := c.store[key]
	return v, ok, nil
}

func (c *fakeCache) Set(_ context.Context, key string, value []byte, _ time.Duration) error {
	c.store[key] = value
	return nil
}

func (c *fakeCache) Delete(_ context.Context, keys ...string) error {
	for _, k := range keys {
		delete(c.store, k)
	}
	return nil
}

func newTestUsecase() (domain.EmployeeUsecase, *fakeRepository, *fakeCache) {
	repo := newFakeRepository()
	cache := newFakeCache()
	return usecase.NewEmployeeUsecase(repo, cache), repo, cache
}

func TestCreate_Success(t *testing.T) {
	uc, _, _ := newTestUsecase()

	hired, _ := domain.ParseDate("2024-06-01")
	created, err := uc.Create(context.Background(), domain.CreateEmployeeInput{
		Name: "John Doe", Position: "Software Engineer", Salary: 60000, HiredDate: hired,
	})

	require.NoError(t, err)
	assert.Equal(t, int64(1), created.ID)
	assert.Equal(t, "John Doe", created.Name)
}

func TestCreate_ValidationErrors(t *testing.T) {
	uc, _, _ := newTestUsecase()
	hired, _ := domain.ParseDate("2024-06-01")

	testCases := []struct {
		name  string
		input domain.CreateEmployeeInput
	}{
		{"empty name", domain.CreateEmployeeInput{Name: "", Position: "Engineer", Salary: 100, HiredDate: hired}},
		{"empty position", domain.CreateEmployeeInput{Name: "Jane", Position: "", Salary: 100, HiredDate: hired}},
		{"negative salary", domain.CreateEmployeeInput{Name: "Jane", Position: "Engineer", Salary: -1, HiredDate: hired}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := uc.Create(context.Background(), tc.input)
			require.Error(t, err)
			assert.ErrorIs(t, err, domain.ErrInvalidInput)
		})
	}
}

func TestGetByID_NotFound(t *testing.T) {
	uc, _, _ := newTestUsecase()

	_, err := uc.GetByID(context.Background(), 999)

	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrEmployeeNotFound)
}

func TestGetByID_UsesCacheOnSecondCall(t *testing.T) {
	uc, repo, cache := newTestUsecase()
	hired, _ := domain.ParseDate("2024-06-01")
	created, err := uc.Create(context.Background(), domain.CreateEmployeeInput{
		Name: "Cached Employee", Position: "Engineer", Salary: 1000, HiredDate: hired,
	})
	require.NoError(t, err)

	// First read populates the cache.
	_, err = uc.GetByID(context.Background(), created.ID)
	require.NoError(t, err)
	assert.Len(t, cache.store, 1)

	// Removing the record directly from the repository proves the second
	// read is served from cache rather than hitting the repository again.
	delete(repo.employees, created.ID)

	fetched, err := uc.GetByID(context.Background(), created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.Name, fetched.Name)
}

func TestUpdate_InvalidatesCache(t *testing.T) {
	uc, _, cache := newTestUsecase()
	hired, _ := domain.ParseDate("2024-06-01")
	created, err := uc.Create(context.Background(), domain.CreateEmployeeInput{
		Name: "Old Name", Position: "Engineer", Salary: 1000, HiredDate: hired,
	})
	require.NoError(t, err)

	_, err = uc.GetByID(context.Background(), created.ID) // warm the cache
	require.NoError(t, err)
	require.Contains(t, cache.store, "employee:1")

	_, err = uc.Update(context.Background(), created.ID, domain.UpdateEmployeeInput{
		Name: "New Name", Position: "Senior Engineer", Salary: 2000, HiredDate: hired,
	})
	require.NoError(t, err)

	assert.NotContains(t, cache.store, "employee:1")
}

func TestDelete_NotFound(t *testing.T) {
	uc, _, _ := newTestUsecase()

	err := uc.Delete(context.Background(), 42)

	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrEmployeeNotFound)
}

func TestList_Empty(t *testing.T) {
	uc, _, _ := newTestUsecase()

	employees, err := uc.List(context.Background())

	require.NoError(t, err)
	assert.Empty(t, employees)
}
