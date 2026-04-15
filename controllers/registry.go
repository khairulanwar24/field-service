package controllers

import (
	fieldConttroller "field-service/controllers/field"
	fieldScheduleController "field-service/controllers/fieldschedule"
	timeController "field-service/controllers/time"
	"field-service/services"
)

// Registry adalah tempat penyimpanan (wadah) yang meregistrasi semua controller yang ada di aplikasi ini.
// Analogi: Seperti kepala resepsionis yang punya kontak semua \"pelayan restoran\" (controller), Registry 
// ini mengumpulkan semua controller agar nantinya mudah dipanggil dan dikelola oleh router utama.
type Registry struct {
	service services.IServiceRegistry
}

// IControllerRegistry adalah kontrak (interface) yang memastikan bahwa semua controller 
// yang kita perlukan tersedia (bisa diambil). Ini seperti daftar menu layanan yang ditawarkan resepsionis.
type IControllerRegistry interface {
	GetField() fieldConttroller.IFieldController
	GetFieldSchedule() fieldScheduleController.IFieldScheduleController
	GetTime() timeController.ITimeController
}

// NewControllerRegistry adalah fungsi pembuat (constructor) untuk menginisialisasi Registry.
// Analogi: Seperti saat membuka restoran (menjalankan aplikasi), kita mempekerjakan 
// kepala resepsionis dan memberinya akses ke departemen belakang (service registry/dapur).
func NewControllerRegistry(service services.IServiceRegistry) IControllerRegistry {
	return &Registry{service: service}
}

// GetField mengambil dan mencetak (membuat) controller khusus untuk urusan Data Lapangan (Field).
func (r *Registry) GetField() fieldConttroller.IFieldController {
	return fieldConttroller.NewFieldController(r.service)
}

// GetFieldSchedule mengambil dan mencetak (membuat) controller khusus untuk urusan Jadwal Lapangan.
func (r *Registry) GetFieldSchedule() fieldScheduleController.IFieldScheduleController {
	return fieldScheduleController.NewFieldScheduleController(r.service)
}

// GetTime mengambil dan mencetak (membuat) controller khusus untuk urusan Waktu operasional.
func (r *Registry) GetTime() timeController.ITimeController {
	return timeController.NewTimeController(r.service)
}
