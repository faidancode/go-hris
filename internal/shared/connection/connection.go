package connection

import (
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func ConnectGORMWithRetry(dsn string, maxRetries int) (*gorm.DB, error) {
	var db *gorm.DB
	var err error

	for i := 1; i <= maxRetries; i++ {
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})

		if err == nil {
			// Validasi koneksi ke DB (sama seperti Ping)
			sqlDB, _ := db.DB()
			if err = sqlDB.Ping(); err == nil {
				log.Println("✅ GORM connected to database")
				return db, nil
			}
		}

		log.Printf("⚠️ DB retry %d/%d failed: %v", i, maxRetries, err)
		if i < maxRetries {
			time.Sleep(5 * time.Second)
		}
	}

	return nil, err
}
