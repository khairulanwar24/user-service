package models

import (
	"time"

	"github.com/google/uuid"
)

// User adalah representasi tabel 'users' di database.
// Menyimpan informasi lengkap dari setiap pengguna aplikasi.
type User struct {
	// ID adalah kunci utama (primary key) internal berbentuk angka, bertambah otomatis
	ID uint `gorm:"primaryKey;autoIncrement"`

	// UUID adalah kunci identitas acak yang unik (biasanya lebih aman digunakan untuk API/tampil ke luar)
	UUID uuid.UUID `gorm:"type:uuid;not null"`

	// Name adalah nama lengkap pengguna. Maksimal 100 karakter dan tidak boleh kosong
	Name string `gorm:"type:varchar(100);not null"`

	Username string `gorm:"type:varchar(20);not null"`

	// Pass adalah tempat menyimpan password pengguna yang telah dienkripsi. Maksimal 255 karakter
	Password string `gorm:"type:varchar(255);not null"`

	// PhoneNumber adalah nomor HP pengguna. Maksimal 15 karakter
	PhoneNumber string `gorm:"type:varchar(15);not null"`

	// Email adalah alamat email pengguna. Maksimal 100 karakter
	Email string `gorm:"type:varchar(100);not null"`

	// RoleID menyimpan ID dari tabel peran (menghubungkan User ini dengan Role tertentu)
	RoleID uint `gorm:"type:uint;not null"`

	// CreateAt mencatat waktu kapan akun ini didaftarkan
	CreateAt *time.Time

	// UpdateAt mencatat waktu kapan data akun ini terakhir diubah
	UpdateAt *time.Time

	// Role adalah relasi/hubungan dari database (Foreign Key).
	// Menandakan bahwa kolom RoleID di sini terhubung dengan kolom ID di tabel Role.
	// Jika Role diperbarui/dihapus, data User yang terikat juga akan terpengaruh (CASCADE).
	Role Role `gorm:"foreignKey:RoleID;references:id;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}
