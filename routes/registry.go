package routes

import (
	"user-service/controllers"
	routes "user-service/routes/user"

	"github.com/gin-gonic/gin"
)

// Registry (Pusat Pendaftaran Rute) adalah sebuah struct yang berfungsi sebagai
// pangkalan utama untuk merangkai semua jalur rute aplikasi kita menjadi satu.
type Registry struct {
	// 'controller' adalah jembatan menuju logika bisnis yang akan dipanggil oleh rute.
	controller controllers.IControllerRegistry
	// 'group' adalah sekumpulan alamat rute dari framework Gin.
	group *gin.RouterGroup
}

// IRouteRegister adalah kontrak antar muka (interface) yang wajib ditaati.
// Aturan utamanya: siapa pun pembuat rute, mereka wajib punya fungsi Serve()
// untuk menghidupkan/menyajikan rutenya.
type IRouteRegister interface {
	Serve()
}

// NewRouteRegistry adalah fungsi pembuat (constructor).
// Di sini kita membuat "pangkalan utama" kita tadi dengan cara memasukkan
// jembatan Controller dan pengelompokan Router (RouterGroup).
func NewRouteRegistry(controller controllers.IControllerRegistry, group *gin.RouterGroup) IRouteRegister {
	return &Registry{controller: controller, group: group}
}

// Serve adalah fungsi untuk mulai "melayani" dan mendaftarkan semua kumpulan rute.
// Kalau aplikasinya semakin besar (misal ada rute produk, rute pesanan, dsb),
// pemasangannya akan dikumpulkan dan dipanggil berurutan di dalam fungsi ini.
func (r *Registry) Serve() {
	// Memanggil fungsi userRoute() (di bawah) lalu menyuruhnya Mendaftarkan rutenya dengan Run().
	r.userRoute().Run()
}

// userRoute adalah fungsi keamanan internal (pembantu).
// Fungsinya hanya merakit atau menciptakan wadah UserRoute baru,
// sambil mendistribusikan Controller dan RouterGroup dari wadah besar ini ke wadah spesifik user.
func (r *Registry) userRoute() routes.IUserRoute {
	return routes.NewUserRoute(r.controller, r.group)
}
