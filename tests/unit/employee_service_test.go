package unit

import (
	"context"
	"testing"
	"time"

	"hitalent-test/internal/models"
	"hitalent-test/internal/services"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockEmployeeRepository struct {
	mock.Mock
}

func (m *MockEmployeeRepository) Create(ctx context.Context, emp *models.Employee) error {
	args := m.Called(ctx, emp)
	return args.Error(0)
}

type MockDepartmentRepoForEmployee struct {
	mock.Mock
}

func (m *MockDepartmentRepoForEmployee) GetByID(ctx context.Context, id uint) (*models.Department, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Department), args.Error(1)
}

func TestEmployeeService_Create_Success(t *testing.T) {
	mockEmpRepo := new(MockEmployeeRepository)
	mockDeptRepo := new(MockDepartmentRepoForEmployee)
	svc := services.NewEmployeeService(mockEmpRepo, mockDeptRepo)

	deptID := uint(1)
	fullName := "John Doe"
	position := "Developer"
	hiredAt := time.Now()

	mockDeptRepo.On("GetByID", mock.Anything, deptID).Return(&models.Department{ID: deptID}, nil)
	mockEmpRepo.On("Create", mock.Anything, mock.MatchedBy(func(emp *models.Employee) bool {
		return emp.DepartmentID == deptID && emp.FullName == fullName && emp.Position == position && emp.HiredAt == &hiredAt
	})).Return(nil)

	emp, err := svc.Create(context.Background(), deptID, fullName, position, &hiredAt)

	assert.NoError(t, err)
	assert.NotNil(t, emp)
	assert.Equal(t, fullName, emp.FullName)
	assert.Equal(t, position, emp.Position)
	assert.Equal(t, deptID, emp.DepartmentID)
	assert.Equal(t, hiredAt, *emp.HiredAt)
	mockDeptRepo.AssertExpectations(t)
	mockEmpRepo.AssertExpectations(t)
}

func TestEmployeeService_Create_DepartmentNotFound(t *testing.T) {
	mockEmpRepo := new(MockEmployeeRepository)
	mockDeptRepo := new(MockDepartmentRepoForEmployee)
	svc := services.NewEmployeeService(mockEmpRepo, mockDeptRepo)

	deptID := uint(999)
	mockDeptRepo.On("GetByID", mock.Anything, deptID).Return(nil, nil)

	emp, err := svc.Create(context.Background(), deptID, "Jane", "QA", nil)

	assert.Error(t, err)
	assert.EqualError(t, err, services.ErrInvalidDepartment)
	assert.Nil(t, emp)
	mockDeptRepo.AssertExpectations(t)
	mockEmpRepo.AssertNotCalled(t, "Create")
}

func TestEmployeeService_Create_EmptyFullName(t *testing.T) {
	mockEmpRepo := new(MockEmployeeRepository)
	mockDeptRepo := new(MockDepartmentRepoForEmployee)
	svc := services.NewEmployeeService(mockEmpRepo, mockDeptRepo)

	deptID := uint(1)
	mockDeptRepo.On("GetByID", mock.Anything, deptID).Return(&models.Department{ID: deptID}, nil)

	emp, err := svc.Create(context.Background(), deptID, "   ", "QA", nil)

	assert.Error(t, err)
	assert.EqualError(t, err, services.ErrEmptyFullName)
	assert.Nil(t, emp)
}

func TestEmployeeService_Create_FullNameTooLong(t *testing.T) {
	mockEmpRepo := new(MockEmployeeRepository)
	mockDeptRepo := new(MockDepartmentRepoForEmployee)
	svc := services.NewEmployeeService(mockEmpRepo, mockDeptRepo)

	deptID := uint(1)
	longName := string(make([]byte, 201))
	mockDeptRepo.On("GetByID", mock.Anything, deptID).Return(&models.Department{ID: deptID}, nil)

	emp, err := svc.Create(context.Background(), deptID, longName, "QA", nil)

	assert.Error(t, err)
	assert.EqualError(t, err, services.ErrFullNameTooLong)
	assert.Nil(t, emp)
}

func TestEmployeeService_Create_EmptyPosition(t *testing.T) {
	mockEmpRepo := new(MockEmployeeRepository)
	mockDeptRepo := new(MockDepartmentRepoForEmployee)
	svc := services.NewEmployeeService(mockEmpRepo, mockDeptRepo)

	deptID := uint(1)
	mockDeptRepo.On("GetByID", mock.Anything, deptID).Return(&models.Department{ID: deptID}, nil)

	emp, err := svc.Create(context.Background(), deptID, "John Doe", "   ", nil)

	assert.Error(t, err)
	assert.EqualError(t, err, services.ErrEmptyPosition)
	assert.Nil(t, emp)
}

func TestEmployeeService_Create_PositionTooLong(t *testing.T) {
	mockEmpRepo := new(MockEmployeeRepository)
	mockDeptRepo := new(MockDepartmentRepoForEmployee)
	svc := services.NewEmployeeService(mockEmpRepo, mockDeptRepo)

	deptID := uint(1)
	longPos := string(make([]byte, 201))
	mockDeptRepo.On("GetByID", mock.Anything, deptID).Return(&models.Department{ID: deptID}, nil)

	emp, err := svc.Create(context.Background(), deptID, "John Doe", longPos, nil)

	assert.Error(t, err)
	assert.EqualError(t, err, services.ErrPositionTooLong)
	assert.Nil(t, emp)
}
