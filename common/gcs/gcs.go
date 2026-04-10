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

// ServiceAccountKeyJSON adalah struktur data (struct) yang merepresentasikan isi dari file
// kredensial (JSON) untuk akun layanan Google Cloud (GCP Service Account).
// Struct ini digunakan untuk memetakan (parsing) data JSON yang diperlukan untuk otentikasi.
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

// GCSClient adalah struct utama yang menyimpan konfigurasi untuk terhubung dengan
// Google Cloud Storage (GCS). Ini menampung kredensial akun layanan (ServiceAccountKeyJSON)
// dan nama bucket tempat file akan disimpan.
type GCSClient struct {
	ServiceAccountKeyJSON ServiceAccountKeyJSON
	BucketName            string
}

// IGCSClient adalah sebuah interface. Interface ini bertindak sebagai "kontrak" yang
// menentukan fungsi apa saja yang harus dimiliki oleh sebuah GCS Client.
// Antarmuka ini hanya memiliki satu metode yaitu UploadFile.
type IGCSClient interface {
	UploadFile(context.Context, string, []byte) (string, error)
}

// NewGCSClient adalah fungsi "pembuat" (constructor). Fungsi ini digunakan untuk
// membuat instance baru dari GCSClient. Fungsi mengembalikan interface IGCSClient.
func NewGCSClient(serviceAccountKeyJSON ServiceAccountKeyJSON, bucketName string) IGCSClient {
	return &GCSClient{
		ServiceAccountKeyJSON: serviceAccountKeyJSON,
		BucketName:            bucketName,
	}
}

// createClient adalah metode (method) dari struct GCSClient untuk menginisialisasi
// client bawaan dari SDK Google Cloud Storage secara langsung menggunakan kredensial JSON.
// Metode ini bersifat privat (diawali huruf kecil) karena hanya digunakan di dalam file gcs.go ini.
func (g *GCSClient) createClient(ctx context.Context) (*storage.Client, error) {
	// 1. Siapkan sebuah penyangga (buffer) untuk menyimpan data JSON
	reqBodyBytes := new(bytes.Buffer)

	// 2. Ubah struct ServiceAccountKeyJSON menjadi wujud JSON kembali dan masukkan ke dalam penyangga (buffer)
	err := json.NewEncoder(reqBodyBytes).Encode(g.ServiceAccountKeyJSON)
	if err != nil {
		logrus.Errorf("failed to encode service account key json: %v", err)
		return nil, err
	}

	// 3. Ambil data JSON dalam bentuk kumpulan byte (slice of bytes)
	jsonByte := reqBodyBytes.Bytes()

	// 4. Buatlah client penyimpanan (storage client) Google Cloud Storage baru
	// dengan memasukkan kredensial dari data JSON tersebut
	client, err := storage.NewClient(ctx, option.WithCredentialsJSON(jsonByte))
	if err != nil {
		logrus.Errorf("failed to create client: %v", err)
		return nil, err
	}

	// 5. Kembalikan client yang telah berhasil dibuat
	return client, nil
}

// UploadFile adalah metode untuk mengunggah file yang sudah berupa byte ke bucket GCS.
// Fungsi ini membutuhkan: ctx (konteks atau latar pelaksanaan), filename (nama file tujuan),
// dan data (isi dari file dalam bentuk array byte).
func (g *GCSClient) UploadFile(ctx context.Context, filename string, data []byte) (string, error) {
	var (
		// Tipe konten standar (berlaku untuk semua tipe data/file yang tidak didefinisikan spesifik)
		contentType = "application/octet-stream"
		// Batas waktu maksimal saat proses upload ke GCS, yang pada kasus ini diset 60 detik (1 menit)
		timeoutInSeconds = 60
	)

	// 1. Buat client GCS menggunakan kredensial yang dimiliki
	client, err := g.createClient(ctx)
	if err != nil {
		logrus.Errorf("failed to create client: %v", err)
		return "", err
	}

	// 2. Tunda eksekusi (defer) untuk menutup sambungan client secara otomatis
	// saat fungsi UploadFile ini selesai dijalankan agar tidak ada kebocoran memory
	defer func(client *storage.Client) {
		err := client.Close()
		if err != nil {
			logrus.Errorf("failed to close client: %v", err)
			return
		}
	}(client)

	// 3. Batasi waktu penyelesaian (timeout) untuk proses upload dengan batas 60 detik.
	// Jika melewati itu, ctx akan dibatalkan otomatis melalui fungsi cancel()
	ctx, cancel := context.WithTimeout(ctx, time.Duration(timeoutInSeconds)*time.Second)
	defer cancel() // Pastikan cancel dijalankan saat proses selesai

	// 4. Hubungkan dengan bucket dan persiapkan objek dengan nama sesuai parameter 'filename'
	bucket := client.Bucket(g.BucketName)
	object := bucket.Object(filename)

	// 5. Siapkan data file yang masuk sebagai buffer (ruang memori sementara) lalu buat penulis (writer)
	buffer := bytes.NewBuffer(data)
	writer := object.NewWriter(ctx)
	writer.ChunkSize = 0 // Jangan membagi data ke pecahan spesifik (chunk) dan langung upload saja

	// 6. Eksekusi salin/unggah (copy) dari buffer (data milik kita) menuju ke writer (objek di GCS)
	_, err = io.Copy(writer, buffer)
	if err != nil {
		logrus.Errorf("failed to copy: %v", err)
		return "", err
	}

	// 7. Tutup 'writer' untuk menandakan bahwa pengunggahan file telah selesai sepenuhnya
	err = writer.Close()
	if err != nil {
		logrus.Errorf("failed to close: %v", err)
		return "", err
	}

	// 8. Tentukan keterangan tambahan dari objek (Metadata) berupa tipe kontennya (ContentType)
	_, err = object.Update(ctx, storage.ObjectAttrsToUpdate{ContentType: contentType})
	if err != nil {
		logrus.Errorf("failed to update: %v", err)
		return "", err
	}

	// 9. Susun dan kembalikan penautan (URL) yang mengarah kepada file yang baru saja diunggah ke Google Cloud Storage
	url := fmt.Sprintf("https://storage.googleapis.com/%s/%s", g.BucketName, filename)
	return url, nil
}
