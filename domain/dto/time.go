package dto

import (
	"github.com/google/uuid"
	"time"
)

// TimeRequest adalah format data yang diterima khusus saat kita ingin membuat jam/slot waktu baru.
// Contohnya: mengirimkan format JSON {"startTime": "08:00:00", "endTime": "09:00:00"} dari Postman atau Frontend.
type TimeRequest struct {
	// json:"startTime" berarti dari data JSON kita mengambil nilai di dalam kunci "startTime".
	// validate:"required" berarti isian jam mulai wajib ada dan tidak boleh dikosongi!
	StartTime string `json:"startTime" validate:"required"`
	
	// Jam selesai wajib diisi (contoh: "09:00:00")
	EndTime   string `json:"endTime" validate:"required"`
}

// TimeResponse adalah bentuk data keluaran yang akan dikirimkan KEMBALI oleh API (server) ke peminta (klien/Frontend).
// Bentuknya sudah kita atur (tambahkan UUID dan waktu dibuat) agar mudah ditampilkan.
type TimeResponse struct {
	UUID      uuid.UUID  `json:"uuid"`       // ID unik global dari slot waktu tersebut
	StartTime string     `json:"startTime"`  // Informasi jam mulai main
	EndTime   string     `json:"endTime"`    // Informasi jam selesai main
	CreatedAt *time.Time `json:"createdAt"`  // Waktu data ini pertama kali dimasukkan ke database
	UpdatedAt *time.Time `json:"updatedAt"`  // Waktu data ini terakhir diubah
}
