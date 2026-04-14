package services

import (
	"field-service/common/gcs"
	"field-service/repositories"
	fieldService "field-service/services/field"
	fieldScheduleService "field-service/services/fieldschedule"
	timeService "field-service/services/time"
)

// Registry adalah pusat pendaftaran semua service yang ada.
// Dengan menggunakan Registry, kita cukup memanggil satu tempat untuk mendapatkan service apapun.
// Ini membantu kita mengelola dependensi (seperti repository dan GCS) di satu tempat saja.
type Registry struct {
	repository repositories.IRepositoryRegistry
	gcs        gcs.IGCSClient
}

// IServiceRegistry adalah interface yang menetapkan service apa saja yang tersedia di aplikasi ini.
type IServiceRegistry interface {
	GetField() fieldService.IFieldService
	GetFieldSchedule() fieldScheduleService.IFieldScheduleService
	GetTime() timeService.ITimeService
}

// NewServiceRegistry membuat instance baru dari Registry.
// Di sini kita memasukkan dependensi utama: repository (database) dan gcs (storage).
func NewServiceRegistry(repository repositories.IRepositoryRegistry, gcs gcs.IGCSClient) IServiceRegistry {
	return &Registry{
		repository: repository,
		gcs:        gcs,
	}
}

// GetField menginisialisasi dan mengambil service untuk Lapangan.
func (r *Registry) GetField() fieldService.IFieldService {
	return fieldService.NewFieldService(r.repository, r.gcs)
}

// GetFieldSchedule menginisialisasi dan mengambil service untuk Jadwal.
func (r *Registry) GetFieldSchedule() fieldScheduleService.IFieldScheduleService {
	return fieldScheduleService.NewFieldScheduleService(r.repository)
}

// GetTime menginisialisasi dan mengambil service untuk Master Jam.
func (r *Registry) GetTime() timeService.ITimeService {
	return timeService.NewTimeService(r.repository)
}
