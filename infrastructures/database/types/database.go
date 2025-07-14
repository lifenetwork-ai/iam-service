package types

import (
	"gorm.io/gorm"
)

type SQLClient interface {
	Connect() *gorm.DB
}
