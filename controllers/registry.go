package controllers

import (
	controllers "user-service/controllers/user"
	"user-service/services"
)

// Registry adalah sebuah struct (tipe data berstruktur) yang bertugas
// sebagai "wadah" atau pusat pendaftaran dari semua controller yang ada di aplikasi ini.
type Registry struct {
	// Variabel service digunakan untuk menyimpan kumpulan service (logika bisnis)
	// agar nanti bisa disalurkan/diberikan ke setiap controller.
	service services.IServiceRegistry
}

// IControllerRegistry adalah sebuah interface (daftar kontrak kerja).
// Interface ini mendaftar fungsi-fungsi apa saja yang wajib dimiliki oleh struktur Registry d atas.
// Di sini baru ada satu fungsi, yaitu mengambil UserController.
type IControllerRegistry interface {
	GetUserController() controllers.IUserController
}

// NewControllerRegistry adalah fungsi pembuat (constructor).
// Fungsi ini berguna untuk membuat objek Registry baru saat aplikasi pertama kali dijalankan.
// Saat membuat Registry, kita harus memasukkan 'service' (sebagai dependensi) ke dalamnya.
func NewControllerRegistry(service services.IServiceRegistry) IControllerRegistry {
	return &Registry{service: service}
}

// GetUserController adalah fungsi milik struct Registry.
// Tugasnya adalah membuat dan mengembalikan (return) objek UserController baru agar siap dihubungkan ke rute HTTP.
// Fungsi ini juga otomatis menyuntikkan (inject) 'service' dari dalam Registry ke dalam UserController tersebut.
func (u *Registry) GetUserController() controllers.IUserController {
	return controllers.NewUserController(u.service)
}
