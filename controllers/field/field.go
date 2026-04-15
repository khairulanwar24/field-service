package controllers

import (
	errValidation "field-service/common/error"
	"field-service/common/response"
	"field-service/domain/dto"
	"field-service/services"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"net/http"
)

// FieldController adalah struktur ("Pelayan") yang menangani permintaan terkait entitas Lapangan (Field).
// Analogi: Pelayan spesialis Lapangan ini akan menerima tiket/request dari pengunjung (misalnya mau lihat daftar, buat, atau hapus lapangan), 
// lalu membawanya ke Dapur (Service) agar diproses.
type FieldController struct {
	service services.IServiceRegistry
}

// IFieldController adalah daftar kemampuan yang WAJIB dimiliki pelayan lapangan ini.
// Dia wajib bisa melayani permintaan: ambil semua secara halam per halaman (pagination), ambil semua, ambil spesifik (UUID), buat baru (Create), ubah (Update), dan hapus (Delete).
type IFieldController interface {
	GetAllWithPagination(*gin.Context)
	GetAllWithoutPagination(*gin.Context)
	GetByUUID(*gin.Context)
	Create(*gin.Context)
	Update(*gin.Context)
	Delete(*gin.Context)
}

// NewFieldController adalah fungsi untuk merekrut (membuat) pelayan baru bagian Lapangan dan menghubungkannya ke Dapur (service).
func NewFieldController(service services.IServiceRegistry) IFieldController {
	return &FieldController{service: service}
}

// GetAllWithPagination mengambil daftar Lapangan dengan fitur "Halaman" (Pagination).
// Analogi: Seperti pengunjung yang minta, "Tolong tunjukkan daftar lapangan ke saya, tapi selembar maksimal 10 ya biar ga kepanjangan bacanya."
func (f *FieldController) GetAllWithPagination(c *gin.Context) {
	// 1. Pelayan siapkan kertas (struct params) khusus untuk mencatat catatan Query Pagination.
	var params dto.FieldRequestParam
	// 2. Pelayan mencatat pesan dari URL (contoh: ?page=1&limit=10) pakai ShouldBindQuery.
	err := c.ShouldBindQuery(&params)
	if err != nil {
		// Jika format yang diketik di URL ngawur (Contoh huruf dicampur angka di kolom limit), tolak!
		response.HttpResponse(response.ParamHTTPResp{
			Code: http.StatusBadRequest,
			Err:  err,
			Gin:  c,
		})
		return
	}

	// 3. Inspektur/Validator mengecek kertas tadi, apakah parameter yang dimasukkan sah atau melanggar aturan.
	validate := validator.New()
	err = validate.Struct(params)
	if err != nil {
		// Jika melanggar/kurang lengkap, kembalikan ke klien dengan status 422 (UnprocessableEntity).
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

	// 4. Syarat beres, serahkan ke Dapur (Service) untuk dicari dan dihitung halamannya.
	result, err := f.service.GetField().GetAllWithPagination(c, &params)
	if err != nil {
		// Jika Dapur (misal database) error
		response.HttpResponse(response.ParamHTTPResp{
			Code: http.StatusBadRequest,
			Err:  err,
			Gin:  c,
		})
		return
	}

	// 5. Sukses! Ini hasilnya, silakan dinikmati.
	response.HttpResponse(response.ParamHTTPResp{
		Code: http.StatusOK,
		Data: result,
		Gin:  c,
	})
}

// GetAllWithoutPagination mengambil SEMUA daftar lapangan sekaligus tanpa dipotong-potong halaman.
// Analogi: "Pelayan, kasih saya seluruh menu sekaligus!"
func (f *FieldController) GetAllWithoutPagination(c *gin.Context) {
	// 1. Pelayan langsung suruh dapur mengambil semuanya.
	result, err := f.service.GetField().GetAllWithoutPagination(c)
	if err != nil {
		// Kalau gagal, laporkan kembali.
		response.HttpResponse(response.ParamHTTPResp{
			Code: http.StatusBadRequest,
			Err:  err,
			Gin:  c,
		})
		return
	}

	// Berhasil
	response.HttpResponse(response.ParamHTTPResp{
		Code: http.StatusOK,
		Data: result,
		Gin:  c,
	})
}

// GetByUUID mengambil satu buah Lapangan yang spesifik berdasarkan nomor unik (UUID)-nya.
func (f *FieldController) GetByUUID(c *gin.Context) {
	// Pelayan mencatat UUID dari URL (c.Param), lalu meminta data dari Dapur (service).
	result, err := f.service.GetField().GetByUUID(c, c.Param("uuid"))
	if err != nil {
		// Jika datanya tidak ada
		response.HttpResponse(response.ParamHTTPResp{
			Code: http.StatusBadRequest,
			Err:  err,
			Gin:  c,
		})
		return
	}

	// Sukses ditemukan
	response.HttpResponse(response.ParamHTTPResp{
		Code: http.StatusOK,
		Data: result,
		Gin:  c,
	})
}

// Create berguna untuk membuat data Lapangan baru. Endpoint ini sering menerima fitur unggah foto/berkas.
// Makanya dia menggunakan binding.FormMultipart yang bisa menampung data form beserta file (foto lapangan).
func (f *FieldController) Create(c *gin.Context) {
	// 1. Siapkan catatan pesanan.
	var request dto.FieldRequest
	// 2. Baca pesanan dalam format "FormMultipart" (artinya campuran teks dan file foto/berkas).
	err := c.ShouldBindWith(&request, binding.FormMultipart)
	if err != nil {
		// Kalau gagal baca formatnya, tolak.
		response.HttpResponse(response.ParamHTTPResp{
			Code: http.StatusBadRequest,
			Err:  err,
			Gin:  c,
		})
		return
	}

	// 3. Inspeksi form ini, apa nama lapangannya sudah diisi? apa harganya valid?
	validate := validator.New()
	if err = validate.Struct(request); err != nil {
		// Kalau isian kurang/salah
		errMessage := http.StatusText(http.StatusUnprocessableEntity)
		errorResponse := errValidation.ErrValidationResponse(err)
		response.HttpResponse(response.ParamHTTPResp{
			Err:     err,
			Code:    http.StatusUnprocessableEntity,
			Message: &errMessage,
			Data:    errorResponse,
			Gin:     c,
		})
		return
	}

	// 4. Form valid, minta Dapur (service) menyimpan data sekaligus memproses foto tersebut.
	result, err := f.service.GetField().Create(c, &request)
	if err != nil {
		response.HttpResponse(response.ParamHTTPResp{
			Code: http.StatusBadRequest,
			Err:  err,
			Gin:  c,
		})
		return
	}

	// 5. Berhasil (Created).
	response.HttpResponse(response.ParamHTTPResp{
		Code: http.StatusCreated,
		Data: result,
		Err:  err,
		Gin:  c,
	})
}

// Update berguna untuk merevisi atau memperbarui data lapangan. Sama seperti Create, ia bisa mengupdate foto.
func (f *FieldController) Update(c *gin.Context) {
	var request dto.UpdateFieldRequest
	// 1. Ambil data perubahannya (teks & file)
	err := c.ShouldBindWith(&request, binding.FormMultipart)
	if err != nil {
		response.HttpResponse(response.ParamHTTPResp{
			Code: http.StatusBadRequest,
			Err:  err,
			Gin:  c,
		})
		return
	}

	// 2. Validasi perubahannya
	validate := validator.New()
	err = validate.Struct(request)
	if err != nil {
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

	// 3. Eksekusi perubahannya di Dapur (Service) berdasarkan spesifik UUID yang diberi di URL.
	result, err := f.service.GetField().Update(c, c.Param("uuid"), &request)
	if err != nil {
		response.HttpResponse(response.ParamHTTPResp{
			Code: http.StatusBadRequest,
			Err:  err,
			Gin:  c,
		})
		return
	}

	// 4. Sukses
	response.HttpResponse(response.ParamHTTPResp{
		Code: http.StatusOK,
		Data: result,
		Gin:  c,
	})
}

// Delete berguna untuk menghapus lapangan yang tidak dipakai lagi.
func (f *FieldController) Delete(c *gin.Context) {
	// Langsung minta service menghapus berdasarkan ID yang ada di URL (Param)
	err := f.service.GetField().Delete(c, c.Param("uuid"))
	if err != nil {
		// Kalau gagal waktu menghapus
		response.HttpResponse(response.ParamHTTPResp{
			Code: http.StatusBadRequest,
			Err:  err,
			Gin:  c,
		})
		return
	}

	// Sukses menghapus
	response.HttpResponse(response.ParamHTTPResp{
		Code: http.StatusOK,
		Gin:  c,
	})
}
