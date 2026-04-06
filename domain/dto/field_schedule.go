package dto

import (
	"field-service/constants"
	"github.com/google/uuid"
	"time"
)

// FieldScheduleRequest adalah format data yang diterima saat membuat jadwal baru.
// Berbeda dari form data biasa, format ini menerima data dalam bentuk JSON.
type FieldScheduleRequest struct {
	// json:"fieldID" berarti data mentah yang masuk dari JSON memakai kunci "fieldID".
	// validate:"required" berarti ID lapangan ini wajib ada.
	FieldID string   `json:"fieldID" validate:"required"`
	Date    string   `json:"date" validate:"required"`     // Tanggal untuk jadwal dibuat (contoh format: "2024-05-01")
	TimeIDs []string `json:"timeIDs" validate:"required"`  // Daftar ID slot waktu (array of string), misalnya ["1", "2"]
}

// GenerateFieldScheduleForOneMonthRequest digunakan untuk sebuah fitur praktis:
// membuat jadwal otomatis selama satu bulan penuh untuk satu lapangan tertentu.
type GenerateFieldScheduleForOneMonthRequest struct {
	FieldID string `json:"fieldID" validate:"required"` // Lapangan mana yang ingin di-generate satu bulan?
}

// UpdateFieldScheduleRequest adalah bungkus data ketika kita ingin mengubah/memperbarui
// tanggal atau slot waktu di jadwal yang sudah ada.
type UpdateFieldScheduleRequest struct {
	Date   string `json:"date" validate:"required"`
	TimeID string `json:"timeID" validate:"required"`
}

// UpdateStatusFieldScheduleRequest digunakan ketika kita hanya ingin mengubah 
// status jadwal (contoh: dari 'Tersedia' menjadi 'Dipesan/Booked') secara sekaligus banyak (bulk/massal).
type UpdateStatusFieldScheduleRequest struct {
	// Menggunakan List/Array agar memperbarui data ganda sekaligus lebih cepat.
	FieldScheduleIDs []string `json:"fieldScheduleIDs" validate:"required"` 
}

// FieldScheduleResponse adalah data balasan utuh yang dikirim KEMBALI ke klien/pengguna.
// Strukturnya sudah dirakit dan disesuaikan agar mudah dibaca di Frontend.
type FieldScheduleResponse struct {
	UUID         uuid.UUID                         `json:"uuid"`         // ID unik global jadwal
	FieldName    string                            `json:"fieldName"`    // Nama lapangan yang terkait
	PricePerHour int                               `json:"pricePerHour"` // Harga sewa per jam
	Date         string                            `json:"date"`         // Tanggal jadwal tersebut
	Status       constants.FieldScheduleStatusName `json:"status"`       // Menampilkan status (contoh: AVAILABLE / BOOKED)
	Time         string                            `json:"time"`         // Informasi rentang waktu (contoh: "08:00 - 09:00")
	CreatedAt    *time.Time                        `json:"createdAt"`    // Waktu data dicatat
	UpdatedAt    *time.Time                        `json:"updatedAt"`
}

// FieldScheduleForBookingResponse adalah versi "lebih ringkas" dari FieldScheduleResponse.
// Digunakan di saat pengguna memilih jam booking, karena frontend tidak butuh data yang terlalu lengkap/berat.
type FieldScheduleForBookingResponse struct {
	UUID         uuid.UUID                         `json:"uuid"`
	PricePerHour string                            `json:"pricePerHour"` // Harga dikonversi menjadi tipe string
	Date         string                            `json:"date"`
	Status       constants.FieldScheduleStatusName `json:"status"`
	Time         string                            `json:"time"`
}

// FieldScheduleRequestParam menampung parameter tambahan untuk proses mengambil Daftar/List Jadwal.
// Biasa dipakai untuk fitur Pembagian Halaman (Pagination) dan "Sorting" (Mengurutkan Data).
type FieldScheduleRequestParam struct {
	Page       int     `form:"page" validate:"required"`  // Mau memuat halaman nomor berapa?
	Limit      int     `form:"limit" validate:"required"` // Berapa biji data per halamannya?
	SortColumn *string `form:"sortColumn"`                // Berdasarkan kolom mana kita ingin mengurutkan? (Opsional)
	SortOrder  *string `form:"sortOrder"`                 // Urutan membesar (asc) atau mengecil (desc)? (Opsional)
}

// FieldScheduleByFieldIDAndDateRequestParam digunakan saat mencari jadwal
// berdasarkan 1 kriteria tertentu untuk ditaruh di URL pencarian (contoh url: ?date=2024-05-12).
type FieldScheduleByFieldIDAndDateRequestParam struct {
	Date string `form:"date" validate:"required"` // Harus memasukkan tanggal agar dapat dicari.
}
