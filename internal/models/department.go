package models

import "time"

type Department struct {
	ID        uint         `gorm:"primaryKey" json:"id"`
	Name      string       `gorm:"size:200;not null;index:idx_parent_name,unique" json:"name"`
	ParentID  *uint        `gorm:"index" json:"parent_id,omitempty"`
	CreatedAt time.Time    `gorm:"autoCreateTime" json:"created_at"`
	Children  []Department `gorm:"-" json:"children,omitempty"`
	Employees []Employee   `gorm:"-" json:"employees,omitempty"`
}
