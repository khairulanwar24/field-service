package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

// Field adalah representasi dari tabel 'fields' di database.
// Struct ini digunakan sebagai cetak biru (blueprint) data lapangan.
type Field struct {
	// ID adalah kunci utama (Primary Key) yang sifatnya bertambah otomatis (autoIncrement)
	ID            uint           `gorm:"primaryKey;autoIncrement"`
	
	// UUID adalah identifier unik global (seperti nomor KTP yang unik untuk setiap data)
	UUID          uuid.UUID      `gorm:"type:uuid;not null"`
	
	// Code adalah kode lapangan (contoh: 'LAP-01'), maksimal 15 karakter
	Code          string         `gorm:"type:varchar(15);not null"`
	
	// Name adalah nama lapangan (contoh: 'Lapangan Sintetis A')
	Name          string         `gorm:"type:varchar(100);not null"`
	
	// PricePerHour adalah harga sewa lapangan per jam
	PricePerHour  int            `gorm:"type:int;not null"`
	
	// Images adalah daftar URL atau text untuk gambar-gambar lapangan
	Images        pq.StringArray `gorm:"type:text[]; not null"`
	
	// CreatedAt akan otomatis menyimpan waktu kapan data pertama kali dibuat
	CreatedAt     *time.Time
	
	// UpdatedAt akan otomatis menyimpan waktu kapan data terakhir kali diubah
	UpdatedAt     *time.Time
	
	// DeletedAt digunakan oleh GORM untuk "Soft Delete" (data seolah terhapus tapi masih ada di database)
	DeletedAt     *gorm.DeletedAt
	
	// MENDAPATKAN RELASI (Relationship):
	// Satu lapangan (Field) bisa memiliki banyak Jadwal (FieldSchedule).
	// Ini adalah relasi One-to-Many (Satu ke Banyak).
	FieldSchedule []FieldSchedule `gorm:"foreignKey:field_id;references:id;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}
