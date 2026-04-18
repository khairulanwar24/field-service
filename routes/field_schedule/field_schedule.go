package routes

import (
	"field-service/clients"
	"field-service/constants"
	"field-service/controllers"
	"field-service/middlewares"

	"github.com/gin-gonic/gin"
)

// FieldScheduleRoute adalah struktur data yang menyimpan dependensi khusus untuk routing jadwal lapangan.
// controller: objek untuk memanggil fungsi-fungsi logika (handler) terkait jadwal dari controller.
// group: objek dari Gin framework untuk mengelompokkan path API (rute).
// client: objek untuk melakukan komunikasi dengan aplikasi service lain.
type FieldScheduleRoute struct {
	controller controllers.IControllerRegistry
	group      *gin.RouterGroup
	client     clients.IClientRegistry
}

// IFieldScheduleRoute adalah kontrak antar-muka (interface) yang wajib dimiliki,
// yaitu fungsi Run() untuk mendaftarkan seluruh rute (endpoints) jadwal lapangan ke dalam aplikasi.
type IFieldScheduleRoute interface {
	Run()
}

// NewFieldScheduleRoute adalah konstruktor (pembuat objek) dari FieldScheduleRoute.
// Fungsi ini menyiapkan rute dasar (prefix) "/field" agar dapat diproses lebih lanjut oleh fungsi Run.
func NewFieldScheduleRoute(router *gin.Engine, controller controllers.IControllerRegistry, client clients.IClientRegistry) IFieldScheduleRoute {
	return &FieldScheduleRoute{
		controller: controller,
		group:      router.Group("/field"), // Mengelompokkan rute URL berawalan "/field"
		client:     client,
	}
}

// Run adalah fungsi yang mendaftarkan seluruh rute (endpoint) HTTP untuk entitas Jadwal Lapangan (Field Schedule).
func (f *FieldScheduleRoute) Run() {
	// Membuat sub-grup rute. Hasil akhirnya rute ini akan menjadi "/field/schedule".
	group := f.group.Group("/field/schedule")

	// Endpoint GET ke "/field/schedule" untuk mengambil semua jadwal lapangan berdasarkan ID Lapangan dan Tanggal.
	// Middleware AuthenticateWithoutToken: Boleh diakses tanpa token login (publik).
	group.GET("", middlewares.AuthenticateWithoutToken(), f.controller.GetFieldSchedule().GetAllByFieldIDAndDate)

	// Endpoint PATCH ke "/field/schedule" untuk memperbarui status pada suatu jadwal.
	// Akses ini diatur publik, namun pastikan logika di dalam controllernya tetap aman.
	group.PATCH("", middlewares.AuthenticateWithoutToken(), f.controller.GetFieldSchedule().UpdateStatus)

	// Mulai dari baris ini ke bawah, setiap request WAJIB menyertakan token yang valid.
	// Group.Use akan menerapkan fungsi pengaman (middleware) ke rute-rute di bawahnya.
	group.Use(middlewares.Authenticate())

	// Endpoint GET ke "/field/schedule/pagination" untuk mengambil data jadwal dalam bentuk per halaman (paginasi).
	// Pengecekan hak akses (CheckRole): Endpoint ini khusus dijalankan oleh role Admin atau Customer.
	group.GET("/pagination", middlewares.CheckRole([]string{
		constants.Admin,
		constants.Customer,
	}, f.client), f.controller.GetFieldSchedule().GetAllWithPagination)

	// Endpoint POST ke "/field/schedule" untuk menambah data jadwal lapangan baru secara manual.
	// Hak akses spesifik: Hanya diperbolehkan untuk akun dengan role Admin.
	group.POST("", middlewares.CheckRole([]string{
		constants.Admin,
	}, f.client), f.controller.GetFieldSchedule().Create)

	// Endpoint POST tambahan (ke rute yang sama) untuk memicu pembuatan otomatis jadwal selama 1 bulan.
	// *Catatan: Mungkin perlu penyesuaian path (misal: "/one-month") agar tidak bentrok dengan POST Create di atas.
	// Hak akses spesifik: Hanya diperbolehkan untuk akun dengan role Admin.
	group.POST("", middlewares.CheckRole([]string{
		constants.Admin,
	}, f.client), f.controller.GetFieldSchedule().GenerateScheduleForOneMonth)

	// Endpoint PUT ke "/field/schedule/:uuid" untuk mengubah atau memperbarui informasi jadwal secara utuh.
	// Hak akses spesifik: Hanya diperbolehkan untuk akun dengan role Admin.
	group.PUT("/:uuid", middlewares.CheckRole([]string{
		constants.Admin,
	}, f.client), f.controller.GetFieldSchedule().Update)

	// Endpoint DELETE ke "/field/schedule/:uuid" untuk menghapus data jadwal berdasarkan UUID.
	// Hak akses spesifik: Hanya diperbolehkan untuk akun dengan role Admin.
	group.DELETE("/:uuid", middlewares.CheckRole([]string{
		constants.Admin,
	}, f.client), f.controller.GetFieldSchedule().Delete)
}
