// Menentukan bahwa file ini berada di dalam package bernama 'services'.
package services

// Mengimpor daftar package atau library lain yang dibutuhkan oleh kode di bawah ini.
import (
	// Package bawaan Go untuk membawa data (seperti request ID, timeout) dari satu fungsi ke fungsi lainnya.
	"context"
	// Package bawaan Go untuk memanipulasi teks (string), misalnya merubah huruf besar ke kecil.
	"strings"
	// Package bawaan Go untuk bekerja dengan waktu dan tanggal.
	"time"
	// Package internal dari proyek ini untuk mengambil konfigurasi aplikasi (seperti rahasia JWT).
	"user-service/config"
	// Package internal dari proyek ini yang berisi nilai-nilai konstan (nilai yang tidak berubah).
	"user-service/constants"
	// Package internal yang berisi struktur data DTO (Data Transfer Object) untuk memindahkan data.
	"user-service/domain/dto"
	// Package internal yang berisi struktur data model (representasi tabel dalam database).
	"user-service/domain/models"
	// Package internal untuk mengakses lapisan media penyimpan (database).
	"user-service/repositories"

	// Package pembantu (alias) untuk error konstan (pesan-pesan error buatan sendiri).
	errConstant "user-service/constants/error"

	// Package pihak ketiga untuk menggunakan JSON Web Token (JWT) sebagai metode masuk (login).
	"github.com/golang-jwt/jwt/v5"
	// Package pihak ketiga untuk melakukan hal kriptografi, khususnya bcrypt (pengacakan password).
	"golang.org/x/crypto/bcrypt"
)

// UserService adalah sebuah struct (struktur data) yang membungkus komponen-komponen yang dibutuhkan service ini.
type UserService struct {
	// Variabel ini menyimpan akses ke kumpulan fungsi repository untuk berinteraksi dengan database.
	repository repositories.IRepositoryRegistry
}

// IUserService adalah sebuah antarmuka (interface) atau semacam kontrak/blueprint yang mendaftar fungsi-fungsi apa saja
// yang harus dimiliki oleh sebuah User Service.
type IUserService interface {
	// Fungsi untuk login. Menerima konteks dan data request login, lalu mengembalikan respons login (token) atau error.
	Login(context.Context, *dto.LoginRequest) (*dto.LoginResponse, error)
	// Fungsi untuk daftar pengguna baru. Menerima request register, dan mengembalikan data pengguna baru atau error.
	Register(context.Context, *dto.RegisterRequest) (*dto.RegisterResponse, error)
	// Fungsi untuk mengupdate profil pengguna berdasarkan ID acak (UUID).
	Update(context.Context, *dto.UpdateRequest, string) (*dto.UserResponse, error)
	// Fungsi untuk mendapatkan data profil dari pengguna yang saat ini sedang login secara aktif.
	GetUserLogin(context.Context) (*dto.UserResponse, error)
	// Fungsi untuk mendapatkan data pengguna secara spesifik dengan mencari dari ID acaknya.
	GetUserByUUID(context.Context, string) (*dto.UserResponse, error)
}

// Claims adalah semacam koper (payload) yang berisi informasi tentang user yang akan diselipkan dan disegel di dalam JWT (Token Web JSON).
type Claims struct {
	// Menaruh data response user ke dalam token.
	User *dto.UserResponse
	// Komponen standar bawaan library jwt untuk informasi kapan rilis, batas waktu token (expires at), dll.
	jwt.RegisteredClaims
}

// NewUserService adalah sebuah constructor function (fungsi pembangun) yang bertugas membuat/menciptakan objek UserService baru.
func NewUserService(repository repositories.IRepositoryRegistry) IUserService {
	// Mengembalikan objek/instance UserService dan mengisi tempat kosong 'repository' dengan yang dikirim melalui parameter.
	return &UserService{repository: repository}
}

// Login merupakan fungsi bawaan dari (u *UserService) yang ditugaskan memverifikasi dan mengelola sesi masuk pengguna.
func (u *UserService) Login(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error) {
	// Langkah 1: Mencari data lengkap pengguna (user) di database berdasarkan username dari inputan Login.
	user, err := u.repository.GetUser().FindByUsername(ctx, req.Username)
	// Jika terjadi error (misalnya tidak ketemu atau koneksi database putus), kembalikan error tersebut dan jangan kasih data apapun (nil).
	if err != nil {
		return nil, err
	}

	// Langkah 2: Membandingkan/mencocokkan (Compare) password acak yang ada di database dengan ketikan password biasa dari si pengguna.
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	// Jika password ternyata salah (berbeda), langsung kembalikan error.
	if err != nil {
		return nil, err
	}

	// Langkah 3: Jika berhasil melewati cek di atas, atur batas waktu kedaluwarsa token ini.
	// Didapat dari (waktu sekarang + durasi dari file konfigurasi JwtExpirationTime * 1 Menit)
	expirationTime := time.Now().Add(time.Duration(config.Config.JwtExpirationTime) * time.Minute).Unix()

	// Siapkan data user (DTO) yang akan dimasukkan ke dalam token (agar kelak tidak perlu panggil database terus untuk tau nama user dsb).
	data := &dto.UserResponse{
		UUID:        user.UUID,        // ID unik pengguna
		Name:        user.Name,        // Nama pengguna
		Username:    user.Username,    // Nama ID login (username)
		PhoneNumber: user.PhoneNumber, // Nomor HP
		Email:       user.Email,       // Email
		// Strings ToLower untuk mengubah huruf kapital pada Roles menjadi huruf kecil (admin -> admin).
		Role:        strings.ToLower(user.Role.Code),
	}

	// Langkah 4: Membuat lembar koper "Claims" tadi yang isinya Token beserta datanya, plus waktu kedaluwarsa.
	claims := &Claims{
		User: data,
		RegisteredClaims: jwt.RegisteredClaims{
			// Merubah format detik (Unix) ke format NumericDate untuk library JWT
			ExpiresAt: jwt.NewNumericDate(time.Unix(expirationTime, 0)),
		},
	}

	// Langkah 5: Bungkus dan ciptakan Token baru dengan metode pengaman tipe HS256
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// Gunakan kunci rahasia (JwtSecretKey) layaknya stempel gembok agar token sah dan aman dari ubahan (SignedString).
	tokenString, err := token.SignedString([]byte(config.Config.JwtSecretKey))

	// Jika proses penstempelan token gagal, maka kembalikan balasan kosong berupa error.
	if err != nil {
		return nil, err
	}

	// Langkah 6: Siapkan bingkisan respons berisi profil user dan cetakan token bentuk huruf (string)  
	response := &dto.LoginResponse{
		User:  *data,       // Value dari isi variable data (pointer yang diambil valuenya)
		Token: tokenString, // String kombinasi karakter JWT
	}

	// Kembalikan responsnya jika berhasil, dengan error nil (yang berarti sukses tanpa ada error satupun).
	return response, nil
}

// isUsernameExist adalah fungsi bantuan (helper/utility) privat milik u *UserService untuk mengecek keberadaan suatu username saja.
func (u *UserService) isUsernameExist(ctx context.Context, username string) bool {
	// Mencari user ke database dengan nama (username) tersebut.
	user, err := u.repository.GetUser().FindByUsername(ctx, username)
	// Kalau error terjadi (yang berarti biasanya tidak ketemu / error db), anggap username tersebut 'false' (tidak dipakai orang/tidak eksis).
	if err != nil {
		return false
	}

	// Tetapi jika datanya tidak kosong (tidak bernilai nil / tidak terhapus), berarti "sudah eksis / benar ada".
	if user != nil {
		// maka kembalikan nilai bool "true"
		return true
	}

	// Default balikan, asumsikan belum ada pengguna dengan username itu.
	return false
}

// isEmailExist adalah fungsi bantuan mirip fungsi diatas, tetapi khusus mencari kolom email saja.
func (u *UserService) isEmailExist(ctx context.Context, email string) bool {
	// Minta database dicarikan user berdasarkan email ini
	user, err := u.repository.GetUser().FindByEmail(ctx, email)
	// Apabila error, anggap tidak ditemukan dan belum dipakai emailnya.
	if err != nil {
		return false
	}

	// Jika variabel berhasil terisi dengan profil user, berarti "Benar, Email itu sudah dipakai".
	if user != nil {
		return true
	}

	// Secara utuh, kembalikan posisi email tidak dipakai (false).
	return false
}

// Register adalah metode fungsi yang mengatur logika layanan daftar orang baru
func (u *UserService) Register(ctx context.Context, req *dto.RegisterRequest) (*dto.RegisterResponse, error) {
	// Langkah 1: Kunci/Samarkan (GenerateFromPassword) password yang di input oleh pengguna (format byte) dengan cost bawaan (DefaultCost).
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	// Apabila komputer gagal mengacak password, gagalkan pendaftaran.
	if err != nil {
		return nil, err
	}

	// Langkah 2: Verifikasi dulu ke database pakai fungsi "isUsernameExist",  kalau bernilai "true" berarti nama sudah dipinang orang.
	if u.isUsernameExist(ctx, req.Username) {
		// Stop, gagalkan! kembalikan error yang mengatakan 'Username sudah ada/terpakai'
		return nil, errConstant.ErrUsernameExist
	}

	// Langkah 3: Sama seperti username, pastikan email juga tidak ada yang memakai duplikat.
	if u.isEmailExist(ctx, req.Email) {
		// Gagalkan dan lempar error konstanta spesial 'Email sudah ada'
		return nil, errConstant.ErrEmailExist
	}

	// Langkah 4: Cek kesamaan input tipe ketik manual antara kolom 'Password' vs kolom 'ConfirmPassword'.
	if req.Password != req.ConfirmPassword {
		// Kalau tulisan hurufnya tidak sama berarti salah ketik ulangnya, keluarkan error konstanta ketidakcocokan.
		return nil, errConstant.ErrPasswordDoesNotMatch
	}

	// Langkah 5: Proses kirim ke layer dasar Repository agar database memproses aksi daftar dan menyimpannya (Insert).
	user, err := u.repository.GetUser().Register(ctx, &dto.RegisterRequest{
		Name:        req.Name,                   // Menyimpan input Nama
		Username:    req.Username,               // Menyimpan input Username login
		Password:    string(hashedPassword),     // WAJIB: Menyimpan password yang sudah di Gembok/Hash (Dikonversi ke text string)
		PhoneNumber: req.PhoneNumber,            // Nomor Telepon 
		Email:       req.Email,                  // Alamat surat eletronik
		RoleID:      constants.Customer,         // Tentukan semua pendaftar baru perannya (Role) sebagai tipe "Customer" 
	})

	// Cek apakah database bermasalah sewaktu menyimpan
	if err != nil {
		return nil, err
	}

	// Langkah 6: Jika registrasi sukses dari database, bungkus balasan atau responsenya
	response := &dto.RegisterResponse{
		// Cuma ambil atau keluarkan informasi dasar non-rahasia supaya dilihat oleh pengguna. (Tanpa password)
		User: dto.UserResponse{
			UUID:        user.UUID,
			Name:        user.Name,
			Username:    user.Username,
			PhoneNumber: user.PhoneNumber,
			Email:       user.Email,
		},
	}

	// Kirimkan balik hasil Response dari database
	return response, nil
}

// Update digunakan saat profil data diri si user ingin diubah isinya.
func (u *UserService) Update(ctx context.Context, request *dto.UpdateRequest, uuid string) (*dto.UserResponse, error) {
	// Kita akan menyiapkan "keranjang kosong" atau variabel penampungan di awal supaya teratur.
	var (
		password                  string          // Keranjang wadah penampung teks dari password setelah diubah 
		checkUsername, checkEmail *models.User    // Keranjang tipe data ceklis saat periksa nama/email
		hashedPassword            []byte          // Keranjang khusus wujud bilangan byte biner
		user, userResult          *models.User    // Keranjang isi data pengguna satu dari tarikan, satu pasca revisi
		err                       error           // Keranjang khusus penanganan jika terjadi masalah (error)
		data                      dto.UserResponse// Keranjang khusus kemasan yang ramah pengguna
	)

	// Langkah 1: Tarik dan muat profil si pengguna sesuai UUID nya ke dalam variabel user.
	user, err = u.repository.GetUser().FindByUUID(ctx, uuid)
	// Kalau datanya nggak ada (user tidak valid/dihapus?), berhenti!
	if err != nil {
		return nil, err
	}

	// Langkah 2: Pengecekan sebelum merubah kolom Username.
	// Tanya ke helper isUsernameExist() apakah username bari yang diminta sudah ada (true/false).
	isUsernameExist := u.isUsernameExist(ctx, request.Username)
	// Kalau nilai (true) berarti dipakai orang lain, TAPI... pastikan apakah username-nya dia sendiri atau beda. (jika dia ngga rubah kolomnya abaikan) 
	if isUsernameExist && user.Username != request.Username {
		// Kalau beda dengan datanya yang lama, cek pakai databasenya langsung juga untuk jamin konsistensi.
		checkUsername, err = u.repository.GetUser().FindByUsername(ctx, request.Username)
		if err != nil {
			return nil, err
		}

		// Kalau cek username berwujud, ada, dan eksis. Tolak penggantian nama.
		if checkUsername != nil {
			return nil, errConstant.ErrUsernameExist
		}
	}

	// Langkah 3: Sama seperti pengecekan sebelumnya, kini gilirannya email.
	// Cari kebenaran eksistensi emailnya.
	isEmailExist := u.isEmailExist(ctx, request.Email)
	// Jika statusnya True (sudah ada) dan email target tidak sama (berbeda) dengan alamat email saat ini.
	if isEmailExist && user.Email != request.Email {
		// Telusuri profil di database mana yang nyangkut.
		checkEmail, err = u.repository.GetUser().FindByEmail(ctx, request.Email)
		if err != nil {
			return nil, err
		}

		// Bila ketemu pemilik email lain tersebut, gagalkan perubahan (email tidak boleh kembar antarakun).
		if checkEmail != nil {
			return nil, errConstant.ErrEmailExist
		}
	}

	// Langkah 4: Opsional (Jika si user berniat mengganti Passwordnya yang lama)
	// Apabila request Password itu isinya tidak sama dengan kosong (Nil)
	if request.Password != nil {
		// Periksa kecocokan masukan dengan konfirmasi.
		// (tanda * berfungsi membaca isi asli data dari pointer)
		if *request.Password != *request.ConfirmPassword {
			return nil, errConstant.ErrPasswordDoesNotMatch
		}

		// Karena ingin mengubah jadinya kita Generate (Enkripsi baru) kata sandi tsb. 
		hashedPassword, err = bcrypt.GenerateFromPassword([]byte(*request.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err // Return ketika gagal Enkripsi.
		}
		// Masukkan password terenkripsi kedalam keranjang password bentuk String.
		password = string(hashedPassword)
	}

	// Langkah 5: Tahap final ke Database melalui repository Update untuk mengubah entitas/baris kolom secara resmi.
	userResult, err = u.repository.GetUser().Update(ctx, &dto.UpdateRequest{
		Name:        request.Name,
		Username:    request.Username,
		Password:    &password, // Dikirim dengan petunjuk rujukan (pointer &)
		Email:       request.Email,
		PhoneNumber: request.PhoneNumber,
	}, uuid) // Memberikan spesifik ID-nya (UUID)

	// Pantau kalau error
	if err != nil {
		return nil, err
	}

	// Langkah 6: Konstruksi atau bungkus lagi pembalasan menggunakan dto bentuk Response
	data = dto.UserResponse{
		UUID:        userResult.UUID,          // Memberikan data userResult terbaru ke dto pembungkus
		Name:        userResult.Name,
		Username:    userResult.Username,
		PhoneNumber: userResult.PhoneNumber,
		Email:       userResult.Email,
	}

	// Lempar variabel data (harus diambil alamat ruang memorinya memakai &) lalu kirim.
	return &data, nil
}

// GetUserLogin difungsikan untuk mendapatkan profil akun orang (User) yang tengah login dan menggunakan API saat ini.
func (u *UserService) GetUserLogin(ctx context.Context) (*dto.UserResponse, error) {
	// Menyiapkan variabel keranjang...
	var (
		// Ingat kita bawa Context? di dalam aliran context ini sudah disematkan identitas login dari tingkat Middleware. 
		// Context.Value itu ditarik memakai kata kunci (constants.UserLogin) dan di konfirmasikan tipe objectnya (type-assertion) dari DTO UserResponse.
		userLogin = ctx.Value(constants.UserLogin).(*dto.UserResponse)
		data      dto.UserResponse // Ini wadah rapinya.
	)

	// Tata dan masuk-masukkan hasil bongkaran dari Context yang menyimpan JWT ke DTO yang benar.
	data = dto.UserResponse{
		UUID:        userLogin.UUID,
		Name:        userLogin.Name,
		Username:    userLogin.Username,
		PhoneNumber: userLogin.PhoneNumber,
		Email:       userLogin.Email,
		Role:        userLogin.Role, // Di sini perannya juga di oper
	}

	// Kembalikan DTO pembungkus ini sebagai balikan fungsi akhir.
	return &data, nil
}

// GetUserByUUID adalah layanan tunggal mencari dan mendapatkan keseluruhan atribut seorang User cuma dengan bermodal UUID-nya.
func (u *UserService) GetUserByUUID(ctx context.Context, uuid string) (*dto.UserResponse, error) {
	// Temukan di database dengan perantara panggil FindByUUID dari repository (Repositori berinteraksi langsung SQL)
	user, err := u.repository.GetUser().FindByUUID(ctx, uuid)
	// Jika terjadi error tidak ketemu atau error SQL, munculkan dan kembalikan "nil" (Tidak ada datanya)
	if err != nil {
		return nil, err
	}

	// Buatkan pemetaan bungkus rapi dengan standar format kembalian dari UserService (UserResponse DTO)
	data := dto.UserResponse{
		UUID:        user.UUID,
		Name:        user.Name,
		Username:    user.Username,
		PhoneNumber: user.PhoneNumber,
		Email:       user.Email,
	}

	// Mengembalikan pointer memori alamat DTO "data" (&) menuju lapisan yang manggil layanan ini (Cth: Controller)
	return &data, nil
}
