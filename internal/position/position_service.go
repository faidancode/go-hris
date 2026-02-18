package position

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
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
}

func NewService(db *sql.DB, repo Repository) Service {
	return &service{db: db, repo: repo}
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

	dept := &Position{
		ID:           uuid.New(),
		Name:         req.Name,
		CompanyID:    uuid.MustParse(companyID),
		DepartmentID: uuid.MustParse(req.DepartmentID),
	}

	if err := qtx.Create(ctx, dept); err != nil {
		return PositionResponse{}, err
	}

	if err := tx.Commit(); err != nil {
		return PositionResponse{}, err
	}

	return mapToResponse(*dept), nil
}

func (s *service) GetAll(
	ctx context.Context,
	companyID string,
) ([]PositionResponse, error) {

	depts, err := s.repo.FindAllByCompany(ctx, companyID)
	if err != nil {
		return nil, err
	}

	return mapToListResponse(depts), nil
}

func (s *service) GetByID(
	ctx context.Context,
	companyID, id string,
) (PositionResponse, error) {

	dept, err := s.repo.FindByIDAndCompany(ctx, companyID, id)
	if err != nil {
		return PositionResponse{}, err
	}

	return mapToResponse(*dept), nil
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

	dept, err := qtx.FindByIDAndCompany(ctx, companyID, id)
	if err != nil {
		return PositionResponse{}, err
	}

	dept.Name = req.Name
	dept.DepartmentID = uuid.MustParse(req.DepartmentID)

	if err := qtx.Update(ctx, dept); err != nil {
		return PositionResponse{}, err
	}

	if err := tx.Commit(); err != nil {
		return PositionResponse{}, err
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

	return tx.Commit()
}

func mapToResponse(dept Position) PositionResponse {
	resp := PositionResponse{
		ID:        dept.ID.String(),
		Name:      dept.Name,
		CompanyID: dept.CompanyID.String(),
	}
	if dept.DepartmentID != uuid.Nil {
		resp.DepartmentID = dept.DepartmentID.String()
	}
	if dept.Department != nil {
		resp.DepartmentName = dept.Department.Name
	}
	if !dept.CreatedAt.IsZero() {
		resp.CreatedAt = dept.CreatedAt.Format(time.RFC3339)
	}
	if !dept.UpdatedAt.IsZero() {
		resp.UpdatedAt = dept.UpdatedAt.Format(time.RFC3339)
	}
	return resp
}

func mapToListResponse(depts []Position) []PositionResponse {
	res := make([]PositionResponse, len(depts))
	for i, d := range depts {
		res[i] = mapToResponse(d)
	}
	return res
}
