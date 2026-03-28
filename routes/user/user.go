package routes

import (
	"user-service/controllers"
	"user-service/middlewares"

	"github.com/gin-gonic/gin"
)

// UserRoute adalah struktur data yang mewadahi alat-alat kebutuhan rute.
// Di dalamnya ada 'controller' (sebagai penyedia logika akhir) dan 'group'
// (sebagai pengelompok URL/rute bawaan framework Gin).
type UserRoute struct {
	controller controllers.IControllerRegistry
	group      *gin.RouterGroup
}

// IUserRoute adalah interface kontrak yang mewajibkan struktur UserRoute
// di atas untuk memiliki setidaknya satu fungsi utama, yaitu Run().
type IUserRoute interface {
	Run()
}

// NewUserRoute adalah fungsi pembuat (constructor) untuk menyuntikkan
// dependensi Controller dan RouterGroup ke dalam objek UserRoute.
func NewUserRoute(controller controllers.IControllerRegistry, group *gin.RouterGroup) IUserRoute {
	return &UserRoute{controller: controller, group: group}
}

// Run adalah inti dari file ini. Di sinilah kita mendata seluruh jalur API (Endpoint).
func (u *UserRoute) Run() {
	// 1. Membuat awalan kelompok rute. Semua URL di bawah ini akan diimbuhi dengan "/auth"
	// Misalnya: www.domain-kita.com/auth/login
	group := u.group.Group("/auth")

	// 2. Mendaftarkan rute dengan metode GET untuk alamat "/auth/user".
	// Rute ini dilindungi oleh middlewares.Authenticate() yg mewajibkan user sedang login (punya tiket JWT).
	// Jika pemeriksaannya aman, request dikirimkan ke tujuan akhir yaitu GetUserLogin di Controller.
	group.GET("/user", middlewares.Authenticate(), u.controller.GetUserController().GetUserLogin)

	// 3. Mendaftarkan rute GET dengan URL fleksibel "/auth/:uuid" (contoh: /auth/qweqwe-123).
	// URL ini untuk mendapatkan sebuah dokumen profil user. Rute ini juga diamankan dengan Authenticate.
	group.GET("/:uuid", middlewares.Authenticate(), u.controller.GetUserController().GetUserByUUID)

	// 4. Mendaftarkan rute POST untuk URL pendaftaran masuk "/auth/login".
	// Karena ini gerbang depan, rute ini dibiarkan terbuka (PUBLIK) tanpa middleware keamanan Authenticate.
	group.POST("/login", u.controller.GetUserController().Login)

	// 5. Mendaftarkan rute POST "/auth/register" (untuk membuat akun baru pertama kali).
	// Rute ini juga pastinya tetap rute publik/bebas seperti layaknya halaman daftar.
	group.POST("/register", u.controller.GetUserController().Register)

	// 6. Mendaftarkan rute PUT (metode khusus HTTP untuk menyunting/mengubah data) ke "/auth/:uuid".
	// Khusus untuk kegiatan ubah data ini kembali dikunci dan diteruskan ke fungsi Update Controller.
	group.PUT("/:uuid", middlewares.Authenticate(), u.controller.GetUserController().Update)
}
