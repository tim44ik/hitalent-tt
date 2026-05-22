package unit

import (
	"context"
	"hitalent-test/internal/models"
	"hitalent-test/internal/services"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockCycleRepository struct {
	mock.Mock
}

func (m *MockCycleRepository) GetByID(ctx context.Context, id uint) (*models.Department, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Department), args.Error(1)
}

func (m *MockCycleRepository) Update(ctx context.Context, dept *models.Department) error {
	args := m.Called(ctx, dept)
	return args.Error(0)
}

func (m *MockCycleRepository) CheckNameUniqueness(ctx context.Context, parentID *uint, name string, excludeID uint) (bool, error) {
	args := m.Called(ctx, parentID, name, excludeID)
	return args.Bool(0), args.Error(1)
}

func (m *MockCycleRepository) IsDescendant(ctx context.Context, ancestorID, descendantID uint) (bool, error) {
	args := m.Called(ctx, ancestorID, descendantID)
	return args.Bool(0), args.Error(1)
}

func (m *MockCycleRepository) Create(ctx context.Context, dept *models.Department) error { return nil }
func (m *MockCycleRepository) Delete(ctx context.Context, id uint) error                 { return nil }
func (m *MockCycleRepository) GetChildrenIDs(ctx context.Context, parentID uint) ([]uint, error) {
	return nil, nil
}
func (m *MockCycleRepository) ReassignEmployees(ctx context.Context, fromDeptID, toDeptID uint) error {
	return nil
}
func (m *MockCycleRepository) ReassignChildren(ctx context.Context, oldParentID, newParentID uint) error {
	return nil
}
func (m *MockCycleRepository) GetSubtreeIDs(ctx context.Context, rootID uint) ([]uint, error) {
	return nil, nil
}
func (m *MockCycleRepository) GetDepartmentTree(ctx context.Context, id uint, depth int, includeEmployees bool) (*models.Department, error) {
	return nil, nil
}

func TestDepartmentService_Move_Cyclic(t *testing.T) {
	mockRepo := new(MockCycleRepository)
	svc := services.NewDepartmentService(mockRepo)

	ctx := context.Background()
	deptID := uint(1)
	newParentID := uint(2)

	existingDept := &models.Department{ID: deptID, Name: "IT", ParentID: nil}
	mockRepo.On("GetByID", ctx, deptID).Return(existingDept, nil)

	parentDept := &models.Department{ID: newParentID, Name: "Parent"}
	mockRepo.On("GetByID", ctx, newParentID).Return(parentDept, nil)

	mockRepo.On("IsDescendant", ctx, newParentID, deptID).Return(true, nil)

	updatedDept, err := svc.Update(ctx, deptID, nil, &newParentID)

	assert.Error(t, err)
	assert.EqualError(t, err, services.ErrCyclicMove)
	assert.Nil(t, updatedDept)

	mockRepo.AssertCalled(t, "IsDescendant", ctx, newParentID, deptID)
	mockRepo.AssertNotCalled(t, "CheckNameUniqueness")
	mockRepo.AssertNotCalled(t, "Update")
}

func TestDepartmentService_Move_NoCycle(t *testing.T) {
	mockRepo := new(MockCycleRepository)
	svc := services.NewDepartmentService(mockRepo)

	ctx := context.Background()
	deptID := uint(1)
	newParentID := uint(2)

	existingDept := &models.Department{ID: deptID, Name: "IT", ParentID: nil}
	mockRepo.On("GetByID", ctx, deptID).Return(existingDept, nil)
	mockRepo.On("GetByID", ctx, newParentID).Return(&models.Department{ID: newParentID}, nil)

	mockRepo.On("IsDescendant", ctx, newParentID, deptID).Return(false, nil)

	mockRepo.On("CheckNameUniqueness", ctx, &newParentID, existingDept.Name, deptID).Return(true, nil)
	mockRepo.On("Update", ctx, mock.MatchedBy(func(dept *models.Department) bool {
		return dept.ID == deptID && dept.ParentID != nil && *dept.ParentID == newParentID
	})).Return(nil)

	updatedDept, err := svc.Update(ctx, deptID, nil, &newParentID)

	assert.NoError(t, err)
	assert.NotNil(t, updatedDept)
	assert.Equal(t, newParentID, *updatedDept.ParentID)
	mockRepo.AssertCalled(t, "IsDescendant", ctx, newParentID, deptID)
	mockRepo.AssertCalled(t, "CheckNameUniqueness", ctx, &newParentID, existingDept.Name, deptID)
	mockRepo.AssertCalled(t, "Update", mock.Anything, mock.Anything)
}
