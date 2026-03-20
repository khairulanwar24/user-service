// Package services adalah tempat di mana kita meletakkan seluruh business logic (logika bisnis) dari aplikasi.
// Package ini menjadi jembatan antara Controller (yang memproses input) dan Repository (yang mengakses database).
package services

import (
	// Mengimpor folder repositories yang berisi kontrak (interface) untuk berinteraksi langsung dengan database.
	"user-service/repositories"
	// Mengimpor folder services/user yang berisi antarmuka (interface) dan logika spesifik untuk fitur User.
	// Kita memberikan alias 'services' (sebelum string import) untuk folder ini agar fungsinya mudah dipanggil di kode ini.
	services "user-service/services/user"
)

// Registry adalah sebuah struct (kumpulan variabel/data) yang berfungsi sebagai wadah untuk menyimpan semua dependency (ketergantungan).
// Di sini kita menyimpan 'repository' sebagai dependency utama agar menghubungkan service dengan database.
type Registry struct {
	repository repositories.IRepositoryRegistry
}

// IServiceRegistry adalah sebuah antarmuka (interface) yang mendefinisikan kontrak apa saja yang harus ada di dalam Registry.
// Interface ini menyatakan bahwa sesiapa yang menggunakannya harus menaati kontrak untuk memiliki fungsi GetUser() 
// yang mengembalikan nilai berupa services.IUserService.
type IServiceRegistry interface {
	GetUser() services.IUserService
}

// NewServiceRegistry adalah sebuah fungsi constructor (pembuat). Biasanya di Golang fungsi pembuat diawali dengan kata 'New'.
// Fungsi ini dipanggil pertama kali saat aplikasi berjalan untuk membuat instance (wujud nyata) dari Registry.
// Parameter yang diterima adalah 'repository', dan fungsi ini mengembalikan tipe antarmuka (interface) IServiceRegistry.
func NewServiceRegistry(repository repositories.IRepositoryRegistry) IServiceRegistry {
	// Mengembalikan pointer (alamat memori, ditandai dengan &) dari struct Registry yang sudah diisi dengan parameter repository.
	return &Registry{repository: repository}
}

// GetUser adalah sebuah method (fungsi yang menempel pada sebuah tipe data, dalam hal ini struct Registry).
// Tanda `(r *Registry)` disebut receiver. Ini berarti bahwa fungsi GetUser() adalah milik tipe daya (struct) Registry.
// Kita menggunakan pointer receiver (*) agar mengambil referensi dari struct aslinya, sehingga mencegah penyalinan data (copy memory).
// Fungsi ini bertugas untuk membuat dan memberikan layanan user (UserService) baru dengan bekal parameter repository.
func (r *Registry) GetUser() services.IUserService {
	// Memanggil fungsi constructor NewUserService dari package services/user
	return services.NewUserService(r.repository)
}
