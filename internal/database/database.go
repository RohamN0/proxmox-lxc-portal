package database

import (
	"proxmox-lxc-portal/internal/models"
	"sync"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var (
	DB   *gorm.DB
	once sync.Once
)

func ConnectDB() {
	once.Do(func() {
		db, err := gorm.Open(sqlite.Open("hosting.db"), &gorm.Config{})
		if err != nil {
			panic("failed to connect database")
		}

		err = db.AutoMigrate(&models.User{}, &models.RefreshToken{})
		if err != nil {
			panic("failed to migrate database")
		}

		DB = db
	})

}
