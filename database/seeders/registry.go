package seeders

import (
	"gorm.io/gorm"
)

// Registry adalah struct yang menyimpan instance database gorm
// untuk digunakan pada saat proses seeding data.
type Registry struct {
	db *gorm.DB
}

// ISeederRegistry adalah interface untuk mendefinisikan contract
// method Run() yang akan mengeksekusi semua seeder.
type ISeederRegistry interface {
	Run()
}

// NewSeederRegistry berfungsi untuk membuat dan mengembalikan
// instance baru dari Registry yang mengimplementasikan ISeederRegistry.
func NewSeederRegistry(db *gorm.DB) ISeederRegistry {
	return &Registry{db: db}
}

// Run berfungsi untuk menjalankan semua proses seeding yang ada.
// Method ini akan memanggil setiap fungsi seeder secara terpisah.
func (s *Registry) Run() {
	// Menjalankan seeder untuk tabel roles
	RunRoleSeeder(s.db)
	// Menjalankan seeder untuk tabel users
	RunUserSeeder(s.db)
}
