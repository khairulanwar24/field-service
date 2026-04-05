package error

import (
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"strings"
)

// ValidationResponse adalah struktur untuk menampung pesan kesalahan validasi tiap kolom/field.
type ValidationResponse struct {
	Field   string `json:"field,omitempty"`   // Nama kolom yang bermasalah
	Message string `json:"message,omitempty"` // Pesan detail mengapa bermasalah
}

// ErrValidator adalah map/kamus untuk menampung pesan error buatan khusus (custom).
var ErrValidator = map[string]string{}

// ErrValidationResponse berfungsi untuk menerjemahkan error bawaan Go Validator menjadi format respon yang ramah dibaca (JSON).
func ErrValidationResponse(err error) (validationResponse []ValidationResponse) {
	var fieldErrors validator.ValidationErrors
	// Jika error yang masuk memang berwujud error dari paket validator...
	if errors.As(err, &fieldErrors) {
		// Mengulang (loop) setiap field yang memiliki error (karena error bisa lebih dari satu sekaligus)
		for _, err := range fieldErrors {
			// Mengecek jenis tag validasinya (contoh: apakah gara-gara aturan "required"?)
			switch err.Tag() {
			case "required":
				// Jika required, siapkan pesan wajib diisi.
				validationResponse = append(validationResponse, ValidationResponse{
					Field:   err.Field(),
					Message: fmt.Sprintf("%s is required", err.Field()),
				})
			case "email":
				// Jika tagnya "email", beri pesan bahwa email tidak valid
				validationResponse = append(validationResponse, ValidationResponse{
					Field:   err.Field(),
					Message: fmt.Sprintf("%s is not a valid email address", err.Field()),
				})
			case "oneof":
				// Jika tag "oneof" (harus salah satu dari opsi tertentu)
				validationResponse = append(validationResponse, ValidationResponse{
					Field:   err.Field(),
					Message: fmt.Sprintf("%s must be an oneof [%s]", err.Field(), err.Param()),
				})
			default:
				// Jika tidak ada di blok bawaan (default), cek ke daftar error custom yang udah didaftarkan (ErrValidator).
				errValidator, ok := ErrValidator[err.Tag()]
				if ok {
					count := strings.Count(errValidator, "%s")
					if count == 1 {
						validationResponse = append(validationResponse, ValidationResponse{
							Field:   err.Field(),
							Message: fmt.Sprintf(errValidator, err.Field()),
						})
					} else {
						validationResponse = append(validationResponse, ValidationResponse{
							Field:   err.Field(),
							Message: fmt.Sprintf(errValidator, err.Field(), err.Param()),
						})
					}
				} else {
					// Paling mentok jika tak diketahui masalahnya, tampilkan teks generic "something wrong".
					validationResponse = append(validationResponse, ValidationResponse{
						Field:   err.Field(),
						Message: fmt.Sprintf("something wrong on %s; %s", err.Field(), err.Tag()),
					})
				}
			}
		}
	}

	return validationResponse
}

func WrapError(err error) error {
	logrus.Errorf("error: %v", err)
	return err
}
