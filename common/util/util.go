package util

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/dustin/go-humanize"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	_ "github.com/spf13/viper/remote"
	"math"
	"os"
	"reflect"
	"strconv"
	"strings"
)

// PaginationParam adalah cetak biru parameter masukan untuk membuat sistem halaman (pagination).
type PaginationParam struct {
	Count int64       `json:"count"` // Total keseluruhan data.
	Page  int         `json:"page"`  // Halaman saat ini.
	Limit int         `json:"limit"` // Batas jumlah data per halaman.
	Data  interface{} `json:"data"`  // Kumpulan data asli yang ingin ditampilkan.
}

// PaginationResult adalah wujud hasil akhir kembalian Pagination yang siap dikirim.
type PaginationResult struct {
	TotalPage    int         `json:"totalPage"`    // Total seluruh kemungkinan halaman.
	TotalData    int64       `json:"totalData"`    // Total data (sama seperti Count).
	NextPage     *int        `json:"nextPage"`     // Nomor Halaman berikutnya (bisa nil/kosong jika ini halaman terakhir).
	PreviousPage *int        `json:"previousPage"` // Nomor Halaman sebelumnya (bisa nil/kosong jika ini halaman pertama).
	Page         int         `json:"page"`         // Halaman saat ini.
	Limit        int         `json:"limit"`        // Batas data per halaman.
	Data         interface{} `json:"data"`         // Datanya.
}

// GeneratePagination meracik data masukan menjadi struktur bersistem halaman (pagination) yang rapi.
func GeneratePagination(params PaginationParam) PaginationResult {
	// 1. Hitung total halaman dengan cara membagi (Total Data / Limit Data), lalu dibulatkan ke atas (Ceil).
	totalPage := int(math.Ceil(float64(params.Count) / float64(params.Limit)))

	var (
		nextPage     int
		previousPage int
	)
	
	// 2. Jika halaman saat ini masih kurang dari total halaman, berarti masih ada halaman "Selanjutnya" (Next).
	if params.Page < totalPage {
		nextPage = params.Page + 1
	}

	// 3. Jika halaman saat ini lebih dari 1, berarti sudah pasti ada halaman "Sebelumnya" (Previous).
	if params.Page > 1 {
		previousPage = params.Page - 1
	}

	// 4. Susun semua perhitungannya ke dalam struct khusus (PaginationResult).
	result := PaginationResult{
		TotalPage:    totalPage,
		TotalData:    params.Count,
		NextPage:     &nextPage,
		PreviousPage: &previousPage,
		Page:         params.Page,
		Limit:        params.Limit,
		Data:         params.Data,
	}
	return result
}

// GenerateSHA256 men-generate hash/kode rahasia acak searah (SHA-256) menggunakan sebuah kata sandi asli.
func GenerateSHA256(inputString string) string {
	hash := sha256.New()
	hash.Write([]byte(inputString))
	hashBytes := hash.Sum(nil)
	hashString := hex.EncodeToString(hashBytes) // Mengubah kumpulan byte ruwet jadi tulisan yang bisa dibaca ("hex").
	return hashString
}

// RupiahFormat mengubah nilai angka mentah float64 menjadi teks bersimbol "Rp." layaknya mata uang Indonesia.
func RupiahFormat(amount *float64) string {
	stringValue := "0"
	if amount != nil {
		// Pasangkan koma setiap 3 angka pakai humanize, lalu tukar tanda koma (,) jadi titik (.).
		humanizeValue := humanize.CommafWithDigits(*amount, 0)
		stringValue = strings.ReplaceAll(humanizeValue, ",", ".")
	}
	// Tambahkan tulisan "Rp. " di depan nilai uangnya.
	return fmt.Sprintf("Rp. %s", stringValue)
}

// BindFromJSON dipakai untuk membaca file berakhiran ".json" lokal ke dalam variabel Go.
func BindFromJSON(dest any, filename, path string) error {
	v := viper.New()

	v.SetConfigType("json") // Menentukan jenis format file-nya (JSON).
	v.AddConfigPath(path)   // Mengatur di mana letak filenya.
	v.SetConfigName(filename) // Mengatur nama filenya.

	err := v.ReadInConfig() // Baca bentuk fisiknya.
	if err != nil {
		return err
	}

	// Unmarshal: proses menterjemahkan isi file JSON tersebut untuk lalu dicetak ulang menyatu dengan 'dest' (Data Struct milik Anda).
	err = v.Unmarshal(&dest)
	if err != nil {
		logrus.Errorf("failed to unmarshal: %v", err)
		return err
	}

	return nil
}

// SetEnvFromConsulKV berfungsi mengekstrak informasi di dalam Consul KV agar dimasukkan otomatis ke file Environment (Env) OS Server Anda.
func SetEnvFromConsulKV(v *viper.Viper) error {
	env := make(map[string]any)

	err := v.Unmarshal(&env)
	if err != nil {
		logrus.Errorf("failed to unmarshal: %v", err)
		return err
	}

	// 1. Ulangi untuk setiap baris data config. Ubah format datanya menjadi "String/teks" berdasarkan jenis aslinya (bool, float, dll).
	for k, v := range env {
		var (
			valOf = reflect.ValueOf(v)
			val   string
		)

		switch valOf.Kind() {
		case reflect.String:
			val = valOf.String()
		case reflect.Int:
			val = strconv.Itoa(int(valOf.Int()))
		case reflect.Uint:
			val = strconv.Itoa(int(valOf.Uint()))
		case reflect.Float32:
			val = strconv.Itoa(int(valOf.Float()))
		case reflect.Float64:
			val = strconv.Itoa(int(valOf.Float()))
		case reflect.Bool:
			val = strconv.FormatBool(valOf.Bool())
		}

		// 2. Set nilainya menjadi Environmental Variable tingkat Server OS.
		err = os.Setenv(k, val)
		if err != nil {
			logrus.Errorf("failed to set env: %v", err)
			return err
		}
	}

	return nil
}

// BindFromConsul adalah fungsi yang mengambil pengaturan/konfigurasi tersimpan layaknya dari server cloud Consul.
func BindFromConsul(dest any, endPoint, path string) error {
	v := viper.New()
	v.SetConfigType("json")
	// 1. Hubungkan provider (si penyedia rahasia) ke Remote Consul server.
	err := v.AddRemoteProvider("consul", endPoint, path)
	if err != nil {
		logrus.Errorf("failed to add remote provider: %v", err)
		return err
	}

	// 2. Jika berhasil terhubung, mulailah mengunduh lalu membaca data dari jarak jauh.
	err = v.ReadRemoteConfig()
	if err != nil {
		logrus.Errorf("failed to read remote config: %v", err)
		return err
	}

	// 3. Masukkan data-data awan tersebut ke tempat penyimpanan yang kita tentukan ('dest').
	err = v.Unmarshal(&dest)
	if err != nil {
		logrus.Errorf("failed to unmarshal: %v", err)
		return err
	}

	// 4. Setelah memindahkannya ke Dest, sekalian set pula nilainya ke System Environment (Biar Go-nya familiar).
	err = SetEnvFromConsulKV(v)
	if err != nil {
		logrus.Errorf("failed to set env from consul kv: %v", err)
		return err
	}

	return nil
}
