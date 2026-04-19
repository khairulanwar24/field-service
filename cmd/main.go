package cmd

import (
	"encoding/base64"
	"field-service/clients"
	"field-service/common/gcs"
	"field-service/common/response"
	"field-service/config"
	"field-service/constants"
	"field-service/controllers"
	"field-service/domain/models"
	"field-service/middlewares"
	"field-service/repositories"
	"field-service/routes"
	"field-service/services"
	"fmt"
	"net/http"
	"time"

	"github.com/didip/tollbooth"
	"github.com/didip/tollbooth/limiter"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

// command adalah instance dari library Cobra. Library ini sering digunakan
// untuk membuat aplikasi berbasis CLI (Command Line Interface).
// Di sini, kita mendefinisikan sebuah perintah bernama "serve" yang berfungsi
// untuk menyalakan/menjalankan server aplikasi kita.
var command = &cobra.Command{
	Use:   "serve",
	Short: "Start the server",
	// Fungsi Run akan dijalankan saat command "serve" dipanggil melalui terminal.
	Run: func(c *cobra.Command, args []string) {
		// godotenv.Load() akan membaca file .env untuk memuat variabel environment
		// (seperti konfigurasi port, kredensial database, dsb) ke dalam aplikasi.
		_ = godotenv.Load()
		// Inisialisasi/menyiapkan konfigurasi agar siap digunakan oleh aplikasi (membaca variabe di file .env).
		config.Init()
		
		// Inisialisasi koneksi ke Database (misal: MySQL/PostgreSQL).
		db, err := config.InitDatabase()
		if err != nil {
			// panic akan mematikan aplikasi secara spontan/langsung jika DB gagal terkoneksi.
			panic(err)
		}

		// Karena Go (Golang) secara bawaan menggunakan UTC (waktu dunia),
		// secara manual zona waktu global aplikasi diubah menjadi zona Asia/Jakarta (WIB).
		loc, err := time.LoadLocation("Asia/Jakarta")
		if err != nil {
			panic(err)
		}
		time.Local = loc

		// AutoMigrate adalah fitur canggih dari GORM untuk mencocokkan struktur/tabel database
		// secara otomatis berdasarkan struktur (struct) Models yang kita buat di kode.
		err = db.AutoMigrate(
			&models.Field{},
			&models.FieldSchedule{},
			&models.Time{},
		)
		if err != nil {
			panic(err)
		}

		// --- Inisialisasi Lapisan-Lapisan Arsitektur (Dependency Injection) ---
		// 1. GCS: Google Cloud Storage untuk media upload (gambar dsb).
		gcs := initGCS()
		// 2. Client: layer untuk melakukan panggilan API (request) ke microservice server lain.
		client := clients.NewClientRegistry()
		// 3. Repository: layer yang isinya semua Query (SQL) untuk berinteraksi langsung ke Database.
		repository := repositories.NewRepositoryRegistry(db)
		// 4. Service: layer (Logic Bisnis) - otak program di mana semua logika validasi dirangkai.
		service := services.NewServiceRegistry(repository, gcs)
		// 5. Controller: layer terluar untuk membaca Request dari user dan membalas dalam wujud JSON.
		controller := controllers.NewControllerRegistry(service)

		// Membuat instance engine web baru dari framework Gin.
		router := gin.Default()

		// Menambahkan middleware global penahan "panic" agar server/aplikasi tidak langsung mati (crash/berhenti) jika terjadi bug code.
		router.Use(middlewares.HandlePanic())

		// Menentukan response dasar jika pengguna (client) mencoba akses endpoint yang tidak pernah kita buat (URL asal-asalan).
		router.NoRoute(func(c *gin.Context) {
			c.JSON(http.StatusNotFound, response.Response{
				Status:  constants.Error,
				Message: fmt.Sprintf("Path %s", http.StatusText(http.StatusNotFound)),
			})
		})

		// Endpoint ROOT bawaan (Health-Check) "/", sangat berguna untuk memastikan bahwa web server kita sudah hidup ketika di-deploy.
		router.GET("/", func(c *gin.Context) {
			c.JSON(http.StatusOK, response.Response{
				Status:  constants.Success,
				Message: "Welcome to Field Service",
			})
		})

		// Mengkustomisasi Middleware CORS (Cross-Origin Resource Sharing).
		// Sangat penting agar aplikasi (frontend browser) di web berbeda dapat mengakses API ini.
		router.Use(func(c *gin.Context) {
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*") // * artinya mengizinkan siapapun
			c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH")
			c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, x-service-name, x-request-at, x-api-key")
			
			// Jika request merupakan pengecekan (pre-flight check) jenis OPTIONS, langsung hentikan proses dengan pesan sukses.
			if c.Request.Method == "OPTIONS" {
				c.AbortWithStatus(204)
				return
			}
			c.Next()
		})

		// Pengaturan Rate Limiter (Pembatas Kecepatan / Lalu Lintas HTTP).
		// Melindungi server kita kebebanan dari spam/request secara beruntun dan liar.
		lmt := tollbooth.NewLimiter(
			config.Config.RateLimiterMaxRequest,
			&limiter.ExpirableOptions{
				DefaultExpirationTTL: time.Duration(config.Config.RateLimiterTimeSecond) * time.Second,
			})
		router.Use(middlewares.RateLimiter(lmt))

		// Mendaftarkan versi API kita (berawalan URL: /api/v1).
		// Lalu mengirimkan controller client ke registry rute-rute di mana endpoint spesifik seperti field akan didaftarkan.
		group := router.Group("/api/v1")
		route := routes.NewRouteRegistry(controller, group, client)
		route.Serve()

		port := fmt.Sprintf(":%d", config.Config.Port)
		// Memulai / Mengerjakan server Gin di port yang sudah disediakan konfigurasi.
		router.Run(port)
	},
}

// Run (di luar struktur cobra di atas) adalah punggawa utama.
// Fungsi utilitas ini nantinya akan dipanggil dari file main.go root aplikasi.
func Run() {
	// Menjalankan instruksi CLI Cobra dari variabel command "serve" yang terdaftarkan di atas
	err := command.Execute()
	if err != nil {
		panic(err)
	}
}

// initGCS adalah fungsi khusus untuk membaca kredensial akun rahasia GCS (Google Cloud Storage)
// yang digunakan untuk mengunggah dan menampilkan file secara online.
func initGCS() gcs.IGCSClient {
	// Mengurai/menguraikan bahasa rahasia (private key) GCS yang tadinya disembunyikan dalam bentuk "Base64"
	decode, err := base64.StdEncoding.DecodeString(config.Config.GCSPrivateKey)
	if err != nil {
		panic(err)
	}

	stringPrivateKey := string(decode)
	// Membuat representasi Json Token otentikasi akun service GCP berlandaskan .env
	gcsServiceAccount := gcs.ServiceAccountKeyJSON{
		Type:                    config.Config.GCSType,
		ProjectID:               config.Config.GCSProjectID,
		PrivateKeyID:            config.Config.GCSPrivateKeyID,
		PrivateKey:              stringPrivateKey,
		ClientEmail:             config.Config.GCSClientEmail,
		ClientID:                config.Config.GCSClientID,
		AuthURI:                 config.Config.GCSAuthURI,
		TokenURI:                config.Config.GCSTokenURI,
		AuthProviderX509CertURL: config.Config.GCSAuthProviderX509CertURL,
		ClientX509CertURL:       config.Config.GCSClientX509CertURL,
		UniverseDomain:          config.Config.GCSUniverseDomain,
	}
	// Mengembalikan objek kredensial storage GCS tersebut agar bisa digunakan app kita
	gcsClient := gcs.NewGCSClient(
		gcsServiceAccount,
		config.Config.GCSBucketName,
	)
	return gcsClient
}
