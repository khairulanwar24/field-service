package error

import "errors"

// Kumpulan variabel pesan error khusus untuk urusan Lapangan (Field).
var (
	ErrFieldNotFound = errors.New("field not found") // Error jika data lapangan yang dicari tidak ada di database.
)

// FieldErrors array penampung khusus kumpulan error Lapangan untuk dicocokkan nanti oleh modul error_mapping.go.
var FieldErrors = []error{
	ErrFieldNotFound,
}
