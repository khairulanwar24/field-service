// Package repositories menangani semua operasi atau interaksi langsung dengan database
// File ini khusus untuk mengelola data 'Waktu' (Time) seperti jam operasional
package repositories

import (
	// package standard Go
	"context" // untuk mengontrol durasi request dan lifecycle-nya
	"errors"  // untuk mengelola pesan error

	// package internal proyek
	errWrap "field-service/common/error"         // helper untuk membungkus error agar informatif
	errConstant "field-service/constants/error"   // kumpulan konstanta error umum (DB error, dll)
	errTime "field-service/constants/error/time" // kumpulan konstanta error spesifik domain Waktu
	"field-service/domain/models"                // struktur tabel database
	"fmt"                                        // untuk menampilkan log/output ke terminal

	// package pihak ketiga
	"github.com/google/uuid" // untuk membuat ID unik (UUID) secara otomatis
	"gorm.io/gorm"           // library ORM untuk memudahkan query database
)

// TimeRepository adalah struktur data yang menyimpan koneksi database
type TimeRepository struct {
	db *gorm.DB
}

// ITimeRepository adalah Interface (kontrak) yang mendefinisikan apa saja yang bisa dilakukan repo ini
type ITimeRepository interface {
	// Mengambil semua data waktu yang ada di tabel
	FindAll(context.Context) ([]models.Time, error)
	// Mencari waktu berdasarkan UUID (ID string unik)
	FindByUUID(context.Context, string) (*models.Time, error)
	// Mencari waktu berdasarkan ID (angka auto-increment)
	FindByID(context.Context, int) (*models.Time, error)
	// Menambahkan data waktu baru ke database
	Create(context.Context, *models.Time) (*models.Time, error)
}

// NewTimeRepository adalah fungsi untuk menginisialisasi repository baru
func NewTimeRepository(db *gorm.DB) ITimeRepository {
	return &TimeRepository{db: db}
}

// FindAll mengambil seluruh daftar jam/waktu yang tersedia dari database
func (t *TimeRepository) FindAll(ctx context.Context) ([]models.Time, error) {
	var times []models.Time
	// .WithContext(ctx) memastikan query mengikuti aturan context (timeout, dll)
	// .Find(&times) menjalankan 'SELECT * FROM times' dan memasukkannya ke slice 'times'
	err := t.db.WithContext(ctx).Find(&times).Error
	if err != nil {
		// Jika ada error database, kita bungkus dengan pesan error SQL umum
		return nil, errWrap.WrapError(errConstant.ErrSQLError)
	}

	return times, nil
}

// FindByUUID mencari data waktu menggunakan kolom UUID
func (t *TimeRepository) FindByUUID(ctx context.Context, uuid string) (*models.Time, error) {
	var time models.Time
	// .Where("uuid = ?", uuid) menambahkan filter pencarian
	// .First(&time) mengambil satu baris pertama yang ditemukan
	err := t.db.WithContext(ctx).Where("uuid = ?", uuid).First(&time).Error
	if err != nil {
		// Cek apakah error-nya disebabkan karena datanya memang tidak ada
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errWrap.WrapError(errTime.ErrTimeNotFound)
		}
		// Jika error teknis lainnya (misal: koneksi putus)
		return nil, errWrap.WrapError(errConstant.ErrSQLError)
	}

	return &time, nil
}

// FindByID mencari data waktu menggunakan Primary Key (ID angka)
func (t *TimeRepository) FindByID(ctx context.Context, id int) (*models.Time, error) {
	var time models.Time
	// Mirip dengan FindByUUID, tapi pencarian dilakukan via kolom ID
	err := t.db.WithContext(ctx).Where("id = ?", id).First(&time).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errWrap.WrapError(errTime.ErrTimeNotFound)
		}
		return nil, errWrap.WrapError(errConstant.ErrSQLError)
	}

	return &time, nil
}

// Create menyimpan data waktu baru ke dalam tabel database
func (t *TimeRepository) Create(ctx context.Context, time *models.Time) (*models.Time, error) {
	// Sebelum disimpan, kita buatkan ID unik (UUID) baru
	time.UUID = uuid.New()
	
	// Print ke terminal untuk mempermudah debugging (opsional)
	fmt.Println("Membuat data waktu baru:", time)

	// .Create(time) menjalankan perintah 'INSERT INTO times ...'
	err := t.db.WithContext(ctx).Create(time).Error
	if err != nil {
		return nil, errWrap.WrapError(errConstant.ErrSQLError)
	}
	return time, nil
}

