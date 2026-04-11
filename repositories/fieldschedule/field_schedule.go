// Package repositories menangani interaksi langsung dengan database
// File ini spesifik menangani jadwal lapangan (field schedule)
package repositories

import (
	// package standard Go
	"context" // digunakan untuk membatasi durasi eksekusi query (timeout/cancel)
	"errors"  // digunakan untuk pengecekan tipe error

	// package internal aplikasi
	errWrap "field-service/common/error"                           // helper untuk pembungkus error
	"field-service/constants"                                       // konstanta aplikasi (misal status jadwal)
	errConstant "field-service/constants/error"                     // pesan error umum
	errFieldSchedule "field-service/constants/error/fieldschedule" // pesan error spesifik jadwal
	"field-service/domain/dto"                                      // objek data untuk request dari user
	"field-service/domain/models"                                   // struktur tabel database
	"fmt"                                                           // untuk manipulasi string

	"gorm.io/gorm" // library ORM untuk database
)

// FieldScheduleRepository menyimpan dependensi database (GORM)
type FieldScheduleRepository struct {
	db *gorm.DB
}

// IFieldScheduleRepository adalah kontrak (interface) yang mendefinisikan fungsi apa saja yang tersedia
type IFieldScheduleRepository interface {
	// Mengambil data dengan pagination (halaman)
	FindAllWithPagination(context.Context, *dto.FieldScheduleRequestParam) ([]models.FieldSchedule, int64, error)
	// Mengambil jadwal berdasarkan ID lapangan dan tanggal tertentu
	FindAllByFieldIDAndDate(context.Context, int, string) ([]models.FieldSchedule, error)
	// Mencari satu jadwal berdasarkan UUID
	FindByUUID(context.Context, string) (*models.FieldSchedule, error)
	// Mencari jadwal berdasarkan tanggal, jam, dan lapangan (untuk validasi bentrok)
	FindByDateAndTimeID(context.Context, string, int, int) (*models.FieldSchedule, error)
	// Membuat banyak jadwal sekaligus (batch insert)
	Create(context.Context, []models.FieldSchedule) error
	// Mengubah data jadwal yang sudah ada
	Update(context.Context, string, *models.FieldSchedule) (*models.FieldSchedule, error)
	// Mengubah hanya status jadwal (misal: dari 'tersedia' menjadi 'dipesan')
	UpdateStatus(context.Context, constants.FieldScheduleStatus, string) error
	// Menghapus jadwal berdasarkan UUID
	Delete(context.Context, string) error
}

// NewFieldScheduleRepository adalah constructor untuk membuat instance repository
func NewFieldScheduleRepository(db *gorm.DB) IFieldScheduleRepository {
	return &FieldScheduleRepository{db: db}
}

// FindAllWithPagination mengambil data jadwal dengan dukungan halaman dan urutan
func (f *FieldScheduleRepository) FindAllWithPagination(
	ctx context.Context,
	param *dto.FieldScheduleRequestParam,
) ([]models.FieldSchedule, int64, error) {
	var (
		fieldSchedules []models.FieldSchedule
		sort           string
		total          int64
	)

	// Tentukan kolom pengurutan (sorting)
	if param.SortColumn != nil {
		sort = fmt.Sprintf("%s %s", *param.SortColumn, *param.SortOrder)
	} else {
		sort = "created_at desc"
	}

	limit := param.Limit
	offset := (param.Page - 1) * limit

	// Query utama
	err := f.db.
		WithContext(ctx).
		Preload("Field"). // Preload mengambil data dari tabel 'fields' yang berelasi
		Preload("Time").  // Preload mengambil data dari tabel 'times' yang berelasi
		Limit(limit).
		Offset(offset).
		Order(sort).
		Find(&fieldSchedules). // Ambil datanya
		Error
	if err != nil {
		return nil, 0, errWrap.WrapError(errConstant.ErrSQLError)
	}

	// Hitung total data untuk keperluan Frontend
	err = f.db.
		WithContext(ctx).
		Model(&fieldSchedules).
		Count(&total).
		Error
	if err != nil {
		return nil, 0, errWrap.WrapError(errConstant.ErrSQLError)
	}

	return fieldSchedules, total, nil
}

// FindAllByFieldIDAndDate mencari jadwal khusus untuk satu lapangan pada tanggal tertentu
func (f *FieldScheduleRepository) FindAllByFieldIDAndDate(
	ctx context.Context,
	fieldID int,
	date string,
) ([]models.FieldSchedule, error) {
	var fieldSchedules []models.FieldSchedule
	err := f.db.
		WithContext(ctx).
		Preload("Field").
		Preload("Time").
		Where("field_id = ?", fieldID).
		Where("date = ?", date).
		// Joins digunakan untuk menggabungkan tabel 'times' agar bisa mengurutkan berdasarkan jam mulai
		Joins("LEFT JOIN times ON field_schedules.time_id = times.id").
		Order("times.start_time asc").
		Find(&fieldSchedules).
		Error
	if err != nil {
		return nil, errWrap.WrapError(errConstant.ErrSQLError)
	}
	return fieldSchedules, nil
}

// FindByUUID mencari satu jadwal spesifik
func (f *FieldScheduleRepository) FindByUUID(ctx context.Context, uuid string) (*models.FieldSchedule, error) {
	var fieldSchedule models.FieldSchedule
	err := f.db.
		WithContext(ctx).
		Preload("Field").
		Preload("Time").
		Where("uuid = ?", uuid).
		First(&fieldSchedule). // Ambil record pertama yang cocok
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errWrap.WrapError(errFieldSchedule.ErrFieldScheduleNotFound)
		}
		return nil, errWrap.WrapError(errConstant.ErrSQLError)
	}
	return &fieldSchedule, nil
}

// FindByDateAndTimeID mengecek apakah sudah ada jadwal di jam dan tanggal yang sama (mencegah duplikasi)
func (f *FieldScheduleRepository) FindByDateAndTimeID(
	ctx context.Context,
	date string,
	timeID int,
	fieldID int) (*models.FieldSchedule, error) {
	var fieldSchedule models.FieldSchedule
	err := f.db.
		WithContext(ctx).
		Where("date = ?", date).
		Where("time_id = ?", timeID).
		Where("field_id = ?", fieldID).
		First(&fieldSchedule).
		Error
	if err != nil {
		// Jika tidak ditemukan, kita return nil tanpa error karena ini hal yang normal dalam pengecekan
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, errWrap.WrapError(errConstant.ErrSQLError)
	}
	return &fieldSchedule, nil
}

// Create menyimpan satu atau beberapa jadwal baru sekaligus (batch insert)
func (f *FieldScheduleRepository) Create(ctx context.Context, req []models.FieldSchedule) error {
	// GORM secara otomatis melakukan batch insert jika inputnya adalah slice/array
	err := f.db.WithContext(ctx).Create(&req).Error
	if err != nil {
		return errWrap.WrapError(errConstant.ErrSQLError)
	}
	return nil
}

// Update mengubah data tanggal pada jadwal
func (f *FieldScheduleRepository) Update(
	ctx context.Context,
	uuid string,
	req *models.FieldSchedule,
) (*models.FieldSchedule, error) {
	// Cari dulu datanya ada atau tidak
	fieldSchedule, err := f.FindByUUID(ctx, uuid)
	if err != nil {
		return nil, err
	}

	fieldSchedule.Date = req.Date
	// Save akan memperbarui seluruh kolom berdasarkan data yang ada di struct
	err = f.db.WithContext(ctx).Save(&fieldSchedule).Error
	if err != nil {
		return nil, errWrap.WrapError(errConstant.ErrSQLError)
	}
	return fieldSchedule, nil
}

// UpdateStatus mengubah status jadwal (misal: 'Available' -> 'Booked')
func (f *FieldScheduleRepository) UpdateStatus(
	ctx context.Context,
	status constants.FieldScheduleStatus,
	uuid string,
) error {
	// Cari dulu datanya
	fieldSchedule, err := f.FindByUUID(ctx, uuid)
	if err != nil {
		return err
	}

	fieldSchedule.Status = status
	// Save melakukan pembaruan ke database
	err = f.db.WithContext(ctx).Save(&fieldSchedule).Error
	if err != nil {
		return errWrap.WrapError(errConstant.ErrSQLError)
	}
	return nil
}

// Delete menghapus data jadwal berdasarkan UUID
func (f *FieldScheduleRepository) Delete(ctx context.Context, uuid string) error {
	// Delete menghapus record yang cocok dengan kriteria where
	err := f.db.WithContext(ctx).Where("uuid = ?", uuid).Delete(&models.FieldSchedule{}).Error
	if err != nil {
		return errWrap.WrapError(errConstant.ErrSQLError)
	}
	return nil
}

