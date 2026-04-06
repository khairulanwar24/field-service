package models

import (
	"field-service/constants"
	"github.com/google/uuid"
	"time"
)

// FieldSchedule adalah representasi dari tabel 'field_schedules' di database.
// Struct ini digunakan untuk mencatat jadwal lapangan (tanggal berapa, jam berapa, lapangan mana).
type FieldSchedule struct {
	// ID adalah kunci utama (Primary Key) untuk tabel ini
	ID        uint                          `gorm:"primaryKey;autoIncrement"`
	
	// UUID adalah kode unik global untuk data jadwal ini
	UUID      uuid.UUID                     `gorm:"type:uuid;not null"`
	
	// FieldID adalah "kunci tamu" (Foreign Key) yang merujuk pada ID di tabel Field (Lapangan)
	FieldID   uint                          `gorm:"type:int;not null"`
	
	// TimeID adalah "kunci tamu" (Foreign Key) yang merujuk pada ID di tabel Time (Waktu)
	TimeID    uint                          `gorm:"type:int;not null"`
	
	// Date menyimpan informasi tanggal jadwal tersebut
	Date      time.Time                     `gorm:"type:date;not null"`
	
	// Status menunjukkan keadaan jadwal (misal: tersedia, sudah disewa, dll)
	Status    constants.FieldScheduleStatus `gorm:"type:int;not null"`
	
	// Penanda waktu otomatis untuk pencatatan (Kapan dibuat, diubah, maupun dihapus)
	CreatedAt *time.Time
	UpdatedAt *time.Time
	DeletedAt *time.Time
	
	// MENDAPATKAN RELASI (Relationship):
	// Relasi ke data asal Lapangan (Berelasi dengan FieldID di atas)
	Field     Field `gorm:"foreignKey:field_id;references:id;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	
	// Relasi ke data asal Waktu (Berelasi dengan TimeID di atas)
	Time      Time  `gorm:"foreignKey:time_id;references:id;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}
