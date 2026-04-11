// Package middlewares berisi middleware yang digunakan untuk memproses HTTP request
// Middleware adalah fungsi yang dijalankan sebelum request sampai ke handler utama
// Middleware biasa digunakan untuk: autentikasi, validasi, rate limiting, error handling, dll
package middlewares

import (
	// package standard library Go
	"context"       // untuk mengelola konteks (data yang dikirim antar fungsi)
	"crypto/sha256" // untuk membuat hash SHA256
	"encoding/hex"  // untuk enkoding/dekoding hexadecimal
	"fmt"           // untuk formatting string (printf, sprintf, dll)
	"net/http"      // untuk HTTP status code dan metode HTTP
	"strings"       // untuk manipulasi string

	// package internal dari aplikasi field-service
	"field-service/clients"                     // untuk mengelola client registry
	"field-service/common/response"             // untuk struktur response HTTP
	"field-service/config"                      // untuk konfigurasi aplikasi
	"field-service/constants"                   // untuk konstanta yang digunakan
	errConstant "field-service/constants/error" // untuk error message (dengan alias "errConstant")

	// package eksternal dari library pihak ketiga
	"github.com/didip/tollbooth"         // library untuk rate limiting
	"github.com/didip/tollbooth/limiter" // limiter dari tollbooth
	"github.com/gin-gonic/gin"           // Web framework Gin
	"github.com/sirupsen/logrus"         // library untuk logging/mencatat log aplikasi
)

// HandlePanic adalah middleware yang menangkap panic/error dalam aplikasi
// Jika terjadi panic, middleware akan menangkapnya dan mengirimkan response error ke client
// Tanpa middleware ini, aplikasi akan crash ketika ada panic
// Middleware type "gin.HandlerFunc" berarti fungsi yang bisa digunakan di Gin framework
func HandlePanic() gin.HandlerFunc {
	// Return function yang sebenarnya dijalankan oleh Gin framework
	// Parameter "c" adalah context dari request/response HTTP
	return func(c *gin.Context) {
		// defer digunakan untuk menjalankan code di akhir function
		// recover() digunakan untuk menangkap panic dan mencegah aplikasi crash
		defer func() {
			// Jika ada panic yang tertangkap, r akan berisi nilai panic tersebut
			if r := recover(); r != nil {
				// Log error ke file/console menggunakan logrus
				logrus.Errorf("Recovered from panic: %v", r)
				// Kirim response error ke client dengan HTTP status 500 (Internal Server Error)
				c.JSON(http.StatusInternalServerError, response.Response{
					Status:  constants.Error,                            // status: "error"
					Message: errConstant.ErrInternalServerError.Error(), // pesan error
				})
				// Abort menghentikan eksekusi handler berikutnya
				c.Abort()
			}
		}()
		// c.Next() memanggil handler berikutnya (middleware atau route handler)
		c.Next()
	}
}

// RateLimiter adalah middleware yang membatasi jumlah request dari client
// Misal: jika client mengirim 100 request dalam 1 menit, maka request ke 101 akan ditolak
// Tujuan: mencegah abuse/penyalahgunaan API dan melindungi server dari beban berlebihan
// Parameter "lmt" adalah limiter yang sudah dikonfigurasi (misalnya: max 10 request per detik)
func RateLimiter(lmt *limiter.Limiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		// tollbooth.LimitByRequest() mengecek apakah request melebihi batas yang ditentukan
		// c.Writer dan c.Request dikirimkan untuk mengakses info request/response
		// Jika limit terlampaui, err akan berisi error
		err := tollbooth.LimitByRequest(lmt, c.Writer, c.Request)
		if err != nil {
			// Jika limit terlampaui, kirim response error dengan HTTP status 429 (Too Many Requests)
			c.JSON(http.StatusTooManyRequests, response.Response{
				Status:  constants.Error,                        // status: "error"
				Message: errConstant.ErrTooManyRequests.Error(), // pesan: "terlalu banyak request"
			})
			// Hentikan eksekusi handler berikutnya (jangan proses request)
			c.Abort()
		}
		// Jika tidak ada error (dalam batas), lanjutkan ke handler berikutnya
		c.Next()
	}
}

// extractBearerToken mengambil token autentikasi dari header Authorization
// Header Authorization biasanya format: "Bearer <token_string>"
// Fungsi ini memisahkan kata "Bearer" dan mengambil token-nya saja
// Contoh:
//
//	Input: "Bearer eyJhbGciOiJIUzI1NiIs..."
//	Output: "eyJhbGciOiJIUzI1NiIs..."
//
// Parameter "token" adalah nilai dari header Authorization
func extractBearerToken(token string) string {
	// strings.Split() memisahkan string berdasarkan delimiter (dalam hal ini spasi " ")
	// Hasilnya adalah slice (array) dari string
	// Contoh: "Bearer abc123" menjadi ["Bearer", "abc123"]
	arrayToken := strings.Split(token, " ")
	// Cek apakah hasil split menghasilkan 2 bagian (Bearer dan token)
	if len(arrayToken) == 2 {
		// arrayToken[1] adalah elemen kedua (index 1), yaitu token
		return arrayToken[1]
	}
	// Jika format salah, return empty string ""
	return ""
}

// responseUnauthorized mengirimkan response error 401 Unauthorized ke client
// Digunakan saat user tidak terautentikasi atau tidak punya akses
// Parameter "message" adalah pesan error yang ingin dikirimkan ke client
func responseUnauthorized(c *gin.Context, message string) {
	// c.JSON() mengirimkan response dalam format JSON
	// http.StatusUnauthorized adalah status code 401
	c.JSON(http.StatusUnauthorized, response.Response{
		Status:  constants.Error, // status: "error"
		Message: message,         // pesan error yang dikirimkan
	})
	// c.Abort() menghentikan eksekusi middleware/handler berikutnya
	c.Abort()
}

// validateAPIKey memvalidasi signature dari request untuk memastikan request berasal dari service yang authorized
// Perbandingan: seperti memberikan "password" untuk setiap request
// Cara kerja: URL-safe signature dibuat dengan hash SHA256 dari kombinasi serviceName + signatureKey + timestamp
// Jika client mengirim signature yang sama, berarti client terpercaya
// Return: error jika signature tidak valid, nil jika valid
func validateAPIKey(c *gin.Context) error {
	// Ambil header-header yang diperlukan dari request
	apiKey := c.GetHeader(constants.XApiKey)           // signature yang dikirim client
	requestAt := c.GetHeader(constants.XRequestAt)     // timestamp request
	serviceName := c.GetHeader(constants.XServiceName) // nama service yang request
	// signatureKey adalah secret key yang tersimpan di config server
	signatureKey := config.Config.SignatureKey

	// Buat format string: "serviceName:signatureKey:timestamp"
	// Contoh: "user-service:secret123:2024-04-11T10:30:00Z"
	validateKey := fmt.Sprintf("%s:%s:%s", serviceName, signatureKey, requestAt)

	// sha256.New() membuat hasher baru untuk SHA256
	hash := sha256.New()
	// hash.Write() mengisi data yang akan di-hash
	// []byte() mengkonversi string menjadi byte array
	hash.Write([]byte(validateKey))

	// hash.Sum(nil) menghasilkan hash dalam bentuk byte array
	// hex.EncodeToString() mengkonversi byte array menjadi string hexadecimal
	// Contoh: [255, 204, 170] menjadi "ffccaa"
	resultHash := hex.EncodeToString(hash.Sum(nil))

	// Bandingkan hash yang diterima dari client dengan hash yang kita hitung
	// Jika tidak sama, berarti signature tidak valid
	if apiKey != resultHash {
		return errConstant.ErrUnauthorized
	}
	// Signature valid
	return nil
}

// contains mengecek apakah value "role" ada di dalam slice "roles"
// Ini adalah fungsi utility (helper) untuk memberikan lebih mudah
// Contoh:
//
//	roles := []string{"admin", "user", "guest"}
//	contains(roles, "admin") -> true
//	contains(roles, "superadmin") -> false
//
// Parameter:
//   - roles: slice (array) string yang berisi daftar role
//   - role: string role yang ingin dicari
//
// Return: true jika role ada, false jika tidak ada
func contains(roles []string, role string) bool {
	// for loop untuk iterasi setiap elemen di dalam slice roles
	// r adalah nilai elemen yang sedang diiterasi, seperti: "admin", "user", "guest"
	for _, r := range roles {
		// Jika elemen sama dengan role yang dicari, return true
		if r == role {
			return true
		}
	}
	// Jika loop selesai dan tidak menemukan, return false
	return false
}

// CheckRole adalah middleware untuk memvalidasi apakah user punya role yang diizinkan
// Contoh: jika route hanya bisa diakses "admin", middleware ini akan cek apakah user adalah admin
// Parameter:
//   - roles: slice of string yang berisi role-role yang diizinkan (contoh: ["admin", "manager"])
//   - client: client registry untuk mengakses service layanan lain (dalam hal ini service user)
//
// Return: gin.HandlerFunc yang bisa digunakan sebagai middleware di Gin framework
func CheckRole(roles []string, client clients.IClientRegistry) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Ambil user data berdasarkan token yang ada di context (sudah diset oleh middleware Authenticate)
		// c.Request.Context() berisi konteks dari request
		// GetUserByToken() adalah method yang mengambil user berdasarkan token mereka
		user, err := client.GetUser().GetUserByToken(c.Request.Context())
		// Jika ada error (token tidak valid, user tidak ditemukan, dll)
		if err != nil {
			// Kirim response error 401 Unauthorized
			responseUnauthorized(c, errConstant.ErrUnauthorized.Error())
			return
		}

		// Cek apakah role user ada di dalam slice roles yang diizinkan
		// Menggunakan fungsi contains() yang sudah didefinisikan
		if !contains(roles, user.Role) {
			// Jika role user tidak ada di list yang diizinkan, kirim error
			responseUnauthorized(c, errConstant.ErrUnauthorized.Error())
			return
		}
		// Jika semua validasi lolos, lanjutkan ke handler berikutnya
		c.Next()
	}
}

// Authenticate adalah middleware untuk validasi autentikasi
// Middleware ini memverifikasi:
//  1. Header Authorization ada dan tidak kosong
//  2. Signature API Key valid (menggunakan validateAPIKey)
//  3. Token Bearer ada dan disimpan di context untuk digunakan handler berikutnya
//
// Return: gin.HandlerFunc yang bisa digunakan sebagai middleware di Gin framework
func Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		var err error
		// Ambil header Authorization (biasanya berisi "Bearer <token>")
		token := c.GetHeader(constants.Authorization)

		// Validasi: Header Authorization harus ada dan tidak kosong
		if token == "" {
			responseUnauthorized(c, errConstant.ErrUnauthorized.Error())
			return
		}

		// Validasi: Cek apakah signature API Key valid
		// validateAPIKey() menghitung SHA256 dan membandingkan dengan API Key yang dikirim
		err = validateAPIKey(c)
		if err != nil {
			responseUnauthorized(c, err.Error())
			return
		}

		// Ambil token dari header Authorization (pisahkan "Bearer" dan token-nya)
		tokenString := extractBearerToken(token)
		// Simpan token ke dalam context request
		// context.WithValue() membuat context baru dengan menyimpan data key-value
		// constants.Token adalah key, tokenString adalah value
		// c.Request.WithContext() membuat request baru dengan context yang sudah diupdate
		tokenUser := c.Request.WithContext(context.WithValue(c.Request.Context(), constants.Token, tokenString))
		c.Request = tokenUser // Update request dengan context yang baru

		// Lanjutkan ke handler berikutnya (context sudah berisi token)
		c.Next()
	}
}

// AuthenticateWithoutToken adalah middleware untuk validasi API Key tanpa memerlukan token Bearer
// Digunakan untuk public endpoints atau endpoints yang tidak memerlukan user authentication
// Tetapi tetap perlu memverifikasi bahwa request adalah dari service yang authorized (signature valid)
// Return: gin.HandlerFunc yang bisa digunakan sebagai middleware di Gin framework
func AuthenticateWithoutToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Validasi: Cek apakah signature API Key valid
		// Ini adalah validasi sekuritas untuk memastikan request dari service yang terpercaya
		err := validateAPIKey(c)
		if err != nil {
			// Jika signature tidak valid, kirim response error unauthorized
			responseUnauthorized(c, err.Error())
			return
		}
		// Signature valid, lanjutkan ke handler berikutnya
		// Catatan: tidak ada token bearer yang disimpan di context (karena tidak ada token)
		c.Next()
	}
}
