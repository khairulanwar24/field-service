package gcs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"cloud.google.com/go/storage"
	"github.com/sirupsen/logrus"
	"google.golang.org/api/option"
)

// ServiceAccountKeyJSON mencerminkan semua isi (struktur) dari file JSON Google Cloud kredensial Anda.
type ServiceAccountKeyJSON struct {
	Type                    string `json:"type"`
	ProjectID               string `json:"project_id"`
	PrivateKeyID            string `json:"private_key_id"`
	PrivateKey              string `json:"private_key"`
	ClientEmail             string `json:"client_email"`
	ClientID                string `json:"client_id"`
	AuthURI                 string `json:"auth_uri"`
	TokenURI                string `json:"token_uri"`
	AuthProviderX509CertURL string `json:"auth_provider_x509_cert_url"`
	ClientX509CertURL       string `json:"client_x509_cert_url"`
	UniverseDomain          string `json:"universe_domain"`
}

// GCSClient adalah struktur obyek penampung informasi kredensial JSON dan nama *bucket* (penyimpanan) GCS.
type GCSClient struct {
	ServiceAccountKeyJSON ServiceAccountKeyJSON
	BucketName            string
}

// IGCSClient adalah sebuah antarmuka (interface) kontrak yang mengharuskan adanya fungsi UploadFile.
type IGCSClient interface {
	UploadFile(context.Context, string, []byte) (string, error)
}

// NewGCSClient adalah fungsi "Constructor" untuk merakit dan mengembalikan objek GCSClient baru.
func NewGCSClient(serviceAccountKeyJSON ServiceAccountKeyJSON, bucketName string) IGCSClient {
	return &GCSClient{
		ServiceAccountKeyJSON: serviceAccountKeyJSON,
		BucketName:            bucketName,
	}
}

// createClient adalah fungsi internal (private) untuk membuat klien koneksi ke peladen Google Cloud Storage.
func (g *GCSClient) createClient(ctx context.Context) (*storage.Client, error) {
	// 1. Membuat penampung memori (buffer) untuk data JSON.
	reqBodyBytes := new(bytes.Buffer)
	// 2. Menerjemahkan/menulis nilai variabel dari bahasa Go ke dalam bentuk teks JSON murni lalu memasukkannya ke buffer.
	err := json.NewEncoder(reqBodyBytes).Encode(g.ServiceAccountKeyJSON)
	if err != nil {
		logrus.Errorf("failed to encode service account key json: %v", err)
		return nil, err
	}

	// 3. Ubah buffer JSON menjadi kumpulan "byte" murni agar dipahami komputer.
	jsonByte := reqBodyBytes.Bytes()
	// 4. Minta Google SDK untuk membuatkan koneksi dengan berbekal file JSON dalam bentuk *byte* tadi.
	client, err := storage.NewClient(ctx, option.WithCredentialsJSON(jsonByte))
	if err != nil {
		logrus.Errorf("failed to create client: %v", err)
		return nil, err
	}

	return client, nil
}

// UploadFile adalah fungsi publik yang bertugas mengunggah sebuah file data (*byte*) ke Google Cloud Storage (GCS).
func (g *GCSClient) UploadFile(ctx context.Context, filename string, data []byte) (string, error) {
	var (
		contentType      = "application/octet-stream" // Aturan tipe konten umum (file mentah).
		timeoutInSeconds = 60                         // Menunggu unggahan selesai maksimal 60 detik.
	)

	// 1. Panggil fungsi yang telah kita buat (createClient) agar bisa terhubung ke GCS.
	client, err := g.createClient(ctx)
	if err != nil {
		logrus.Errorf("failed to create client: %v", err)
		return "", err
	}

	// 2. Keyword 'defer' memastikan koneksi 'client' akan selalu ditutup di akhir fungsi ini, entah berhasil maupun error.
	defer func(client *storage.Client) {
		err := client.Close()
		if err != nil {
			logrus.Errorf("failed to close client: %v", err)
			return
		}
	}(client)

	// 3. Batasi waktu proses (timeout) agar tidak macet / hang abadi.
	ctx, cancel := context.WithTimeout(ctx, time.Duration(timeoutInSeconds)*time.Second)
	// Keyword 'defer' lagi untuk memastikan proses waktu mundur (timeout) tidak lupa dimatikan secara perlahan di akhir.
	defer cancel()

	// 4. Siapkan lokasi ember (Bucket) dan juga nama file incaran (Object) di Google Cloud.
	bucket := client.Bucket(g.BucketName)
	object := bucket.Object(filename)
	buffer := bytes.NewBuffer(data)

	// 5. Buka jalur untuk mulai menulis (mengirim) data ke penampung online GCS tadi.
	writer := object.NewWriter(ctx)
	writer.ChunkSize = 0 // Chunk 0 berarti tanpa pembagian ukuran, Google yang menentukan porsi memori.

	// 6. Menyalin aliran data file (buffer) mentah untuk dikirim melalui penulis (writer) secara otomatis ke awan/cloud.
	_, err = io.Copy(writer, buffer)
	if err != nil {
		logrus.Errorf("failed to copy: %v", err)
		return "", err
	}

	// 7. Jika proses copy selesai tertulis seluruhnya, mari segera tutup sambungannya.
	err = writer.Close()
	if err != nil {
		logrus.Errorf("failed to close: %v", err)
		return "", err
	}

	// 8. Terapkan informasi tamabahan (Metadata/Type) ke dalam file Google Cloud Storage agar tidak ditolak pembaca web.
	_, err = object.Update(ctx, storage.ObjectAttrsToUpdate{ContentType: contentType})
	if err != nil {
		logrus.Errorf("failed to update: %v", err)
		return "", err
	}

	// 9. Rangkai URL utuh menuju file hasil unggahan (format alamat Google API penyimpanan awan).
	url := fmt.Sprintf("https://storage.googleapis.com/%s/%s", g.BucketName, filename)
	// 10. Kembalikan link URL lengkap tersebut!
	return url, nil
}
