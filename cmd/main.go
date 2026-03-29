package cmd

import (
	"fmt"
	"net/http"
	"time"
	"user-service/common/response"
	"user-service/config"
	"user-service/constants"
	"user-service/controllers"
	"user-service/database/seeders"
	"user-service/domain/models"
	"user-service/middlewares"
	"user-service/repositories"
	"user-service/routes"
	"user-service/services"

	"github.com/didip/tollbooth"
	"github.com/didip/tollbooth/limiter"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

// 'command' adalah variabel yang mendefinisikan sebuah perintah bernama "serve" menggunakan library Cobra.
var command = &cobra.Command{
	// Kata kunci untuk menjalankan perintah ini dari terminal (misal: go run main.go serve).
	Use: "serve",
	// Deskripsi singkat dari perintah ini.
	Short: "Start the server",
	// 'Run' adalah fungsi utama yang akan dieksekusi saat orang memanggil perintah "serve".
	Run: func(c *cobra.Command, args []string) {
		// 1. Memuat file .env ke dalam memori aplikasi agar bisa dibaca. Tanda '_' berarti mengabaikan pesan error jika ada.
		_ = godotenv.Load()

		// 2. Menjalankan fungsi Inisialisasi dari package config untuk mengatur konfigurasi umum.
		config.Init()

		// 3. Membuka koneksi ke Database berdasarkan konfigurasi.
		db, err := config.InitDatabase()
		// Jika terjadi masalah gagal sambung database, aplikasi akan langsung crash / berhenti total (panic).
		if err != nil {
			panic(err)
		}

		// 4. Memuat zona waktu spesifik, dalam hal ini WIB (Waktu Indonesia Barat / Asia/Jakarta).
		loc, err := time.LoadLocation("Asia/Jakarta")
		// Jika sistem host tidak kenal nama zona waktu ini, hentikan aplikasi (panic).
		if err != nil {
			panic(err)
		}
		// Set waktu lokal di Golang agar merujuk ke zona waktu hasil load di atas.
		time.Local = loc

		// 5. Melakukan AutoMigrate, artinya sistem akan menciptakan/memperbarui tabel
		// di database secara otomatis berdasarkan kerangka (struct) dari model Role dan User.
		err = db.AutoMigrate(
			&models.Role{},
			&models.User{},
		)
		// Kalau gagal bikin tabel di DB, maka berhentikan aplikasi dengan panic.
		if err != nil {
			panic(err)
		}

		// 6. Menjalankan fungsi Seeder untuk menyuntikkan data awal (misalnya menambah Role Admin jika belum ada).
		seeders.NewSeederRegistry(db).Run()

		// 7. Proses Dependency Injection (Penyusunan komponen saling bergantung):
		// - Membuat bungkus koleksi Repository yang memiliki koneksi ke 'db'.
		repository := repositories.NewRepositoryRegistry(db)
		// - Membuat bungkus koleksi Service yang mengandalkan fungsi dari 'repository'.
		service := services.NewServiceRegistry(repository)
		// - Membuat bungkus koleksi Controller yang menggunakan logika di dalam 'service'.
		controller := controllers.NewControllerRegistry(service)

		// 8. Membuat objek aplikasi web (router) baru bawaan Gin dengan pengaturan standar (default).
		router := gin.Default()
		// Menitipkan Middleware HandlePanic ke jalurnya, agar kalau ada kodingan bermasalah aplikasi tidak mati total tapi merespons error JSON 500.
		router.Use(middlewares.HandlePanic())

		// 9. Mengatur balasan ketika orang mencari URL rute yang tidak terdaftar (404 Not Found).
		router.NoRoute(func(c *gin.Context) {
			// Kembalikan JSON berstatus 404
			c.JSON(http.StatusNotFound, response.Response{
				Status:  constants.Error,
				Message: fmt.Sprintf("Path %s", http.StatusText(http.StatusNotFound)),
			})
		})

		// 10. Mengatur rute dasar sistem ("/") sebagai rute sambutan/ping.
		router.GET("/", func(c *gin.Context) {
			// Mengirim balik status 200 OK beserta pesan selamat datang.
			c.JSON(http.StatusOK, response.Response{
				Status:  constants.Success,
				Message: "Welcome to User Service",
			})
		})

		// 11. Menambahkan Middleware CORS (Cross-Origin Resource Sharing) secara global.
		// Fungsi ini membolehkan atau memblokir akses ke API kita jika dipanggil dari halaman web berbasis domain lain (Frontend React/Vue).
		router.Use(func(c *gin.Context) {
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
			c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH")
			c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, x-service-name, x-request-at, x-api-key")
			if c.Request.Method == "OPTIONS" {
				c.AbortWithStatus(204)
				return
			}
			// Lanjut ke perutean aslinya.
			c.Next()
		})

		// 12. Mengkonfigurasi Rate Limiter (Tollbooth).
		// Tujuannya agar server tidak jebol terkena serangan bertubi-tubi (DDOS) dari 1 orang terus menerus.
		lmt := tollbooth.NewLimiter(
			// Berapa banyak maksimal permintaan per satuan waktu yang diijinkan sesuai file config (misal 10 requst).
			config.Config.RateLimiterMaxRequest,
			// Menyetel masa kedaluwarsa reset jatah request (misal 1 detik).
			&limiter.ExpirableOptions{
				DefaultExpirationTTL: time.Duration(config.Config.RateLimiterTimeSecond) * time.Second,
			})
		// Menempelkan rate limiter ini sebagai Middleware global di aplikasi.
		router.Use(middlewares.RateLimiter(lmt))

		// 13. Membuat sebuah kelompok awalan URL baru khusus API versi 1 (http://localhost/api/v1/...)
		group := router.Group("/api/v1")
		// Mendaftarkan seluruh kumpulan rute (URL) yang didefinisikan secara terpisah, di package 'routes'.
		route := routes.NewRouteRegistry(controller, group)
		// Menjalankan inisialisasi semua API kita yang mengarah ke Controller.
		route.Serve()

		// 14. Merangkai alamat 'Port' dengan mengambil nilai dari file config (misal 8080 menjadi ":8080").
		port := fmt.Sprintf(":%d", config.Config.Port)
		// 15. Baris penutup: Mulai jalankan server web (listening/menerima koneksi tiada henti) di port yang disiapkan!
		router.Run(port)
	},
}

// Run adalah pintu masuk awal (entry point) dari package 'cmd'.
// File `main.go` yang asli di root proyek akan memanggil `cmd.Run()` ini.
func Run() {
	// Memerintahkan modul Cobra untuk mengeksekusi semua logika dari tipe struktur Command di atas tadi.
	err := command.Execute()
	// Apabila terjadi error yang asalnya dari internal terminal, gagalkan (panic).
	if err != nil {
		panic(err)
	}
}
