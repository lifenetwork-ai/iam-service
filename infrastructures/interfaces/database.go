package interfaces

import (
	"gorm.io/gorm"
)

type SQLClient interface {
	Connect() *gorm.DB
}
