// Package clients bertindak sebagai penghubung ke berbagai layanan/service lain.
package clients

import (
	"field-service/clients/config"
	clients "field-service/clients/user"
	config2 "field-service/config"
)

// ClientRegistry adalah pusat pendaftaran untuk semua client yang ada.
// Ini memudahkan bagian lain aplikasi untuk mendapatkan client yang sudah dikonfigurasi.
type ClientRegistry struct{}

// IClientRegistry adalah interface yang mencatat daftar client yang tersedia.
// Interface ini mendefinisikan kontrak fungsi apa saja yang harus dimiliki oleh registry.
type IClientRegistry interface {
	// GetUser adalah fungsi untuk mendapatkan instance client User.
	GetUser() clients.IUserClient
}

// NewClientRegistry membuat instance baru dari ClientRegistry.
// Fungsi ini mengembalikan interface IClientRegistry agar implementasi internal tersembunyi.
func NewClientRegistry() IClientRegistry {
	return &ClientRegistry{}
}

// GetUser menginisialisasi dan memberikan User Client yang sudah siap pakai.
// Fungsi ini mengambil konfigurasi dari file config global dan menyuntikkannya ke User Client.
func (c *ClientRegistry) GetUser() clients.IUserClient {
	// Di sini kita membuat konfigurasi client mencakup URL host dan Signature Key dari config global.
	return clients.NewUserClient(
		config.NewClientConfig(
			config.WithBaseURL(config2.Config.InternalService.User.Host),
			config.WithSignatureKey(config2.Config.InternalService.User.SignatureKey),
		))
}
