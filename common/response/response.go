package response

import (
	"field-service/constants"
	errConstant "field-service/constants/error"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response adalah struktur data (cetakan) untuk menentukan bentuk dari JSON yang akan dikirim ke client (Postman/Frontend).
type Response struct {
	// Status menunjukkan apakah request berhasil atau gagal (misal: "success" atau "error")
	Status string `json:"status"`
	// Message menampung pesan tambahan (bisa berupa teks String atau tipe data apa saja)
	Message any `json:"message"`
	// Data berisi isi dari response utamanya (seperti daftar user). interface{} berarti bisa menampung tipe apa saja.
	Data interface{} `json:"data"`
	// Token digunakan spesifik untuk mengembalikan sebuah JWT token (misal sehabis login). Sifatnya opsional (omitempty).
	Token *string `json:"token,omitempty"`
}

// ParamHTTPResp adalah tipe khusus (struktur parameter) yang dipakai saat memanggil fungsi HttpResponse.
type ParamHTTPResp struct {
	Code    int          // Kode HTTP Status (contoh: 200, 400, 500)
	Err     error        // Data error dari proses Go
	Message *string      // Pesan khusus (opsional)
	Gin     *gin.Context // Gin context untuk merespons ke client
	Data    interface{}  // Data utama yang akan dikembalikan
	Token   *string      // Token untuk autentikasi (opsional)
}

// HttpResponse adalah fungsi pusat (sentral) untuk membalas/merespons permintaan (request) klien dalam format JSON.
func HttpResponse(param ParamHTTPResp) {
	// 1. Jika tidak ada error sama sekali, berarti sukses.
	if param.Err == nil {
		// Menggunakan param.Gin.JSON untuk mengirim data berformat JSON kembali ke aplikasi peminta.
		param.Gin.JSON(param.Code, Response{
			Status:  constants.Success,              // Tulis status sebagai sukses
			Message: http.StatusText(http.StatusOK), // Teks standar dari angka 200 (yaitu "OK")
			Data:    param.Data,                     // Isi dengan data utamanya
			Token:   param.Token,                    // Isi dengan token
		})
		// Selesaikan/hentikan eksekusi dari fungsi HttpResponse karena sudah sukses merespon
		return
	}

	// 2. Kalau kode sampai sini, berarti ada sebuah ERROR. Kita siapkan pesan error umum "Kesalahan Server" secara bawaan.
	message := errConstant.ErrInternalServerError.Error()

	// 3. Jika programmer lewatkan teks pesan error khusus, maka timpa (overwrite) pesan umumnya.
	if param.Message != nil {
		message = *param.Message
	} else if param.Err != nil {
		// 4. Sebaliknya, gunakan fungsi pengecekan (ErrMapping) untuk melihat apakah tipe error yang terjadi
		// sudah terdaftar sistem (dikenali). Jika ya, tampilkan tulisan aslinya sesuai error dari Go.
		if errConstant.ErrMapping(param.Err) {
			message = param.Err.Error()
		}
	}

	// 5. Berikan respon balik berformat JSON dengan kode HTTP jelek (kesalahan).
	param.Gin.JSON(param.Code, Response{
		Status:  constants.Error, // Statusnya menjadi error (bukan success)
		Message: message,         // Pesan menyesuaikan logika di atas
		Data:    param.Data,      // Umumnya kosong (null) karena error
	})
	return
}
