package routes

import (
	"field-service/clients"
	"field-service/constants"
	"field-service/controllers"
	"field-service/middlewares"

	"github.com/gin-gonic/gin"
)

// FieldRoute adalah struktur data yang menyimpan dependensi untuk routing field.
// controller: objek untuk memanggil fungsi-fungsi logika (handler) dari controller.
// group: objek dari Gin framework untuk mengelompokkan path API (rute).
// client: objek untuk melakukan komunikasi dengan aplikasi service lain.
type FieldScheduleRoute struct {
	controller controllers.IControllerRegistry
	group      *gin.RouterGroup
	client     clients.IClientRegistry
}

// IFieldRoute adalah kontrak antar-muka (interface) yang hanya memiliki satu fungsi,
// yaitu Run() untuk mendaftarkan rute-rute (endpoints) ke dalam aplikasi.
type IFieldScheduleRoute interface {
	Run()
}

// NewFieldRoute adalah konstruktor (pembuat objek) dari FieldRoute.
// Fungsi ini mengelompokkan semua rute terkait field ke dalam prefix "/field".
func NewFieldScheduleRoute(controller controllers.IControllerRegistry, group *gin.RouterGroup, client clients.IClientRegistry) IFieldScheduleRoute {
	return &FieldScheduleRoute{
		controller: controller,
		group:      group, // Mengelompokkan rute URL berawalan "/field"
		client:     client,
	}
}

// Run adalah fungsi yang mendaftarkan seluruh rute (endpoint) HTTP untuk entitas Field (Lapangan).
func (f *FieldScheduleRoute) Run() {
	// Membuat sub-grup rute. Hasil akhirnya rute ini akan menjadi "/field/field".
	group := f.group.Group("/field/schedule")

	// Endpoint GET ke "/field/field" untuk mengambil semua data lapangan (tanpa penomoran halaman / paginasi).
	// Middleware AuthenticateWithoutToken: Boleh diakses meski tanpa token login (publik).
	group.GET("", middlewares.AuthenticateWithoutToken(), f.controller.GetFieldSchedule().GetAllByFieldIDAndDate)

	group.PATCH("", middlewares.AuthenticateWithoutToken(), f.controller.GetFieldSchedule().UpdateStatus)

	// Mulai dari baris ini ke bawah, kita wajib menggunakan token valid.
	// Group.Use akan menerapkan fungsi (middleware) ke rute di bawahnya.
	// middlewares.Authenticate() memvalidasi apakah ada token (bearer) dan apakah valid.
	group.Use(middlewares.Authenticate())

	// Endpoint GET ke "/field/field/pagination" untuk mengambil data lapangan yang terbagi-bagi per halaman (paginasi).
	// Memiliki pengecekan hak akses (CheckRole): Endpoint ini hanya dapat dijalankan oleh role Admin atau Customer.
	group.GET("/pagination", middlewares.CheckRole([]string{
		constants.Admin,
		constants.Customer,
	}, f.client), f.controller.GetFieldSchedule().GetAllWithPagination)

	group.GET("/:uuid", middlewares.CheckRole([]string{
		constants.Admin,
		constants.Customer,
	}, f.client), f.controller.GetFieldSchedule().GetByUUID)

	// Endpoint POST ke "/field/field" untuk menyimpan data lapangan baru.
	// Hak akses spesifik: Hanya diperbolehkan untuk akun ber-role Admin.
	group.POST("", middlewares.CheckRole([]string{
		constants.Admin,
	}, f.client), f.controller.GetFieldSchedule().Create)

	group.POST("/one-month", middlewares.CheckRole([]string{
		constants.Admin,
	}, f.client), f.controller.GetFieldSchedule().GenerateScheduleForOneMonth)

	// Endpoint PUT ke "/field/field/:uuid" untuk mengubah atau update data lapangan berdasarkan UUID.
	// Hak akses spesifik: Hanya diperbolehkan untuk akun ber-role Admin.
	group.PUT("/:uuid", middlewares.CheckRole([]string{
		constants.Admin,
	}, f.client), f.controller.GetFieldSchedule().Update)

	// Endpoint DELETE ke "/field/field/:uuid" untuk menghapus data lapangan berdasarkan UUID.
	// Hak akses spesifik: Hanya diperbolehkan untuk akun ber-role Admin.
	group.DELETE("/:uuid", middlewares.CheckRole([]string{
		constants.Admin,
	}, f.client), f.controller.GetFieldSchedule().Delete)
}
