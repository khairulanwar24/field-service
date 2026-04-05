package error

import "errors"

// Kumpulan variabel pesan error khusus untuk urusan Waktu (Time).
var (
	ErrTimeNotFound = errors.New("time not found") // Error jika slot waktu yang dicari tidak ditemukan di database.
)

// TimeErrors array penampung error Waktu untuk proses pencocokan mapping.
var TimeErrors = []error{
	ErrTimeNotFound,
}
