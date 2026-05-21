package repositories

import (
	"context"
	"errors"
	"hitalent-test/internal/models"

	"gorm.io/gorm"
)

type DepartmentBasicRepository interface {
	Create(ctx context.Context, dept *models.Department) error
	GetByID(ctx context.Context, id uint) (*models.Department, error)
	Update(ctx context.Context, dept *models.Department) error
	Delete(ctx context.Context, id uint) error
}

type DepartmentUniquenessRepository interface {
	CheckNameUniqueness(ctx context.Context, parentID *uint, name string, excludeID uint) (bool, error)
}

type DepartmentCycleRepository interface {
	IsDescendant(ctx context.Context, ancestorID, descendantID uint) (bool, error)
}

type DepartmentTreeRepository interface {
	GetDepartmentTree(ctx context.Context, id uint, depth int, includeEmployees bool) (*models.Department, error)
	GetSubtreeIDs(ctx context.Context, rootID uint) ([]uint, error)
}

type DepartmentReassignRepository interface {
	ReassignEmployees(ctx context.Context, fromDeptID, toDeptID uint) error
	ReassignChildren(ctx context.Context, oldParentID, newParentID uint) error
}

type DepartmentRepository interface {
	DepartmentBasicRepository
	DepartmentUniquenessRepository
	DepartmentCycleRepository
	DepartmentTreeRepository
	DepartmentReassignRepository
}

type departmentRepository struct {
	db *gorm.DB
}

func NewDepartmentRepository(db *gorm.DB) DepartmentRepository {
	return &departmentRepository{db: db}
}

func (r *departmentRepository) Create(ctx context.Context, dept *models.Department) error {
	return r.db.WithContext(ctx).Create(dept).Error
}

func (r *departmentRepository) GetByID(ctx context.Context, id uint) (*models.Department, error) {
	var dept models.Department
	err := r.db.WithContext(ctx).First(&dept, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &dept, err
}

func (r *departmentRepository) Update(ctx context.Context, dept *models.Department) error {
	return r.db.WithContext(ctx).Save(dept).Error
}

func (r *departmentRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.Department{}, id).Error
}

func (r *departmentRepository) CheckNameUniqueness(ctx context.Context, parentID *uint, name string, excludeID uint) (bool, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&models.Department{}).Where("name = ? AND id != ?", name, excludeID)
	if parentID == nil {
		query = query.Where("parent_id IS NULL")
	} else {
		query = query.Where("parent_id = ?", *parentID)
	}
	err := query.Count(&count).Error
	return count == 0, err
}

func (r *departmentRepository) IsDescendant(ctx context.Context, ancestorID, descendantID uint) (bool, error) {
	if ancestorID == descendantID {
		return true, nil
	}
	var exists bool
	err := r.db.WithContext(ctx).Raw(`
		WITH RECURSIVE subtree AS (
			SELECT id FROM departments WHERE id = ?
			UNION ALL
			SELECT d.id FROM departments d
			INNER JOIN subtree s ON d.parent_id = s.id
		)
		SELECT EXISTS(SELECT 1 FROM subtree WHERE id = ?)
	`, ancestorID, descendantID).Scan(&exists).Error
	return exists, err
}

func (r *departmentRepository) ReassignEmployees(ctx context.Context, fromDeptID, toDeptID uint) error {
	return r.db.WithContext(ctx).Model(&models.Employee{}).
		Where("department_id = ?", fromDeptID).
		Update("department_id", toDeptID).Error
}

func (r *departmentRepository) ReassignChildren(ctx context.Context, oldParentID, newParentID uint) error {
	return r.db.WithContext(ctx).Model(&models.Department{}).
		Where("parent_id = ?", oldParentID).
		Update("parent_id", newParentID).Error
}

func (r *departmentRepository) GetSubtreeIDs(ctx context.Context, rootID uint) ([]uint, error) {
	var ids []uint
	err := r.db.WithContext(ctx).Raw(`
		WITH RECURSIVE subtree AS (
			SELECT id FROM departments WHERE id = ?
			UNION ALL
			SELECT d.id FROM departments d
			INNER JOIN subtree s ON d.parent_id = s.id
		)
		SELECT id FROM subtree
	`, rootID).Scan(&ids).Error
	return ids, err
}

func (r *departmentRepository) GetDepartmentTree(ctx context.Context, id uint, depth int, includeEmployees bool) (*models.Department, error) {
	var root models.Department
	if err := r.db.WithContext(ctx).First(&root, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	if err := r.loadChildren(ctx, &root, depth-1, includeEmployees); err != nil {
		return nil, err
	}
	if includeEmployees {
		var employees []models.Employee
		if err := r.db.WithContext(ctx).Where("department_id = ?", root.ID).Find(&employees).Error; err != nil {
			return nil, err
		}
		root.Employees = employees
	}
	return &root, nil
}

func (r *departmentRepository) loadChildren(ctx context.Context, dept *models.Department, remainingDepth int, includeEmployees bool) error {
	if remainingDepth < 0 {
		return nil
	}
	var children []models.Department
	if err := r.db.WithContext(ctx).Where("parent_id = ?", dept.ID).Find(&children).Error; err != nil {
		return err
	}
	for i := range children {
		if err := r.loadChildren(ctx, &children[i], remainingDepth-1, includeEmployees); err != nil {
			return err
		}
		if includeEmployees {
			var employees []models.Employee
			if err := r.db.WithContext(ctx).Where("department_id = ?", children[i].ID).Find(&employees).Error; err != nil {
				return err
			}
			children[i].Employees = employees
		}
	}
	dept.Children = children
	return nil
}
