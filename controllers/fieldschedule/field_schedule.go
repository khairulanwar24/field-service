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

// FiledScheduleController adalah Pelayan yang khusus menangani masalah "Jadwal Lapangan".
// Analogi: Jika ada pengunjung yang mau pesan/booking di jam-jam tertentu, pelayan inilah yang 
// mengecek ketersediaannya dan menyampaikannya ke Dapur/Sistem (Service).
type FiledScheduleController struct {
	service services.IServiceRegistry
}

// IFieldScheduleController adalah daftar pelayanan (Skill) dari Pelayan Jadwal ini.
type IFieldScheduleController interface {
	GetAllWithPagination(*gin.Context)
	GetAllByFieldIDAndDate(*gin.Context)
	GetByUUID(*gin.Context)
	Create(*gin.Context)
	Update(*gin.Context)
	UpdateStatus(*gin.Context)
	Delete(*gin.Context)
	GenerateScheduleForOneMonth(*gin.Context)
}

func NewFieldScheduleController(service services.IServiceRegistry) IFieldScheduleController {
	return &FiledScheduleController{service: service}
}

func (f *FiledScheduleController) GetAllWithPagination(c *gin.Context) {
	var params dto.FieldScheduleRequestParam
	err := c.ShouldBindQuery(&params)
	if err != nil {
		response.HttpResponse(response.ParamHTTPResp{
			Code: http.StatusBadRequest,
			Err:  err,
			Gin:  c,
		})
		return
	}

	validate := validator.New()
	err = validate.Struct(params)
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

	result, err := f.service.GetFieldSchedule().GetAllWithPagination(c, &params)
	if err != nil {
		response.HttpResponse(response.ParamHTTPResp{
			Code: http.StatusBadRequest,
			Err:  err,
			Gin:  c,
		})
		return
	}

	response.HttpResponse(response.ParamHTTPResp{
		Code: http.StatusOK,
		Data: result,
		Gin:  c,
	})
}

func (f *FiledScheduleController) GetAllByFieldIDAndDate(c *gin.Context) {
	// endpoint ini khusus berguna ketika user mau filter "saya mau lihat jadwal di lapangan A khusus tanggal 1 Januari."
	var params dto.FieldScheduleByFieldIDAndDateRequestParam

	// 1. Ambil informasi tanggal dari URL (query)
	err := c.ShouldBindQuery(&params)
	if err != nil {
		response.HttpResponse(response.ParamHTTPResp{
			Code: http.StatusBadRequest,
			Err:  err,
			Gin:  c,
		})
		return
	}

	// 2. Validasi formatnya (apakah format tanggalnya beneran valid atau cuma tulisan abal-abal)
	validate := validator.New()
	err = validate.Struct(params)
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

	// 3. Minta data ke service, kasih UUID lapangannya dari param /:uuid 
	// dan tanggal dari query ?date=...
	result, err := f.service.GetFieldSchedule().GetAllByFieldIDAndDate(c, c.Param("uuid"), params.Date)
	if err != nil {
		response.HttpResponse(response.ParamHTTPResp{
			Code: http.StatusBadRequest,
			Err:  err,
			Gin:  c,
		})
		return
	}

	// 4. Jadwalnya berhasil ditarik, lalu kirim kembali response-nya.
	response.HttpResponse(response.ParamHTTPResp{
		Code: http.StatusOK,
		Data: result,
		Gin:  c,
	})
}

func (f *FiledScheduleController) GetByUUID(c *gin.Context) {
	result, err := f.service.GetFieldSchedule().GetByUUID(c, c.Param("uuid"))
	if err != nil {
		response.HttpResponse(response.ParamHTTPResp{
			Code: http.StatusBadRequest,
			Err:  err,
			Gin:  c,
		})
		return
	}

	response.HttpResponse(response.ParamHTTPResp{
		Code: http.StatusOK,
		Data: result,
		Gin:  c,
	})
}

func (f *FiledScheduleController) Create(c *gin.Context) {
	var params dto.FieldScheduleRequest
	err := c.ShouldBindJSON(&params)
	if err != nil {
		response.HttpResponse(response.ParamHTTPResp{
			Code: http.StatusBadRequest,
			Err:  err,
			Gin:  c,
		})
		return
	}

	validate := validator.New()
	err = validate.Struct(params)
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

	err = f.service.GetFieldSchedule().Create(c, &params)
	if err != nil {
		response.HttpResponse(response.ParamHTTPResp{
			Code: http.StatusBadRequest,
			Err:  err,
			Gin:  c,
		})
		return
	}

	response.HttpResponse(response.ParamHTTPResp{
		Code: http.StatusCreated,
		Gin:  c,
	})
}

// GenerateScheduleForOneMonth adalah fitur unik di mana aplikasi bisa secara otomatis membuatkan slot-jadwal selama 1 bulan penuh.
// Analogi: Pelayan mendapatkan alat canggih (otomasi), "Tolong buatkan kalender kosong untuk dipesan di Lapangan A selama 30 hari ke depan!"
func (f *FiledScheduleController) GenerateScheduleForOneMonth(c *gin.Context) {
	var params dto.GenerateFieldScheduleForOneMonthRequest
	err := c.ShouldBindJSON(&params) // Request berbentuk data JSON.
	if err != nil {
		response.HttpResponse(response.ParamHTTPResp{
			Code: http.StatusBadRequest,
			Err:  err,
			Gin:  c,
		})
		return
	}

	// Validasi apakah properti yang diperlukan lengkap.
	validate := validator.New()
	err = validate.Struct(params)
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

	// Eksekusi operasi berat (membuat banyak jadwal) ke service.
	err = f.service.GetFieldSchedule().GenerateScheduleForOneMonth(c, &params)
	if err != nil {
		response.HttpResponse(response.ParamHTTPResp{
			Code: http.StatusBadRequest,
			Err:  err,
			Gin:  c,
		})
		return
	}

	// Pembuatan massal berhasil (Created).
	response.HttpResponse(response.ParamHTTPResp{
		Code: http.StatusCreated,
		Gin:  c,
	})
}

func (f *FiledScheduleController) Update(c *gin.Context) {
	var params dto.UpdateFieldScheduleRequest
	err := c.ShouldBindJSON(&params)
	if err != nil {
		response.HttpResponse(response.ParamHTTPResp{
			Code: http.StatusBadRequest,
			Err:  err,
			Gin:  c,
		})
		return
	}

	validate := validator.New()
	err = validate.Struct(params)
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

	result, err := f.service.GetFieldSchedule().Update(c, c.Param("uuid"), &params)
	if err != nil {
		response.HttpResponse(response.ParamHTTPResp{
			Code: http.StatusBadRequest,
			Err:  err,
			Gin:  c,
		})
		return
	}

	response.HttpResponse(response.ParamHTTPResp{
		Code: http.StatusOK,
		Gin:  c,
		Data: result,
	})
}

// UpdateStatus berfungsi untuk mengganti status sebuah jadwal, misalnya jadwal yang "Available" (Kosong) menjadi "Booked" (Dipesan).
// Analogi: Pelayang menghitamkan stiker di daftar jadwal/halaman pesanan agar tamu lain tidak mengambil jam yang sama.
func (f *FiledScheduleController) UpdateStatus(c *gin.Context) {
	var request dto.UpdateStatusFieldScheduleRequest
	err := c.ShouldBindJSON(&request)
	if err != nil {
		response.HttpResponse(response.ParamHTTPResp{
			Code: http.StatusBadRequest,
			Err:  err,
			Gin:  c,
		})
		return
	}

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

	err = f.service.GetFieldSchedule().UpdateStatus(c, &request)
	if err != nil {
		response.HttpResponse(response.ParamHTTPResp{
			Code: http.StatusBadRequest,
			Err:  err,
			Gin:  c,
		})
		return
	}

	response.HttpResponse(response.ParamHTTPResp{
		Code: http.StatusOK,
		Gin:  c,
	})
}

func (f *FiledScheduleController) Delete(c *gin.Context) {
	err := f.service.GetFieldSchedule().Delete(c, c.Param("uuid"))
	if err != nil {
		response.HttpResponse(response.ParamHTTPResp{
			Code: http.StatusBadRequest,
			Err:  err,
			Gin:  c,
		})
		return
	}

	response.HttpResponse(response.ParamHTTPResp{
		Code: http.StatusOK,
		Gin:  c,
	})
}
