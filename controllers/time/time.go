package controllers

import (
	errValidation "field-service/common/error"
	"field-service/common/response"
	"field-service/domain/dto"
	"field-service/services"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"net/http"
)

// TimeController adalah struktur utama yang menangani permintaan (request) klien terkait entitas Waktu.
// Analogi: Controller ini seperti PELAYAN RESTORAN. 
// Pelayan menerima pesanan dari tamu, mengecek apakah pesanan bisa dibaca, lalu memberikan pesanan ke DAPUR (Service).
// Pelayan ini menyimpan referensi ke departemen dapur melalui properti "service".
type TimeController struct {
	service services.IServiceRegistry
}

// ITimeController adalah kontrak yang mendefinisikan daftar keahlian si Pelayan (TimeController).
// Pelayan ini wajib tahu dan bisa menjalankan tugas: ambil semua data (GetAll), ambil satu spesifik (GetByUUID), dan buat data baru (Create).
type ITimeController interface {
	GetAll(*gin.Context)
	GetByUUID(*gin.Context)
	Create(*gin.Context)
}

// NewTimeController adalah fungsi pembuat (constructor).
// Ini adalah saat kita mempekerjakan "Pelayan" khusus Waktu, dan memberitahunya siapa yang ada di "Dapur" (Service Registry),
// sehingga si Pelayan nanti tahu ke mana harus meneruskan instruksi dari pelanggan (client).
func NewTimeController(service services.IServiceRegistry) ITimeController {
	return &TimeController{service: service}
}

// GetAll berfungsi untuk mengambil semua data Waktu dari database.
// Analogi: Pelanggan (klien) meminta, "Tolong tunjukkan semua daftar jam operasional yang kita punya!"
func (t *TimeController) GetAll(c *gin.Context) {
	// 1. Pelayan meneruskan pesanan ini ke Dapur (Service GetTime) untuk dicarikan/disiapkan datanya.
	result, err := t.service.GetTime().GetAll(c)
	if err != nil {
		// 2a. Jika dapur mengabarkan ada kegagalan, pelayan kembali ke pengunjung dan bilang ada yang salah (BadRequest).
		response.HttpResponse(response.ParamHTTPResp{
			Code: http.StatusBadRequest,
			Err:  err,
			Gin:  c,
		})
		return
	}

	// 2b. Jika sukses, pelayan menyajikan hidangan data kepada pelanggan sambil tersenyum (Status OK - 200).
	response.HttpResponse(response.ParamHTTPResp{
		Code: http.StatusOK,
		Data: result,
		Gin:  c,
	})
}

// GetByUUID berfungsi mengambil satu buah data Waktu yang sangat spesifik berdasarkan ID (UUID).
// Analogi: Pelanggan minta, "Tolong bawakan pesanan saya yang nomor referensinya 1234!"
func (t *TimeController) GetByUUID(c *gin.Context) {
	// 1. Pelayan mencatat nomor ID yang dicari dari URL browser/request pelanggan.
	uuid := c.Param("uuid")
	// 2. Pelayan pergi menuju dapur dan memberikan nomor spesifik tersebut.
	result, err := t.service.GetTime().GetByUUID(c, uuid)
	if err != nil {
		// 3a. Kalau ternyata ID/nomornya ga ketemu di database atau dapur bermasalah, sampaikan kemari.
		response.HttpResponse(response.ParamHTTPResp{
			Code: http.StatusBadRequest,
			Err:  err,
			Gin:  c,
		})
		return
	}

	// 3b. Makanan (pesanan spesifik) sudah siap, sajikan kembali (Response OK).
	response.HttpResponse(response.ParamHTTPResp{
		Code: http.StatusOK,
		Data: result,
		Gin:  c,
	})
}

// Create bertanggung jawab menerima data dari pengguna untuk disimpan (membuat waktu operasional baru).
// Analogi: Pelanggan minta tambah menu baru. Pelayan harus periksa tulisannya bisa dibaca tidak, dan bahan-bahannya lengkap atau tidak.
func (t *TimeController) Create(c *gin.Context) {
	// 1. Pelayan menyiapkan kertas kosong (variabel request berbentuk struct) untuk menulis pesanan klien.
	var request dto.TimeRequest
	// 2. Membaca isi pesanan JSON (ShouldBindJSON) dari request pelanggan untuk masuk ke "request".
	err := c.ShouldBindJSON(&request)
	if err != nil {
		// 2a. Kalau format bentuk pesanan salah (JSON tidak valid), langsung tolak tanpa ke dapur! (BadRequest).
		response.HttpResponse(response.ParamHTTPResp{
			Code: http.StatusBadRequest,
			Err:  err,
			Gin:  c,
		})
		return
	}

	// 3. Pesanan bisa dibaca, TAPI isinya wajar ngga? (Contoh: apakah jam-nya benar diisi).
	// Kita butuh alat pemeriksa (Validator).
	validate := validator.New()
	err = validate.Struct(request)
	if err != nil {
		// 3a. Kalau ada bagian form pesanan yang kurang lengkap sesuai wajib isi, kembalikan daftar kekurangannya ke client.
		errMessage := http.StatusText(http.StatusUnprocessableEntity)
		errorResponse := errValidation.ErrValidationResponse(err)
		response.HttpResponse(response.ParamHTTPResp{
			Code:    http.StatusBadRequest,
			Err:     err,
			Message: &errMessage,
			Data:    errorResponse,
			Gin:     c,
		})
		return
	}

	// 4. Semua pengecekan selesai dan pesanan valid. Pelayan melempar pesanan matang ke Dapur (Service) agar dieksekusi masuk Database.
	result, err := t.service.GetTime().Create(c, &request)
	if err != nil {
		// Jika ketika Dapur memasaknya ternyata error / gagal dalam sistem bisnis.
		response.HttpResponse(response.ParamHTTPResp{
			Code: http.StatusBadRequest,
			Err:  err,
			Gin:  c,
		})
		return
	}

	// 5. Berhasil ditambahkan! Kita ucapkan selamat datanya ter-Created (Status 201).
	response.HttpResponse(response.ParamHTTPResp{
		Code: http.StatusCreated,
		Data: result,
		Gin:  c,
	})
}
