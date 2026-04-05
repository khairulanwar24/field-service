package config

import (
	"fmt"
	"net/url"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// InitDatabase adalah fungsi untuk membuka koneksi ke database PostgreSQL menggunakan informasi dari config.
func InitDatabase() (*gorm.DB, error) {
	// 1. Ambil seluruh data konfigurasi yang sudah di-load sebelumnya di variabel Config.
	config := Config

	// 2. Enkripsi/amankan karakter unik pada password (seperti @ atau !) agar tidak merusak format URL.
	encodedPassword := url.QueryEscape(config.Database.Password)

	// 3. Rangkai URL koneksi PostgreSQL (contoh: postgresql://user:pass@localhost:5432/namadb?sslmode=disable).
	uri := fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?sslmode=disable",
		config.Database.Username,
		encodedPassword,
		config.Database.Host,
		config.Database.Port,
		config.Database.Name,
	)

	// 4. Minta gorm (Pustaka ORM untuk Go) untuk mencoba membuka jalan/sambungan ke database memakai URL tadi.
	db, err := gorm.Open(postgres.Open(uri), &gorm.Config{})
	if err != nil {
		return nil, err // Jika gagal menyambung, kembalikan pesan error.
	}

	// 5. Ambil objek database asli (sqlDB) dari dalam GORM agar kita bisa menyetel batas konekinya.
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// 6. Atur batas maksimal dan waktu hidup (umur) setiap sambungan ke database agar ram tidak jebol.
	sqlDB.SetMaxIdleConns(config.Database.MaxIdleConnections)                                    // Batas sambungan yang nganggur/standby.
	sqlDB.SetMaxOpenConns(config.Database.MaxOpenConnections)                                    // Maksimal total sambungan yang boleh dibuka bersamaan.
	sqlDB.SetConnMaxLifetime(time.Duration(config.Database.MaxLifeTimeConnection) * time.Second) // Umur maksimal sambungan sebelum disuruh refresh.
	sqlDB.SetConnMaxIdleTime(time.Duration(config.Database.MaxIdleTime) * time.Second)           // Batas waktu tunggu sambungan yang nganggur sebelum diputus otomatis.

	// 7. Kembalikan objek database (db) yang sudah siap dipakai ke fungsi yang memanggilnya.
	return db, nil
}
