package services

import (
	"context"
	"errors"
	"hitalent-test/internal/models"
	"hitalent-test/internal/repositories"
	"strings"
)

type DepartmentService interface {
	Create(ctx context.Context, name string, parentID *uint) (*models.Department, error)
	GetByID(ctx context.Context, id uint, depth int, includeEmployees bool) (*models.Department, error)
	Update(ctx context.Context, id uint, name *string, parentID *uint) (*models.Department, error)
	Delete(ctx context.Context, id uint, mode string, reassignTo *uint) error
}

type departmentService struct {
	repo repositories.DepartmentRepository
}

func NewDepartmentService(repo repositories.DepartmentRepository) DepartmentService {
	return &departmentService{repo: repo}
}

func (s *departmentService) Create(ctx context.Context, name string, parentID *uint) (*models.Department, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errors.New(ErrEmptyName)
	}
	if len(name) > 200 {
		return nil, errors.New(ErrNameTooLong)
	}

	if parentID != nil && *parentID != 0 {
		parent, err := s.repo.GetByID(ctx, *parentID)
		if err != nil {
			return nil, err
		}
		if parent == nil {
			return nil, errors.New(ErrParentNotFound)
		}
	} else {
		parentID = nil
	}

	unique, err := s.repo.CheckNameUniqueness(ctx, parentID, name, 0)
	if err != nil {
		return nil, err
	}
	if !unique {
		return nil, errors.New(ErrDuplicateName)
	}

	dept := &models.Department{
		Name:     name,
		ParentID: parentID,
	}
	if err := s.repo.Create(ctx, dept); err != nil {
		return nil, err
	}
	return dept, nil
}

func (s *departmentService) GetByID(ctx context.Context, id uint, depth int, includeEmployees bool) (*models.Department, error) {
	dept, err := s.repo.GetDepartmentTree(ctx, id, depth, includeEmployees)
	if err != nil {
		return nil, err
	}
	if dept == nil {
		return nil, errors.New(ErrDepartmentNotFound)
	}
	return dept, nil
}

func (s *departmentService) Update(ctx context.Context, id uint, name *string, parentID *uint) (*models.Department, error) {
	dept, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if dept == nil {
		return nil, errors.New(ErrDepartmentNotFound)
	}

	if name != nil {
		newName := strings.TrimSpace(*name)
		if newName == "" {
			return nil, errors.New(ErrEmptyName)
		}
		if len(newName) > 200 {
			return nil, errors.New(ErrNameTooLong)
		}
		dept.Name = newName
	}

	if parentID != nil {
		if *parentID != 0 {
			newParent, err := s.repo.GetByID(ctx, *parentID)
			if err != nil {
				return nil, err
			}
			if newParent == nil {
				return nil, errors.New(ErrParentNotFound)
			}
			isDesc, err := s.repo.IsDescendant(ctx, *parentID, id)
			if err != nil {
				return nil, err
			}
			if isDesc {
				return nil, errors.New(ErrCyclicMove)
			}
			dept.ParentID = parentID
		} else {
			dept.ParentID = nil
		}
	}

	if name != nil || parentID != nil {
		unique, err := s.repo.CheckNameUniqueness(ctx, dept.ParentID, dept.Name, dept.ID)
		if err != nil {
			return nil, err
		}
		if !unique {
			return nil, errors.New(ErrDuplicateName)
		}
	}

	if err := s.repo.Update(ctx, dept); err != nil {
		return nil, err
	}
	return dept, nil
}

func (s *departmentService) Delete(ctx context.Context, id uint, mode string, reassignTo *uint) error {
	dept, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if dept == nil {
		return errors.New(ErrDepartmentNotFound)
	}

	if mode == "cascade" {
		return s.repo.Delete(ctx, id)
	} else if mode == "reassign" {
		if reassignTo == nil {
			return errors.New(ErrReassignToSelf)
		}
		if *reassignTo == id {
			return errors.New(ErrReassignToSelf)
		}
		isDesc, err := s.repo.IsDescendant(ctx, id, *reassignTo)
		if err != nil {
			return err
		}
		if isDesc {
			return errors.New(ErrReassignToDescendant)
		}
		subtreeIDs, err := s.repo.GetSubtreeIDs(ctx, id)
		if err != nil {
			return err
		}
		for _, subID := range subtreeIDs {
			if subID == id {
				continue
			}
			if err := s.repo.ReassignEmployees(ctx, subID, *reassignTo); err != nil {
				return err
			}
		}
		if err := s.repo.ReassignChildren(ctx, id, *reassignTo); err != nil {
			return err
		}
		if err := s.repo.Delete(ctx, id); err != nil {
			return err
		}
		return nil
	}
	return nil
}
