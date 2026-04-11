// Package repositories mengelola semua interaksi ke database.
// File registry.go ini berperan sebagai "Pusat" atau "Induk" dari semua repository yang ada.
package repositories

import (
	// Mengimpor sub-package repository masing-masing domain (Lahan, Jadwal, Waktu)
	fieldRepo "field-service/repositories/field"
	fieldScheduleRepo "field-service/repositories/fieldschedule"
	timeRepo "field-service/repositories/time"
	
	"gorm.io/gorm" // library untuk koneksi database
)

// Registry adalah struct yang menampung satu koneksi database (db).
// Objek ini akan digunakan untuk menyebarkan koneksi database tersebut ke semua repository anak.
type Registry struct {
	db *gorm.DB
}

// IRepositoryRegistry adalah Interface (kontrak) pusat.
// Tujuannya agar lapisan lain (seperti Service) bisa mengakses semua repository melalui satu pintu saja.
type IRepositoryRegistry interface {
	GetField() fieldRepo.IFieldRepository
	GetFieldSchedule() fieldScheduleRepo.IFieldScheduleRepository
	GetTime() timeRepo.ITimeRepository
}

// NewRepositoryRegistry adalah fungsi untuk membuat "Pusat Repository" baru.
// Kita memasukkannya koneksi database gorm di sini.
func NewRepositoryRegistry(db *gorm.DB) IRepositoryRegistry {
	return &Registry{db: db}
}

// GetField digunakan untuk mendapatkan akses ke repository Lahan (Field).
func (r *Registry) GetField() fieldRepo.IFieldRepository {
	// Di sini kita membuat instance baru dari FieldRepository sambil memberikan koneksi DB
	return fieldRepo.NewFieldRepository(r.db)
}

// GetFieldSchedule digunakan untuk mendapatkan akses ke repository Jadwal (Field Schedule).
func (r *Registry) GetFieldSchedule() fieldScheduleRepo.IFieldScheduleRepository {
	return fieldScheduleRepo.NewFieldScheduleRepository(r.db)
}

// GetTime digunakan untuk mendapatkan akses ke repository Waktu (Time).
func (r *Registry) GetTime() timeRepo.ITimeRepository {
	return timeRepo.NewTimeRepository(r.db)
}

