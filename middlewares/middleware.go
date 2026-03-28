package middlewares

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"user-service/common/response"
	"user-service/config"
	"user-service/constants"

	errConstant "user-service/constants/error"

	"github.com/didip/tollbooth"
	"github.com/didip/tollbooth/limiter"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// HandlePanic adalah middleware pelindung aplikasi.
// Tujuannya agar jika terjadi "panic" (error fatal/crash tiba-tiba) di bagian kode manapun,
// server atau aplikasi backend kita tidak ikut mati/terhenti (ter-shutdown).
func HandlePanic() gin.HandlerFunc {
	return func(c *gin.Context) {
		// defer adalah keyword di Golang untuk "menunda" eksekusi fungsi.
		// Fungsi di dalam defer ini dijamin HANYA AKAN MAJU/DIJALANKAN
		// di paling akhir, persis sebelum fungsi utamanya selesai (atau jika terjadi panic).
		defer func() {
			// recover() berfungsi "menangkap" panic yang sedang terjadi.
			// Jika `r != nil`, berarti sungguhan baru saja terjadi panic/crash.
			if r := recover(); r != nil {
				// 1. Mencatat pesan error (log) ke sistem untuk keperluan debugging.
				logrus.Errorf("Recovered from panic: %v", r)
				// 2. Memberitahu user dengan mengirim respons HTTP 500 (Internal Server Error)
				// secara elegan, bukan tiba-tiba server memutus koneksi.
				c.JSON(http.StatusInternalServerError, response.Response{
					Status:  constants.Error,
					Message: errConstant.ErrInternalServerError.Error(),
				})
				// 3. Menghentikan jalannya request agar tidak diteruskan lebih jauh lagi ke controller.
				c.Abort()
			}
		}()
		// Jika tida ada panic, c.Next() memperbolehkan request berjalan terus ke controller / tujuan utamanya.
		c.Next()
	}
}

// RateLimiter adalah middleware untuk membatasi jumlah request (permintaan) dari satu user dalam waktu tertentu.
// Tujuannya adalah melindungi server dari serangan spam atau kelebihan beban (misal serangan DDoS).
func RateLimiter(lmt *limiter.Limiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. tollbooth.LimitByRequest akan mengecek apakah request ini sudah melewati batas/limit.
		err := tollbooth.LimitByRequest(lmt, c.Writer, c.Request)
		// 2. Jika `err != nil`, berarti user sudah melewati batas limit yang diizinkan (misal terlalu sering klik).
		if err != nil {
			// Tolak permintaannya dan kembalikan status HTTP 429 (Too Many Requests).
			c.JSON(http.StatusTooManyRequests, response.Response{
				Status:  constants.Error,
				Message: errConstant.ErrTooManyRequests.Error(),
			})
			// Hentikan request ini, jangan lanjutkan ke fungsi controller.
			c.Abort()
		}

		// 3. Jika aman (tidak kena limit), "persilakan lewat" dengan c.Next().
		c.Next()

	}
}

// extractBearerToken adalah fungsi bantuan (helper) untuk mengambil isi token saja.
// Biasanya token dikirim dalam format: "Bearer sdf89s7df98s7df...", di mana kata pertamanya selalu "Bearer ".
func extractBearerToken(token string) string {
	// 1. Membelah teks token berdasarkan spasi (" ") menjadi array kata-kata.
	// Contoh: "Bearer 12345" menjadi ["Bearer", "12345"].
	arrayToken := strings.Split(token, " ")

	// 2. Mengecek apakah potongannya benar ada 2 (kata "Bearer" dan isi tokennya).
	if len(arrayToken) == 2 {
		// Jika iya, kembalikan teks bagian kedua (indeks ke-1), yaitu isi tokennya.
		return arrayToken[1]
	}

	// Jika format tidak sesuai, kembalikan teks kosong.
	return ""
}

// responseUnauthorized adalah fungsi bantuan untuk mengirimkan pesan error "Tidak Diberi Akses" (Unauthorized).
func responseUnauthorized(c *gin.Context, message string) {
	// 1. Mengirim balasan format JSON berstatus 401 (Unauthorized) ke user.
	c.JSON(http.StatusUnauthorized, response.Response{
		Status:  constants.Error,
		Message: message,
	})

	// 2. Membatalkan kelanjutan rute agar tidak masuk ke Controller.
	c.Abort()
}

// validateAPIKey adalah fungsi untuk mengecek apakah kunci API (API Key) yang dikirim oleh user
// ke server ini sah (valid) dan tidak diretas orang lain.
func validateAPIKey(c *gin.Context) error {
	// 1. Mengambil data-data penting dari Header request yang dikirimkan user.
	apiKey := c.GetHeader(constants.XApiKey)           // Kunci enkripsi dari user.
	requestAt := c.GetHeader(constants.XRequestAt)     // Waktu request dikirim.
	serviceName := c.GetHeader(constants.XServiceName) // Nama layanan pengirim.

	// 2. Mengambil kunci rahasia (signature) dari berkas konfigurasi kita sendiri.
	signatureKey := config.Config.SignatureKey

	// 3. Merangkai data-data tadi menjadi satu pola string utuh sesuai susunan kesepakatan.
	validateKey := fmt.Sprintf("%s:%s:%s", serviceName, signatureKey, requestAt)

	// 4. Proses Hashing (menyandikan data menjadi kode acak).
	// Menggunakan algoritma SHA256 (salah satu algoritma pembuat sandi searah).
	hash := sha256.New()
	hash.Write([]byte(validateKey))

	// Mengubah hasil sandi (byte array) menjadi format teks biasa (huruf heksadesimal).
	resultHash := hex.EncodeToString(hash.Sum(nil))

	// 5. Verifikasi: Cocokkan apakah kode unik (API Key) dari header sama dengan hasil rumus buatan kita?
	if apiKey != resultHash {
		// Jika beda, berarti pengguna memakai API key/konfigurasi yang salah. Tolak akses!
		return errConstant.ErrUnauthorized
	}

	// Jika kodenya persis, berarti aksesnya SAH. Lanjutkan. (pengembalian nilai nil menandakan tak ada yang salah).
	return nil
}
