// Package clients berisi logika untuk berkomunikasi dengan layanan lain (User Service).
package clients

import (
	"context"
	"field-service/clients/config" // Import konfigurasi dasar client
	"field-service/common/util"   // Tool bantuan seperti SHA256
	config2 "field-service/config" // Config global aplikasi
	"field-service/constants"      // Konstanta untuk nama-nama header
	"fmt"
	"net/http"
	"time"
)

// UserClient adalah struct yang bertugas mengirim request HTTP ke service User.
type UserClient struct {
	client config.IClientConfig // Objek konfigurasi yang menyimpan data client & URL
}

// IUserClient adalah interface kontrak fungsi. 
// Ini memberi tahu kita fungsi apa saja yang harus tersedia di UserClient.
type IUserClient interface {
	GetUserByToken(context.Context) (*UserData, error) // Fungsi utama: ambil data user dari token
}

// NewUserClient adalah fungsi 'Constructor'. Digunakan untuk membuat object UserClient baru.
func NewUserClient(client config.IClientConfig) IUserClient {
	return &UserClient{client: client}
}

// GetUserByToken memanggil API User untuk memvalidasi token dan mendapatkan detail profil.
func (u *UserClient) GetUserByToken(ctx context.Context) (*UserData, error) {
	// 1. Membuat API Key unik sebagai pengaman tambahan.
	// Kita menggabungkan nama aplikasi, kunci rahasia, dan waktu (timestamp).
	unixTime := time.Now().Unix() 
	generateAPIKey := fmt.Sprintf("%s:%s:%d",
		config2.Config.AppName,
		u.client.SignatureKey(),
		unixTime,
	)
	apiKey := util.GenerateSHA256(generateAPIKey) // Hash gabungan tadi menjadi kode SHA256

	// 2. Mengambil Token Autentikasi dari context.
	// Context menyimpan metadata request, di sini kita ambil token yang dikirim user.
	token := ctx.Value(constants.Token).(string)
	bearerToken := fmt.Sprintf("Bearer %s", token)

	var response UserResponse
	// 3. Membangun request HTTP.
	// Kita memasukkan berbagai header keamanan dan Identitas sebelum mengirim.
	request := u.client.Client(). 
		Set(constants.Authorization, bearerToken).      // Header untuk login
		Set(constants.XServiceName, config2.Config.AppName). // Siapa yang memanggil
		Set(constants.XApiKey, apiKey).                  // Kode SHA256 tadi
		Set(constants.XRequestAt, fmt.Sprintf("%d", unixTime)). // Waktu pemanggilan
		Get(fmt.Sprintf("%s/api/v1/auth/user", u.client.BaseURL())) // Tujuan URL API

	// 4. Mengeksekusi request dan menunggu hasilnya masuk ke struct 'response'.
	resp, _, errs := request.EndStruct(&response)
	if len(errs) > 0 {
		return nil, errs[0] // Error jika gagal koneksi ke jaringan
	}

	// 5. Cek apakah status code yang dikembalikan adalah 200 OK.
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("user response: %s", response.Message)
	}

	// 6. Mengembalikan data user yang berhasil diambil.
	return &response.Data, nil
}
