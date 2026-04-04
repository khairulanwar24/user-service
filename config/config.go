package config // Package config digunakan untuk mengatur konfigurasi aplikasi

import (
	"os"                       // Package bawaan Go untuk berinteraksi dengan sistem operasi (misalnya membaca environment variable)
	"user-service/common/util" // Package utility/bantuan bawaan dari project ini sendiri

	"github.com/sirupsen/logrus"      // Library eksternal untuk melakukan logging (mencatat aktivitas/error)
	_ "github.com/spf13/viper/remote" // Library eksternal untuk membaca konfigurasi dari Consul
)

// Config adalah variabel global yang akan menyimpan seluruh konfigurasi aplikasi.
// Variabel ini bisa diakses dari file lain dengan memanggil config.Config
var Config AppConfig

// AppConfig adalah struktur data yang mendefinisikan bentuk konfigurasi utama aplikasi.
// Tag `json:"..."` digunakan agar Go tahu cara membaca data ini dari format JSON.
type AppConfig struct {
	Port                  int      `json:"port"`                  // Port aplikasi berjalan (misal: 8080)
	AppName               string   `json:"appName"`               // Nama aplikasi kita
	AppEnv                string   `json:"appEnv"`                // Environment aplikasi (misal: development, production)
	SignatureKey          string   `json:"signatureKey"`          // Kunci rahasia untuk signature fitur tertentu
	Database              Database `json:"database"`              // Konfigurasi khusus untuk database (berisi struct Database di bawah)
	RateLimiterMaxRequest float64  `json:"rateLimiterMaxRequest"` // Batas jumlah maksimal request (untuk mencegah spam)
	RateLimiterTimeSecond int      `json:"rateLimiterTimeSecond"` // Batas waktu dalam detik untuk hitungan pembatasan request
	JwtSecretKey          string   `json:"jwtSecretKey"`          // Kunci rahasia pembuat dan pemvalidasi token JWT (untuk login)
	JwtExpirationTime     int      `json:"jwtExpirationTime"`     // Waktu berlaku token JWT sebelum kadaluarsa
}

// Database adalah struktur data yang mendefinisikan info detail koneksi ke database.
type Database struct {
	Host                  string `json:"host"`                  // Alamat host server database (misal: localhost atau IP server)
	Port                  int    `json:"port"`                  // Port server database
	Name                  string `json:"name"`                  // Nama database yang akan dituju
	Username              string `json:"username"`              // Username untuk akses masuk database
	Password              string `json:"password"`              // Password untuk akses database
	MaxOpenConnections    int    `json:"maxOpenConnections"`    // Jumlah maksimal koneksi dari aplikasi ke database yang aktif di saat bersamaan
	MaxLifeTimeConnection int    `json:"maxLifeTimeConnection"` // Batas umur maksimal sebuah koneksi database (apabila sudah lewat langsung ditutup)
	MaxIdleConnections    int    `json:"maxIdleConnections"`    // Jumlah maksimal koneksi database yang dibiarkan tetap ada meski sedang tidak dipakai (menganggur)
	MaxIdleTime           int    `json:"maxIdleTime"`           // Waktu maksimal sebuah koneksi dibiarkan saat sedang menganggur
}

// Init adalah fungsi yang dipanggil pertama kali jika ingin memuat/meload konfigurasi ke dalam variabel Config.
func Init() {
	// 1. Mencoba membaca file konfigurasi lokal yang bernama "config.json" di direktori utama "."
	err := util.BindFromJSON(&Config, "config.json", ".")

	// Jika mengalami error (contohnya: file tidak ditemukan)...
	if err != nil {
		// Menulis/mencatat info error ke console agar kita tahu file lokal gagal dibaca
		logrus.Infof("failed to bind config: %v", err)

		// 2. Karena cara pertama gagal, dicoba cara alternatif yakni membaca pengaturan dari Consul Server jarak jauh.
		// Alamat dan path menuju server Consul diambil dari environment/variabel sistem menggunakan os.Getenv
		err = util.BindFromConsul(&Config, os.Getenv("CONSUL_HTTP_URL"), os.Getenv("CONSUL_HTTP_PATH"))

		// Jika cara kedua ini juga mengalami kegagalan...
		if err != nil {
			// Maka sistem akan memicu panic (mengentikan aplikasi sepenuhnya) karena tak ada sumber asal konfigurasi
			panic(err)
		}
	}
}
