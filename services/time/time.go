package services

import (
	"context"
	"field-service/domain/dto"
	"field-service/domain/models"
	"field-service/repositories"
)

// TimeService mengelola data Master Jam (Time Slot) yang tersedia untuk booking.
type TimeService struct {
	repository repositories.IRepositoryRegistry
}

// ITimeService mendefinisikan kontrak untuk layanan Master Jam.
type ITimeService interface {
	GetAll(context.Context) ([]dto.TimeResponse, error)
	GetByUUID(context.Context, string) (*dto.TimeResponse, error)
	Create(context.Context, *dto.TimeRequest) (*dto.TimeResponse, error)
}

// NewTimeService membuat instance baru dari TimeService.
func NewTimeService(repository repositories.IRepositoryRegistry) ITimeService {
	return &TimeService{repository: repository}
}

// GetAll mengambil semua daftar jam yang ada di database.
func (t *TimeService) GetAll(ctx context.Context) ([]dto.TimeResponse, error) {
	// 1. Ambil data dari repository.
	times, err := t.repository.GetTime().FindAll(ctx)
	if err != nil {
		return nil, err
	}

	// 2. Mapping ke DTO response.
	timeResults := make([]dto.TimeResponse, 0, len(times))
	for _, time := range times {
		timeResults = append(timeResults, dto.TimeResponse{
			UUID:      time.UUID,
			StartTime: time.StartTime,
			EndTime:   time.EndTime,
			CreatedAt: time.CreatedAt,
			UpdatedAt: time.UpdatedAt,
		})
	}

	return timeResults, nil
}

// GetByUUID mengambil detail satu jam berdasarkan UUID.
func (t *TimeService) GetByUUID(ctx context.Context, uuid string) (*dto.TimeResponse, error) {
	// 1. Cari data di repository.
	time, err := t.repository.GetTime().FindByUUID(ctx, uuid)
	if err != nil {
		return nil, err
	}

	// 2. Mapping ke response.
	timeResult := dto.TimeResponse{
		UUID:      time.UUID,
		StartTime: time.StartTime,
		EndTime:   time.EndTime,
		CreatedAt: time.CreatedAt,
		UpdatedAt: time.UpdatedAt,
	}

	return &timeResult, nil
}

// Create menambahkan master jam baru (misal: jam 08:00 - 09:00).
func (t *TimeService) Create(ctx context.Context, req *dto.TimeRequest) (*dto.TimeResponse, error) {
	// 1. Siapkan data untuk disimpan.
	time := &dto.TimeRequest{
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
	}

	// 2. Simpan ke database melalui repository.
	timeResult, err := t.repository.GetTime().Create(ctx, &models.Time{
		StartTime: time.StartTime,
		EndTime:   time.EndTime,
	})
	if err != nil {
		return nil, err
	}

	// 3. Mapping hasil simpan ke response.
	response := dto.TimeResponse{
		UUID:      timeResult.UUID,
		StartTime: timeResult.StartTime,
		EndTime:   timeResult.EndTime,
		CreatedAt: timeResult.CreatedAt,
		UpdatedAt: timeResult.UpdatedAt,
	}

	return &response, nil
}
