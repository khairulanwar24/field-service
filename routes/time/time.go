package routes

import (
	"field-service/clients"
	"field-service/constants"
	"field-service/controllers"
	"field-service/middlewares"

	"github.com/gin-gonic/gin"
)

// TimeRoute adalah struktur data (struct) untuk menyimpan kebutuhan endpoint terkait waktu (Time).
// controller: untuk memanggil logika dari controller waktu.
// route: objek dari framework Gin untuk mengelompokkan URL/rute.
// client: untuk berkomunikasi dengan aplikasi/service eksternal (misal: memvalidasi layanan lain).
type TimeRoute struct {
	controller controllers.IControllerRegistry
	route      *gin.RouterGroup
	client     clients.IClientRegistry
}

// ITimeRoute adalah kontrak (interface) yang memastikan
// routing untuk operasi waktu memiliki fungsi Run() untuk mendaftarkan endpoint.
type ITimeRoute interface {
	Run()
}

// NewTimeRoute adalah fungsi pembuat (constructor) untuk instansiasi TimeRoute.
// Fungsi ini menyiapkan grup rute (awalan URL) "/time" pada framework Gin.
func NewTimeRoute(router *gin.Engine, controller controllers.IControllerRegistry, client clients.IClientRegistry) ITimeRoute {
	return &TimeRoute{
		controller: controller,
		route:      router.Group("/time"), // Mendaftarkan awalan rute menjadi "/time"
		client:     client,
	}
}

// Run adalah fungsi pelaksana yang akan mendaftarkan setiap rute (URL),
// beserta HTTP method (misal GET, POST) dan keamanannya untuk entitas Waktu (Time).
func (t *TimeRoute) Run() {
	// Membuat sub-grup rute. Nantinya alamat aslinya menjadi "/time/time".
	group := t.route.Group("/time")

	// Menggunakan middleware pengaman autentikasi ke semua rute di dalam sub-grup ini.
	// Artinya, semua request ke bawah baris ini WAJIB menyertakan token login (Bearer).
	group.Use(middlewares.Authenticate())

	// Endpoint GET ke "/time/time" untuk melihat/mengambil semua daftar ketersediaan waktu.
	// Pengecekan Hak Akses (CheckRole): Hanya diperbolehkan untuk akun berstatus "Admin".
	group.GET("", middlewares.CheckRole([]string{
		constants.Admin,
	}, t.client), t.controller.GetTime().GetAll)

	// Endpoint GET ke "/time/time/:uuid" untuk mencari satu detil waktu berdasarkan UUID di URL.
	// Hak akses: Hanya untuk akun "Admin".
	group.GET("/:uuid", middlewares.CheckRole([]string{
		constants.Admin,
	}, t.client), t.controller.GetTime().GetByUUID)

	// Endpoint POST ke "/time/time" untuk menambahkan pendaftaran data waktu yang baru.
	// Hak akses: Hanya dapat dilakukan oleh akun "Admin".
	group.POST("", middlewares.CheckRole([]string{
		constants.Admin,
	}, t.client), t.controller.GetTime().Create)
}
