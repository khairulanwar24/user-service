package middlewares

import (
	"net/http"
	"user-service/common/response"
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
