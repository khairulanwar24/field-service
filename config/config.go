package config

import (
	"field-service/common/util"
	"os"

	"github.com/sirupsen/logrus"
	_ "github.com/spf13/viper/remote"
)

// Config adalah variabel global yang menyimpan seluruh konfigurasi aplikasi saat berjalan.
var Config AppConfig

// AppConfig adalah cetak biru (struct) dari seluruh data pengaturan (seperti port, nama database, dll).
type AppConfig struct {
	Port                       int             `json:"port"`
	AppName                    string          `json:"appName"`
	AppEnv                     string          `json:"appEnv"`
	SignatureKey               string          `json:"signatureKey"`
	Database                   Database        `json:"database"`
	RateLimiterMaxRequest      float64         `json:"rateLimiterMaxRequest"`
	RateLimiterTimeSecond      int             `json:"rateLimiterTimeSecond"`
	InternalService            InternalService `json:"internalService"`
	GCSType                    string          `json:"gcsType"`
	GCSProjectID               string          `json:"gcsProjectID"`
	GCSPrivateKeyID            string          `json:"gcsPrivateKeyID"`
	GCSPrivateKey              string          `json:"gcsPrivateKey"`
	GCSClientEmail             string          `json:"gcsClientEmail"`
	GCSClientID                string          `json:"gcsClientID"`
	GCSAuthURI                 string          `json:"gcsAuthURI"`
	GCSTokenURI                string          `json:"gcsTokenURI"`
	GCSAuthProviderX509CertURL string          `json:"gcsAuthProviderX509CertURL"`
	GCSClientX509CertURL       string          `json:"gcsClientX509CertURL"`
	GCSUniverseDomain          string          `json:"gcsUniverseDomain"`
	GCSBucketName              string          `json:"gcsBucketName"`
}

// Database menyimpan konfigurasi khusus untuk koneksi ke PostgreSQL/DB.
type Database struct {
	Host                  string `json:"host"`
	Port                  int    `json:"port"`
	Name                  string `json:"name"`
	Username              string `json:"username"`
	Password              string `json:"password"`
	MaxOpenConnections    int    `json:"maxOpenConnections"`
	MaxLifeTimeConnection int    `json:"maxLifeTimeConnection"`
	MaxIdleConnections    int    `json:"maxIdleConnections"`
	MaxIdleTime           int    `json:"maxIdleTime"`
}

// InternalService menampung data URL atau info mikroservis lain yang dihubungi oleh service ini.
type InternalService struct {
	User User `json:"user"`
}

// User menyimpan konfigurasi khusus untuk komunikasi dengan User Service.
type User struct {
	Host         string `json:"host"`
	SignatureKey string `json:"signatureKey"`
}

// Init adalah fungsi untuk membaca file pengaturan (seperti config.json) dan memasukkannya ke variabel Config.
func Init() {
	// 1. Coba baca pengaturan dari file lokal "config.json".
	err := util.BindFromJSON(&Config, "config.json", ".")
	if err != nil {
		// 2. Jika gagal membaca file lokal, catat informasi tersebut.
		logrus.Infof("failed to bind config: %v", err)
		// 3. Kemudian, coba baca pengaturan dari layanan eksternal Consul (Distributed Config).
		err = util.BindFromConsul(&Config, os.Getenv("CONSUL_HTTP_URL"), os.Getenv("CONSUL_HTTP_PATH"))
		if err != nil {
			// 4. Jika dari Consul juga gagal, maka matikan aplikasi karena tanpa config aplikasi tidak bisa jalan (panic).
			panic(err)
		}
	}
}
