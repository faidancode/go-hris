package connection

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func ConnectGORMWithRetry(
	host, user, password, dbname, port, sslmode string,
	maxRetries int,
) (*gorm.DB, error) {

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		host, user, password, dbname, port, sslmode,
	)

	var lastErr error

	for i := 1; i <= maxRetries; i++ {

		db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			lastErr = err
			log.Printf("⚠️ GORM open failed (%d/%d): %v", i, maxRetries, err)
			time.Sleep(5 * time.Second)
			continue
		}

		sqlDB, err := db.DB()
		if err != nil {
			lastErr = err
			log.Printf("⚠️ get sql.DB failed (%d/%d): %v", i, maxRetries, err)
			time.Sleep(5 * time.Second)
			continue
		}

		if err := sqlDB.Ping(); err != nil {
			lastErr = err
			log.Printf("⚠️ DB ping failed (%d/%d): %v", i, maxRetries, err)
			time.Sleep(5 * time.Second)
			continue
		}

		// Pool config
		sqlDB.SetMaxOpenConns(25)
		sqlDB.SetMaxIdleConns(10)
		sqlDB.SetConnMaxLifetime(time.Hour)

		log.Println("✅ GORM connected to database")
		return db, nil
	}

	return nil, fmt.Errorf("database connection failed after %d retries: %w", maxRetries, lastErr)
}

func ConnectRedisWithRetry(addr string, maxRetries int) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	for i := 1; i <= maxRetries; i++ {
		ctx := context.Background()
		if err := rdb.Ping(ctx).Err(); err == nil {
			log.Println("✅ Connected to Redis")
			return rdb, nil
		}

		log.Printf("⚠️ Redis retry %d/%d failed", i, maxRetries)
		time.Sleep(5 * time.Second)
	}

	return nil, fmt.Errorf("failed to connect redis")
}
