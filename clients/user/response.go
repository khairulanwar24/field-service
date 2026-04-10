package clients

import "github.com/google/uuid"

// UserResponse adalah wrapper untuk hasil response dari API User.
type UserResponse struct {
	Code    int      `json:"code"`    // Kode status dari server (misal: 200)
	Status  string   `json:"status"`  // Status tekstual (misal: "Success")
	Message string   `json:"message"` // Pesan penjelasan
	Data    UserData `json:"data"`    // Data utama user
}

// UserData berisi detail profil user yang dikembalikan oleh API.
type UserData struct {
	UUID        uuid.UUID `json:"uuid"`        // ID unik user
	Name        string    `json:"name"`        // Nama lengkap
	Username    string    `json:"username"`    // Nama akun
	Email       string    `json:"email"`       // Alamat email
	Role        string    `json:"role"`        // Peran (misal: admin)
	PhoneNumber string    `json:"phoneNumber"` // Nomor telepon
}
