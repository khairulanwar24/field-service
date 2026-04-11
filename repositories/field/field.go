// Package repositories berisi logika untuk berinteraksi dengan database (Data Access Object)
// Di sini kita mendefinisikan bagaimana aplikasi mengambil, menyimpan, dan mengubah data Lapangan (Field)
package repositories

import (
	// package standard Go
	"context" // digunakan untuk mengelola lifecycle request dan timeout
	"errors"  // digunakan untuk menangani atau mengecek error

	// package internal aplikasi field-service
	errWrap "field-service/common/error"            // helper untuk membungkus (wrap) error
	errConstant "field-service/constants/error"      // daftar error umum (seperti error SQL)
	errField "field-service/constants/error/field"  // daftar error spesifik untuk domain Field
	"field-service/domain/dto"                      // Data Transfer Object untuk request/param
	"field-service/domain/models"                   // Model database (struct)

	"fmt" // untuk formatting string

	// package eksternal
	"github.com/google/uuid" // untuk membuat dan mengelola UUID
	"gorm.io/gorm"           // ORM (Object Relational Mapping) untuk database Go
)

// FieldRepository adalah struct yang menampung koneksi database
// Kita menggunakan *gorm.DB sebagai engine untuk query ke database
type FieldRepository struct {
	db *gorm.DB
}

// IFieldRepository adalah Interface (kontrak) yang mendefinisikan fungsi apa saja yang harus ada
// tujuannya agar kode lebih modular dan mudah di-test (mocking)
type IFieldRepository interface {
	// Menampilkan daftar lapangan dengan fitur halaman (paging), urutan (sorting), dll
	FindAllWithPagination(context.Context, *dto.FieldRequestParam) ([]models.Field, int64, error)
	// Menampilkan semua lapangan sekaligus tanpa batas
	FindAllWithoutPagination(context.Context) ([]models.Field, error)
	// Mencari satu lapangan berdasarkan ID (UUID) uniknya
	FindByUUID(context.Context, string) (*models.Field, error)
	// Membuat data lapangan baru ke database
	Create(context.Context, *models.Field) (*models.Field, error)
	// Mengubah data lapangan yang sudah ada berdasarkan UUID
	Update(context.Context, string, *models.Field) (*models.Field, error)
	// Menghapus data lapangan dari database berdasarkan UUID
	Delete(context.Context, string) error
}

// NewFieldRepository adalah fungsi "Constructor"
// Digunakan untuk membuat instance baru dari FieldRepository
func NewFieldRepository(db *gorm.DB) IFieldRepository {
	return &FieldRepository{db: db}
}

// FindAllWithPagination mengambil data lapangan dari DB dengan batasan jumlah (Limit) dan halaman (Offset)
func (f *FieldRepository) FindAllWithPagination(
	ctx context.Context,
	param *dto.FieldRequestParam,
) ([]models.Field, int64, error) {
	var (
		fields []models.Field // slice untuk menampung hasil list lapangan
		sort   string         // string untuk menyimpan perintah ORDER BY
		total  int64          // variabel untuk menampung total seluruh data di DB
	)

	// Cek apakah user mengirimkan kolom sort (misal: "name" atau "price")
	if param.SortColumn != nil {
		// Format: "nama_kolom asc/desc"
		sort = fmt.Sprintf("%s %s", *param.SortColumn, *param.SortOrder)
	} else {
		// Default urutan adalah berdasarkan waktu dibuat terbaru
		sort = "created_at desc"
	}

	// Hitung limit (berapa data per halaman) dan offset (mulai dari data ke berapa)
	limit := param.Limit
	offset := (param.Page - 1) * limit

	// Query utama menggunakan GORM
	err := f.db.
		WithContext(ctx). // menyambungkan dengan context
		Limit(limit).     // batasi jumlah data
		Offset(offset).   // loncati data sebelumnya (untuk paging)
		Order(sort).      // urutkan data
		Find(&fields).    // eksekusi SELECT * FROM fields... dan simpan ke slice fields
		Error
	if err != nil {
		// Jika ada error database, bungkus error-nya agar rapi
		return nil, 0, errWrap.WrapError(errConstant.ErrSQLError)
	}

	// Query kedua untuk menghitung total data (tanpa limit/offset)
	// Ini penting agar Frontend tahu ada berapa total halaman data
	err = f.db.
		WithContext(ctx).
		Model(&fields).
		Count(&total). // menghitung baris di tabel
		Error
	if err != nil {
		return nil, 0, errWrap.WrapError(errConstant.ErrSQLError)
	}

	return fields, total, nil
}

// FindAllWithoutPagination mengambil semua data lapangan tanpa fitur paging
func (f *FieldRepository) FindAllWithoutPagination(ctx context.Context) ([]models.Field, error) {
	var fields []models.Field
	err := f.db.
		WithContext(ctx).
		Find(&fields). // Ambil semua baris di tabel fields
		Error
	if err != nil {
		return nil, errWrap.WrapError(errConstant.ErrSQLError)
	}
	return fields, nil
}

// FindByUUID mencari satu data lapangan yang spesifik berdasarkan UUID-nya
func (f *FieldRepository) FindByUUID(ctx context.Context, uuid string) (*models.Field, error) {
	var field models.Field
	err := f.db.
		WithContext(ctx).
		Where("uuid = ?", uuid). // cari yang kolom uuid-nya sama dengan input
		First(&field).           // ambil data pertama yang ditemukan
		Error
	if err != nil {
		// Jika data tidak ditemukan di database
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errWrap.WrapError(errField.ErrFieldNotFound)
		}
		// Error teknis database lainnya
		return nil, errWrap.WrapError(errConstant.ErrSQLError)
	}
	return &field, nil
}

// Create memasukkan data lapangan baru ke dalam tabel
func (f *FieldRepository) Create(ctx context.Context, req *models.Field) (*models.Field, error) {
	// Siapkan objek model baru yang akan disimpan
	field := models.Field{
		UUID:         uuid.New(), // generate ID unik baru
		Code:         req.Code,
		Name:         req.Name,
		Images:       req.Images,
		PricePerHour: req.PricePerHour,
	}

	// Perintah INSERT ke database melalui GORM
	err := f.db.WithContext(ctx).Create(&field).Error
	if err != nil {
		return nil, errWrap.WrapError(errConstant.ErrSQLError)
	}
	return &field, nil // kembalikan data yang baru dibuat (lengkap dengan UUID-nya)
}

// Update mengubah data lapangan yang sudah ada di database
func (f *FieldRepository) Update(ctx context.Context, uuid string, req *models.Field) (*models.Field, error) {
	// Siapkan data perubahan
	field := models.Field{
		Code:         req.Code,
		Name:         req.Name,
		Images:       req.Images,
		PricePerHour: req.PricePerHour,
	}

	// Perintah UPDATE ke database
	// Where: cari data yang mana yang akan diupdate (berdasarkan UUID)
	// Updates: ganti nilai kolom dengan data baru dari struct 'field'
	err := f.db.WithContext(ctx).Where("uuid = ?", uuid).Updates(&field).Error
	if err != nil {
		return nil, errWrap.WrapError(errConstant.ErrSQLError)
	}
	return &field, nil
}

// Delete menghapus data lapangan secara permanen (atau soft delete jika diatur di model)
func (f *FieldRepository) Delete(ctx context.Context, uuid string) error {
	// Perintah DELETE dari database
	// Kita harus menyertakan models.Field{} agar GORM tahu tabel mana yang ditarget
	err := f.db.WithContext(ctx).Where("uuid = ?", uuid).Delete(&models.Field{}).Error
	if err != nil {
		return errWrap.WrapError(errConstant.ErrSQLError)
	}
	return nil
}

