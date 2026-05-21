package services

import (
	"context"
	"errors"
	"strings"
	"time"

	"hitalent-test/internal/models"
)

type EmployeeRepository interface {
	Create(ctx context.Context, emp *models.Employee) error
}

type DepartmentRepositoryForEmployee interface {
	GetByID(ctx context.Context, id uint) (*models.Department, error)
}

type EmployeeService interface {
	Create(ctx context.Context, departmentID uint, fullName, position string, hiredAt *time.Time) (*models.Employee, error)
}

type employeeService struct {
	empRepo  EmployeeRepository
	deptRepo DepartmentRepositoryForEmployee
}

func NewEmployeeService(empRepo EmployeeRepository, deptRepo DepartmentRepositoryForEmployee) EmployeeService {
	return &employeeService{
		empRepo:  empRepo,
		deptRepo: deptRepo,
	}
}

func (s *employeeService) Create(ctx context.Context, departmentID uint, fullName, position string, hiredAt *time.Time) (*models.Employee, error) {
	dept, err := s.deptRepo.GetByID(ctx, departmentID)
	if err != nil {
		return nil, err
	}
	if dept == nil {
		return nil, errors.New(ErrInvalidDepartment)
	}

	fullName = strings.TrimSpace(fullName)
	if fullName == "" {
		return nil, errors.New(ErrEmptyFullName)
	}
	if len(fullName) > 200 {
		return nil, errors.New(ErrFullNameTooLong)
	}

	position = strings.TrimSpace(position)
	if position == "" {
		return nil, errors.New(ErrEmptyPosition)
	}
	if len(position) > 200 {
		return nil, errors.New(ErrPositionTooLong)
	}

	emp := &models.Employee{
		DepartmentID: departmentID,
		FullName:     fullName,
		Position:     position,
		HiredAt:      hiredAt,
	}

	if err := s.empRepo.Create(ctx, emp); err != nil {
		return nil, err
	}
	return emp, nil
}
