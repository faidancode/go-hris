package position

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/singleflight"
)

//go:generate mockgen -source=position_service.go -destination=mock/position_service_mock.go -package=mock
type Service interface {
	Create(ctx context.Context, companyID string, req CreatePositionRequest) (PositionResponse, error)
	GetAll(ctx context.Context, companyID string) ([]PositionResponse, error)
	GetByID(ctx context.Context, companyID, id string) (PositionResponse, error)
	Update(ctx context.Context, companyID, id string, req UpdatePositionRequest) (PositionResponse, error)
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
	req CreatePositionRequest,
) (PositionResponse, error) {

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return PositionResponse{}, err
	}
	defer tx.Rollback()

	qtx := s.repo.WithTx(tx)

	post := &Position{
		ID:           uuid.New(),
		Name:         req.Name,
		CompanyID:    uuid.MustParse(companyID),
		DepartmentID: uuid.MustParse(req.DepartmentID),
	}

	if err := qtx.Create(ctx, post); err != nil {
		return PositionResponse{}, err
	}

	if err := tx.Commit(); err != nil {
		return PositionResponse{}, err
	}

	return mapToResponse(*post), nil
}

func (s *service) GetAll(
	ctx context.Context,
	companyID string,
) ([]PositionResponse, error) {
	// 1. Definisikan Key yang unik per Company
	cacheKey := fmt.Sprintf("positions:all:%s", companyID)

	// 2. Coba ambil dari Redis
	if s.rdb != nil {
		cached, err := s.rdb.Get(ctx, cacheKey).Result()
		if err == nil {
			var resp []PositionResponse
			if err := json.Unmarshal([]byte(cached), &resp); err == nil {
				return resp, nil
			}
		}
	}

	// 3. Gunakan Singleflight untuk mencegah query berulang ke DB
	v, err, _ := s.sf.Do(cacheKey, func() (interface{}, error) {
		// Query ke Database
		positions, err := s.repo.FindAllByCompany(ctx, companyID)
		if err != nil {
			return nil, err
		}

		resp := mapToListResponse(positions)

		// 4. Simpan ke Redis (TTL 30 Menit - 1 Jam cukup untuk data Master)
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

	return v.([]PositionResponse), nil
}

func (s *service) GetByID(
	ctx context.Context,
	companyID, id string,
) (PositionResponse, error) {

	post, err := s.repo.FindByIDAndCompany(ctx, companyID, id)
	if err != nil {
		return PositionResponse{}, err
	}

	return mapToResponse(*post), nil
}

func (s *service) Update(
	ctx context.Context,
	companyID, id string,
	req UpdatePositionRequest,
) (PositionResponse, error) {

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return PositionResponse{}, err
	}
	defer tx.Rollback()

	qtx := s.repo.WithTx(tx)

	post, err := qtx.FindByIDAndCompany(ctx, companyID, id)
	if err != nil {
		return PositionResponse{}, err
	}

	post.Name = req.Name
	post.DepartmentID = uuid.MustParse(req.DepartmentID)

	if err := qtx.Update(ctx, post); err != nil {
		return PositionResponse{}, err
	}

	if err := tx.Commit(); err != nil {
		return PositionResponse{}, err
	}

	return mapToResponse(*post), nil
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

	return tx.Commit()
}

func mapToResponse(post Position) PositionResponse {
	resp := PositionResponse{
		ID:        post.ID.String(),
		Name:      post.Name,
		CompanyID: post.CompanyID.String(),
	}
	if post.DepartmentID != uuid.Nil {
		resp.DepartmentID = post.DepartmentID.String()
	}
	if post.Department != nil {
		resp.DepartmentName = post.Department.Name
	}
	if !post.CreatedAt.IsZero() {
		resp.CreatedAt = post.CreatedAt.Format(time.RFC3339)
	}
	if !post.UpdatedAt.IsZero() {
		resp.UpdatedAt = post.UpdatedAt.Format(time.RFC3339)
	}
	return resp
}

func mapToListResponse(posts []Position) []PositionResponse {
	res := make([]PositionResponse, len(posts))
	for i, d := range posts {
		res[i] = mapToResponse(d)
	}
	return res
}
