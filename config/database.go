package config

import (
	"fmt"     // Package bawaan Go untuk memformat string (seperti menggabungkan teks)
	"net/url" // Package bawaan Go untuk memanipulasi URL (misal menyandikan karakter khusus)
	"time"    // Package bawaan Go untuk mengelola waktu dan durasi

	"gorm.io/driver/postgres" // Driver khusus agar GORM bisa berkomunikasi dengan database PostgreSQL
	"gorm.io/gorm"            // Library GORM (Object Relational Mapping) untuk mempermudah interaksi dengan database
)

// InitDatabase adalah fungsi untuk menginisialisasi/membuka koneksi ke database PostgreSQL.
// Fungsi ini mengembalikan object/koneksi *gorm.DB jika sukses, atau error jika gagal.
func InitDatabase() (*gorm.DB, error) {
	// Membaca variabel Config global yang sudah diisi sebelumnya (dari file config.go)
	config := Config

	// Menyandikan password (URL Encode) agar jika ada karakter khusus di password, tetap aman saat dimasukkan ke dalam link/URI
	encodedPassword := url.QueryEscape(config.Database.Password)

	// Menyusun string/URI koneksi (Connection String) dengan format khusus PostgreSQL.
	// Menggabungkan username, password, host, port, dan nama database.
	// sslmode=disable berarti kita tidak menggunakan koneksi aman SSL sementara waktu.
	uri := fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?sslmode=disable",
		config.Database.Username,
		encodedPassword,
		config.Database.Host,
		config.Database.Port,
		config.Database.Name,
	)

	// Membuka koneksi ke database PostgreSQL menggunakan URI yang sudah dibuat di atas, dibungkus dengan pengaturan standar GORM
	db, err := gorm.Open(postgres.Open(uri), &gorm.Config{})

	// Jika terjadi error saat mencoba terhubung ke database...
	if err != nil {
		// Maka kembalikan nilai nil (kosong) dan catatan pesan error-nya
		return nil, err
	}

	// db.DB() digunakan untuk mendapatkan objek koneksi asli bawaan Go (sql.DB) dari GORM.
	// Tujuannya agar kita bebas mengatur parameter tambahan (seperti batas jumlah dan waktu koneksi)
	sqlDB, err := db.DB()

	// Jika gagal mendapatkan objek koneksi aslinya...
	if err != nil {
		return nil, err
	}

	// Mengatur jumlah maksimal koneksi database yang menganggur (idle) yang dibiarkan tetap terbuka
	sqlDB.SetMaxIdleConns(config.Database.MaxIdleConnections)

	// Mengatur jumlah maksimal koneksi aplikasi ke database yang dapat terbuka dalam waktu bersamaan
	sqlDB.SetMaxOpenConns(config.Database.MaxOpenConnections)

	// Mengatur batas waktu maksimal sebuah koneksi boleh hidup/terhubung sebelum diputus (diubah jadi format Durasi Detik)
	sqlDB.SetConnMaxLifetime(time.Duration(config.Database.MaxLifeTimeConnection) * time.Second)

	// Mengatur batas waktu maksimal sebuah koneksi boleh dibiarkan menganggur sebelum ditutup
	sqlDB.SetConnMaxIdleTime(time.Duration(config.Database.MaxIdleTime) * time.Second)

	// Jika semuanya lancar, kembalikan objek koneksi database (db) dan artinya error-nya nil (kosong)
	return db, nil
}
