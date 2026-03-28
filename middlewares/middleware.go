package middlewares

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"user-service/common/response"
	"user-service/config"
	"user-service/constants"
	services "user-service/services/user"

	errConstant "user-service/constants/error"

	"github.com/didip/tollbooth"
	"github.com/didip/tollbooth/limiter"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
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

// validateBearerToken adalah fungsi untuk memastikan token JWT yang dibawa user benar dan valid.
func validateBearerToken(c *gin.Context, token string) error {
	// 1. Cek dulu apakah di dalam token terdapat teks "Bearer".
	if !strings.Contains(token, "Bearer") {
		// Jika tidak ada teks "Bearer", berarti formatnya salah. Kembalikan error "Unauthorized".
		return errConstant.ErrUnauthorized
	}

	// 2. Ambil isi tokennya saja (buang tulisan "Bearer ") menggunakan fungsi bantuan yang kita buat di atas.
	tokenString := extractBearerToken(token)
	if tokenString == "" {
		// Jika isi tokennya kosong setelah dibersihkan, kembalikan error "Unauthorized".
		return errConstant.ErrUnauthorized
	}

	// 3. Menyiapkan variabel 'claims' sebagai wadah kosong.
	// Claims adalah data internal user yang dititipkan/disimpan di dalam token JWT (misal: ID user, email).
	claims := &services.Claims{}

	// 4. Proses membaca token menggunakan package pihak ketiga (golang-jwt).
	tokenJwt, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// 4a. Memeriksa apakah algoritma enkripsi token tersebut adalah algoritma HMAC.
		_, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			// Jika orang jahat mencoba memakai algoritma enkripsi abal-abal, tolak tokennya!
			return nil, errConstant.ErrInvalidToken
		}

		// 4b. Jika algoritmanya sah, kita berikan "Kunci Rahasia" (JwtSecretKey) ke package jwt.
		// Kunci rahasia ini dipakai package JWT untuk membuktikan keaslian token.
		jwtSecret := []byte(config.Config.JwtSecretKey)
		return jwtSecret, nil
	})

	// 5. Cek apakah ada error saat membaca token ATAU jika hasil pemeriksaan menyatakan token invalid (misal: kadaluwarsa).
	if err != nil || !tokenJwt.Valid {
		return errConstant.ErrUnauthorized
	}

	// 6. Menyimpan data user (dari claims) ke dalam isi Request (Context Golang biasa).
	// Ini dilakukan agar fungsi-fungsi di controller nanti gampang kalau mau ambil data spesifik user yang sedang login.
	userLogin := c.Request.WithContext(context.WithValue(c.Request.Context(), constants.UserLogin, claims.User))

	// 7. Mengganti isi Request gin (router) dengan Request baru yang sudah ditempeli info user tadi.
	c.Request = userLogin

	// 8. Menyimpan raw token ke dalam gin Context untuk bisa diakses cepat jika diperlukan.
	c.Set(constants.Token, token)

	// Semua aman! Kembalikan 'nil' (tidak ada error).
	return nil
}

// Authenticate adalah middleware keamanan utama kita.
// Rute (API) apapun yang melewati middleware ini hanya bisa diakses oleh pelanggan yang sudah login & punya API Key.
func Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		var err error

		// 1. Ambil nilai header yang bernama "Authorization" dari request pengguna yang masuk.
		token := c.GetHeader(constants.Authorization)
		if token == "" {
			// Jika sama sekali tidak dikirim token (kosong), lemparkan error tidak punya izin akses.
			responseUnauthorized(c, errConstant.ErrUnauthorized.Error())
			return
		}

		// 2. Jika token terisi, maka verifikasi apakah format JWT miliknya valid dan tidak rusak / dicurangi?
		err = validateBearerToken(c, token)
		if err != nil {
			// Jika hasilnya ada error, tolak permintaannya!
			responseUnauthorized(c, err.Error())
			return
		}

		// 3. Setelah lolos cek JWT, verifikasi lagi apakah API Key (kode aplikasi) si pengguna cocok dengan server kita?
		err = validateAPIKey(c)
		if err != nil {
			// Jika gagal di langkah kunci API, sekali lagi tolak permintaannya!
			responseUnauthorized(c, err.Error())
			return
		}

		// 4. Lulus penjagaan bertingkat ganda! Izinkan pelintas masuk / jalan terus ke Controller tujuan.
		c.Next()
	}
}
