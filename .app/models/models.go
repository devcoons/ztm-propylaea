package models

import (
	"fmt"

	"gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB) {
	if db != nil {
		fmt.Println("[MDL] Automigrate models does not have any models")
	} else {
		fmt.Println("[MDL] Could not migrate models (db missing)")
	}
}
