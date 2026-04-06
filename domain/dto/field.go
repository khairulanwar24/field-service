package dto

import (
	"github.com/google/uuid"
	"mime/multipart"
	"time"
)

// DTO (Data Transfer Object)
// File ini membungkus struktur data yang khusus digunakan untuk alur keluar-masuk (Request/Response)
// aplikasi, berbeda dengan Models yang langsung berhubungan ke Database.

// FieldRequest digunakan ketika menerima input masukan dari pengguna (seperti mengisi form tambah lapangan).
// Kata 'Request' berarti ini adalah 'Permintaan' dari klien/frontend ke server.
type FieldRequest struct {
	// form:"name" artinya struktur ini akan menangkap data dari field input form bernamakan "name".
	// validate:"required" artinya field ini hukumnya wajib diisi (tidak boleh kosong).
	Name         string                 `form:"name" validate:"required"`
	Code         string                 `form:"code" validate:"required"`
	PricePerHour int                    `form:"pricePerHour" validate:"required"`
	
	// multipart.FileHeader digunakan khusus untuk menerima file yang di-upload (diunggah).
	Images       []multipart.FileHeader `form:"images" validate:"required"`
}

// UpdateFieldRequest digunakan saat ingin memperbarui (Edit/Update) data lapangan yang sudah ada.
type UpdateFieldRequest struct {
	Name         string                 `form:"name" validate:"required"`
	Code         string                 `form:"code" validate:"required"`
	PricePerHour int                    `form:"pricePerHour" validate:"required"`
	
	// Perhatikan bahwa di sini tidak ada tag validate:"required".
	// Artinya, saat mengupdate data, tidak update gambar juga tidak apa-apa (opsional).
	Images       []multipart.FileHeader `form:"images"`
}

// FieldResponse adalah bungkus data yang akan dikirim KEMBALI dari server ke klien (Response/Jawaban).
// Tag `json` menandakan data ini akan diubah (encode) ke dalam format JSON sebelum dikirim.
type FieldResponse struct {
	UUID         uuid.UUID  `json:"uuid"`
	Code         string     `json:"code"`
	Name         string     `json:"name"`
	PricePerHour any        `json:"pricePerHour"` // Tipe 'any' berarti bisa menampung bentuk tipe data apa saja
	Images       []string   `json:"images"`       // Karena berupa respon, gambar akan dikirim sebagai kumpulan Link/URL
	CreatedAt    *time.Time `json:"createdAt"`
	UpdatedAt    *time.Time `json:"updatedAt"`
}

// FieldDetailResponse mirip seperti FieldResponse tapi difokuskan ketika kita ingin
// mengembalikan data mendetail untuk 1 lapangan yang spesifik.
type FieldDetailResponse struct {
	Code         string     `json:"code"`
	Name         string     `json:"name"`
	PricePerHour int        `json:"pricePerHour"`
	Images       []string   `json:"images"`
	CreatedAt    *time.Time `json:"createdAt"`
	UpdatedAt    *time.Time `json:"updatedAt"`
}

// FieldRequestParam digunakan untuk menampung parameter permintaan tambahan.
// Biasanya ini disertakan pada saat mengambil data berupa daftar list (Pagination).
type FieldRequestParam struct {
	Page       int     `form:"page" validate:"required"`  // Meminta halaman ke-berapa
	Limit      int     `form:"limit" validate:"required"` // Berapa jumlah data maksimal per-halamannya
	SortColumn *string `form:"sortColumn"`                // Ingin disusun berdasarkan kolom apa? (Opsional, ditandai '*' (Pointer))
	SortOrder  *string `form:"sortOrder"`                 // Urutan 'asc' (kecil-besar) atau 'desc' (besar-kecil)? (Opsional)
}
