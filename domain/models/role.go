package models

import "time"

// Role adalah representasi tabel 'roles' di database.
// Digunakan untuk menyimpan hak akses atau peran pengguna (misal: "admin", "user").
type Role struct {
	// ID adalah kunci utama (primary key) tabel, akan bertambah otomatis (auto increment)
	ID uint `gorm:"primaryKey;autoIncrement"`

	// Code adalah kode unik peran (misal: "ADM"). Maksimal 15 karakter dan tidak boleh kosong (not null)
	Code string `gorm:"varchar(15);not null"`

	// Name adalah nama peran yang lebih jelas (misal: "Administrator"). Maksimal 20 karakter dan tidak boleh kosong
	Name string `gorm:"varchar(20);not null"`

	// CreateAt adalah waktu saat data peran ini pertama kali dibuat di database
	CreateAt *time.Time

	// UpdateAt adalah waktu terakhir kali data peran ini diperbarui di database
	UpdateAt *time.Time
}
