// Package config mengelola konfigurasi dasar untuk client HTTP.
package config

import "github.com/parnurzeal/gorequest"

// ClientConfig adalah struct (struktur data) yang menampung informasi konfigurasi.
// Ini adalah "wadah" untuk client HTTP, URL dasar, dan kunci tanda tangan.
type ClientConfig struct {
	client       *gorequest.SuperAgent // Object client dari library gorequest untuk kirim request
	baseURL      string                // Alamat server tujuan (contoh: https://api.bola.com)
	signatureKey string                // Kunci rahasia untuk keamanan/autentikasi (opsional)
}

// IClientConfig adalah interface (kontrak). 
// Isinya adalah daftar fungsi yang harus dimiliki oleh siapapun yang ingin menjadi "ClientConfig".
// Ini berguna agar kode kita lebih fleksibel dan mudah di-test (mocking).
type IClientConfig interface {
	Client() *gorequest.SuperAgent
	BaseURL() string
	SignatureKey() string
}

// Option adalah tipe data berbentuk fungsi.
// Ini digunakan untuk pola "Functional Options" di Go, agar pembuatan config lebih rapi dan fleksibel.
type Option func(*ClientConfig)

// NewClientConfig adalah fungsi 'Constructor' untuk membuat instance baru dari ClientConfig.
// Parameter 'options' memungkinkan kita mengirimkan banyak konfigurasi tambahan sekaligus.
func NewClientConfig(options ...Option) IClientConfig {
	// 1. Inisialisasi default: membuat client baru dengan header JSON
	clientConfig := &ClientConfig{
		client: gorequest.New().
			Set("Content-Type", "application/json").
			Set("Accept", "application/json"),
	}

	// 2. Terapkan semua pilihan (options) yang diberikan oleh user
	for _, option := range options {
		option(clientConfig)
	}

	// 3. Kembalikan hasilnya dalam bentuk interface IClientConfig
	return clientConfig
}

// Client() mengembalikan object gorequest agar bisa digunakan untuk membuat request HTTP.
func (c *ClientConfig) Client() *gorequest.SuperAgent {
	return c.client
}

// BaseURL() mengambil nilai URL dasar yang sudah disimpan di dalam struct.
func (c *ClientConfig) BaseURL() string {
	return c.baseURL
}

// SignatureKey() mengambil nilai kunci tanda tangan yang sudah disimpan.
func (c *ClientConfig) SignatureKey() string {
	return c.signatureKey
}

// WithBaseURL adalah fungsi pembantu untuk mengisi field baseURL.
// Biasa digunakan saat memanggil NewClientConfig(WithBaseURL("..."))
func WithBaseURL(baseURL string) Option {
	return func(c *ClientConfig) {
		c.baseURL = baseURL
	}
}

// WithSignatureKey adalah fungsi pembantu untuk mengisi field signatureKey.
// Biasa digunakan saat memanggil NewClientConfig(WithSignatureKey("..."))
func WithSignatureKey(signatureKey string) Option {
	return func(c *ClientConfig) {
		c.signatureKey = signatureKey
	}
}
