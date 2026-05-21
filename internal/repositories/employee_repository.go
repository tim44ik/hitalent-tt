package repositories

import (
	"context"
	"hitalent-test/internal/models"

	"gorm.io/gorm"
)

type EmployeeRepository interface {
	Create(ctx context.Context, emp *models.Employee) error
}

type employeeRepository struct {
	db *gorm.DB
}

func NewEmployeeRepository(db *gorm.DB) EmployeeRepository {
	return &employeeRepository{db: db}
}

func (r *employeeRepository) Create(ctx context.Context, emp *models.Employee) error {
	return r.db.WithContext(ctx).Create(emp).Error
}
