package unit

import (
	"context"
	"hitalent-test/internal/models"
	"hitalent-test/internal/services"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockDepartmentRepository struct {
	mock.Mock
}

func (m *MockDepartmentRepository) Create(ctx context.Context, dept *models.Department) error {
	args := m.Called(ctx, dept)
	return args.Error(0)
}

func (m *MockDepartmentRepository) GetByID(ctx context.Context, id uint) (*models.Department, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Department), args.Error(1)
}

func (m *MockDepartmentRepository) Update(ctx context.Context, dept *models.Department) error {
	args := m.Called(ctx, dept)
	return args.Error(0)
}

func (m *MockDepartmentRepository) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockDepartmentRepository) CheckNameUniqueness(ctx context.Context, parentID *uint, name string, excludeID uint) (bool, error) {
	args := m.Called(ctx, parentID, name, excludeID)
	return args.Bool(0), args.Error(1)
}

func (m *MockDepartmentRepository) IsDescendant(ctx context.Context, ancestorID, descendantID uint) (bool, error) {
	args := m.Called(ctx, ancestorID, descendantID)
	return args.Bool(0), args.Error(1)
}

func (m *MockDepartmentRepository) GetChildrenIDs(ctx context.Context, parentID uint) ([]uint, error) {
	args := m.Called(ctx, parentID)
	return args.Get(0).([]uint), args.Error(1)
}

func (m *MockDepartmentRepository) ReassignEmployees(ctx context.Context, fromDeptID, toDeptID uint) error {
	args := m.Called(ctx, fromDeptID, toDeptID)
	return args.Error(0)
}

func (m *MockDepartmentRepository) ReassignChildren(ctx context.Context, oldParentID, newParentID uint) error {
	args := m.Called(ctx, oldParentID, newParentID)
	return args.Error(0)
}

func (m *MockDepartmentRepository) GetSubtreeIDs(ctx context.Context, rootID uint) ([]uint, error) {
	args := m.Called(ctx, rootID)
	return args.Get(0).([]uint), args.Error(1)
}

func (m *MockDepartmentRepository) GetDepartmentTree(ctx context.Context, id uint, depth int, includeEmployees bool) (*models.Department, error) {
	args := m.Called(ctx, id, depth, includeEmployees)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Department), args.Error(1)
}

func TestDepartmentService_Create_Success(t *testing.T) {
	mockRepo := new(MockDepartmentRepository)
	svc := services.NewDepartmentService(mockRepo)

	parentID := uint(1)
	name := "Test Department"

	mockRepo.On("GetByID", mock.Anything, parentID).Return(&models.Department{ID: parentID}, nil)
	mockRepo.On("CheckNameUniqueness", mock.Anything, &parentID, name, uint(0)).Return(true, nil)
	mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(dept *models.Department) bool {
		return dept.Name == name && *dept.ParentID == parentID
	})).Return(nil)

	dept, err := svc.Create(context.Background(), name, &parentID)

	assert.NoError(t, err)
	assert.NotNil(t, dept)
	assert.Equal(t, name, dept.Name)
	assert.Equal(t, parentID, *dept.ParentID)
	mockRepo.AssertExpectations(t)
}

func TestDepartmentService_Create_EmptyName(t *testing.T) {
	mockRepo := new(MockDepartmentRepository)
	svc := services.NewDepartmentService(mockRepo)

	dept, err := svc.Create(context.Background(), "   ", nil)

	assert.Error(t, err)
	assert.EqualError(t, err, services.ErrEmptyName)
	assert.Nil(t, dept)
}

func TestDepartmentService_Create_DuplicateName(t *testing.T) {
	mockRepo := new(MockDepartmentRepository)
	svc := services.NewDepartmentService(mockRepo)

	parentID := uint(1)
	name := "Duplicate"

	mockRepo.On("GetByID", mock.Anything, parentID).Return(&models.Department{ID: parentID}, nil)
	mockRepo.On("CheckNameUniqueness", mock.Anything, &parentID, name, uint(0)).Return(false, nil)

	dept, err := svc.Create(context.Background(), name, &parentID)

	assert.Error(t, err)
	assert.EqualError(t, err, services.ErrDuplicateName)
	assert.Nil(t, dept)
	mockRepo.AssertExpectations(t)
}

func TestDepartmentService_Create_ParentNotFound(t *testing.T) {
	mockRepo := new(MockDepartmentRepository)
	svc := services.NewDepartmentService(mockRepo)

	parentID := uint(999)
	name := "Orphan"

	mockRepo.On("GetByID", mock.Anything, parentID).Return(nil, nil)

	dept, err := svc.Create(context.Background(), name, &parentID)

	assert.Error(t, err)
	assert.EqualError(t, err, services.ErrParentNotFound)
	assert.Nil(t, dept)
	mockRepo.AssertExpectations(t)
}
