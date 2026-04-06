package models

import (
	"github.com/google/uuid"
	"time"
)

// Time adalah representasi dari tabel 'times' atau slot waktu di database.
// Struct ini digunakan untuk mencatat jam mulai dan jam selesai (contoh: 08:00 - 09:00).
type Time struct {
	// ID adalah kunci utama (Primary Key) untuk jadwal waktu
	ID        uint      `gorm:"primaryKey;autoIncrement"`
	
	// UUID adalah identifier unik global untuk setiap slot waktu
	UUID      uuid.UUID `gorm:"type:uuid;not null"`
	
	// StartTime menyimpan waktu mulai dalam format jam (contoh: "08:00:00")
	StartTime string    `gorm:"type:time without time zone;not null"`
	
	// EndTime menyimpan waktu selesai dalam format jam (contoh: "09:00:00")
	EndTime   string    `gorm:"type:time without time zone;not null"`
	
	// CreatedAt akan diisi secara otomatis waktu data ini dibuat ke database
	CreatedAt *time.Time
	
	// UpdatedAt akan diisi secara otomatis waktu data ini diperbarui di database
	UpdatedAt *time.Time
}
