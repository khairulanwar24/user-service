package repositories

import (
	repositories "user-service/repositories/user"

	"gorm.io/gorm"
)

// Registry adalah struct utama yang berfungsi sebagai wadah untuk koneksi database.
// Ini membantu kita mengumpulkan semua repository dalam satu tempat yang sama.
type Registry struct {
	db *gorm.DB
}

// IRepositoryRegistry adalah kontrak (interface) yang wajib dipenuhi oleh Registry.
// Di dalamnya terdapat fungsi-fungsi untuk mengambil repository spesifik, seperti GetUser.
type IRepositoryRegistry interface {
	GetUser() repositories.IUserRepository
}

// NewRepositoryRegistry adalah fungsi constructor (pembuat).
// Fungsi ini dipanggil pertama kali untuk membuat objek Registry baru dengan koneksi database `db` yang diberikan.
func NewRepositoryRegistry(db *gorm.DB) IRepositoryRegistry {
	return &Registry{db: db}
}

// GetUser adalah implementasi fungsi dari interface IRepositoryRegistry.
// Ketika dipanggil, ia akan membuat dan mengembalikan repository khusus untuk interaksi data User.
func (r *Registry) GetUser() repositories.IUserRepository {
	// Memanggil fungsi NewUserRepository di dalam folder/package layer `repositories/user`
	// dan meneruskan koneksi database (r.db).
	return repositories.NewUserRepository(r.db)
}
