package department

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/singleflight"
)

const (
	// Prefix untuk Department
	DepartmentAllKeyPrefix = "departments:all:"
	DepartmentDetailPrefix = "departments:detail:"
)

// Helper untuk mendapatkan key lengkap
func GetDepartmentAllKey(companyID string) string {
	return DepartmentAllKeyPrefix + companyID
}

//go:generate mockgen -source=department_service.go -destination=mock/department_service_mock.go -package=mock
type Service interface {
	Create(ctx context.Context, companyID string, req CreateDepartmentRequest) (DepartmentResponse, error)
	GetAll(ctx context.Context, companyID string) ([]DepartmentResponse, error)
	GetByID(ctx context.Context, companyID, id string) (DepartmentResponse, error)
	Update(ctx context.Context, companyID, id string, req UpdateDepartmentRequest) (DepartmentResponse, error)
	Delete(ctx context.Context, companyID, id string) error
}

type service struct {
	db   *sql.DB
	repo Repository
	rdb  *redis.Client
	sf   *singleflight.Group
}

func NewService(db *sql.DB, repo Repository, rdb *redis.Client) Service {
	return &service{db: db, repo: repo, rdb: rdb, sf: &singleflight.Group{}}
}

func (s *service) Create(
	ctx context.Context,
	companyID string,
	req CreateDepartmentRequest,
) (DepartmentResponse, error) {

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return DepartmentResponse{}, err
	}
	defer tx.Rollback()

	qtx := s.repo.WithTx(tx)

	dept := &Department{
		ID:        uuid.New(),
		Name:      req.Name,
		CompanyID: uuid.MustParse(companyID),
	}

	if err := qtx.Create(ctx, dept); err != nil {
		return DepartmentResponse{}, err
	}

	if err := tx.Commit(); err != nil {
		return DepartmentResponse{}, err
	}

	if s.rdb != nil {
		cacheKey := GetDepartmentAllKey(companyID)
		if err := s.rdb.Del(ctx, cacheKey).Err(); err != nil {
			log.Printf("ERROR: failed to invalidate cache for key %s: %v", cacheKey, err)
		}
	}

	return mapToResponse(*dept), nil
}

func (s *service) GetAll(
	ctx context.Context,
	companyID string,
) ([]DepartmentResponse, error) {
	// Definisikan key cache yang unik per company (konsisten dengan invalidation)
	cacheKey := GetDepartmentAllKey(companyID)

	// Coba ambil dari Redis
	if s.rdb != nil {
		cached, err := s.rdb.Get(ctx, cacheKey).Result()
		if err == nil {
			var resp []DepartmentResponse
			if err := json.Unmarshal([]byte(cached), &resp); err == nil {
				return resp, nil
			}
		}
	}

	// Gunakan Singleflight untuk mencegah query berulang ke DB
	v, err, _ := s.sf.Do(cacheKey, func() (interface{}, error) {
		// Query ke Database
		departments, err := s.repo.FindAllByCompany(ctx, companyID)
		if err != nil {
			return nil, err
		}

		resp := mapToListResponse(departments)

		// Simpan ke Redis (TTL 30 Menit - 1 Jam cukup untuk data Master)
		if s.rdb != nil {
			if jsonData, err := json.Marshal(resp); err == nil {
				s.rdb.Set(ctx, cacheKey, jsonData, 30*time.Minute)
			}
		}

		return resp, nil
	})

	if err != nil {
		return nil, err
	}

	return v.([]DepartmentResponse), nil
}

func (s *service) GetByID(
	ctx context.Context,
	companyID, id string,
) (DepartmentResponse, error) {

	dept, err := s.repo.FindByIDAndCompany(ctx, companyID, id)
	if err != nil {
		return DepartmentResponse{}, err
	}

	return mapToResponse(*dept), nil
}

func (s *service) Update(
	ctx context.Context,
	companyID, id string,
	req UpdateDepartmentRequest,
) (DepartmentResponse, error) {

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return DepartmentResponse{}, err
	}
	defer tx.Rollback()

	qtx := s.repo.WithTx(tx)

	dept, err := qtx.FindByIDAndCompany(ctx, companyID, id)
	if err != nil {
		return DepartmentResponse{}, err
	}

	dept.Name = req.Name

	if err := qtx.Update(ctx, dept); err != nil {
		return DepartmentResponse{}, err
	}

	if err := tx.Commit(); err != nil {
		return DepartmentResponse{}, err
	}

	if s.rdb != nil {
		cacheKey := GetDepartmentAllKey(companyID)
		if err := s.rdb.Del(ctx, cacheKey).Err(); err != nil {
			log.Printf("ERROR: failed to invalidate cache for key %s: %v", cacheKey, err)
		}
	}

	return mapToResponse(*dept), nil
}

func (s *service) Delete(
	ctx context.Context,
	companyID, id string,
) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	qtx := s.repo.WithTx(tx)

	if err := qtx.Delete(ctx, companyID, id); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	// Invalidasi cache dilakukan tepat setelah data di DB resmi terhapus
	if s.rdb != nil {
		cacheKey := GetDepartmentAllKey(companyID)
		if err := s.rdb.Del(ctx, cacheKey).Err(); err != nil {
			log.Printf("ERROR: failed to invalidate cache for key %s: %v", cacheKey, err)
		}
	}

	return nil
}

func mapToResponse(dept Department) DepartmentResponse {
	return DepartmentResponse{
		ID:        dept.ID.String(),
		Name:      dept.Name,
		CompanyID: dept.CompanyID.String(),
	}
}

func mapToListResponse(depts []Department) []DepartmentResponse {
	res := make([]DepartmentResponse, len(depts))
	for i, d := range depts {
		res[i] = mapToResponse(d)
	}
	return res
}
