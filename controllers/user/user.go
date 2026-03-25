package controllers

import (
	"net/http"
	errWrap "user-service/common/error"
	"user-service/common/response"
	"user-service/domain/dto"
	"user-service/services"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// UserController adalah sebuah struct (tipe data berstruktur) yang bertugas
// menghubungkan HTTP request dengan service (logika bisnis).
type UserController struct {
	// service menyimpan dependensi ke IServiceRegistry
	service services.IServiceRegistry
}

// IUserController adalah interface (daftar kontrak fungsi) yang menjelaskan
// aksi-aksi apa saja yang bisa dilakukan oleh UserController ini.
type IUserController interface {
	Login(*gin.Context)
	Register(*gin.Context)
	Update(*gin.Context)
	GetUserLogin(*gin.Context)
	GetUserByUUID(*gin.Context)
}

// NewUserControlle adalah fungsi inisialisasi (pembentuk) untuk menghasilkan objek UserController
func NewUserController(service services.IServiceRegistry) IUserController {
	return &UserController{service: service}
}

// Login adalah fungsi untuk memproses ketika pengguna mencoba masuk (login).
func (u *UserController) Login(ctx *gin.Context) {
	// 1. Membuat variabel kosong 'request' berdasarkan struct LoginRequest
	// untuk menampung data dari user (biasanya berisi email & password).
	request := &dto.LoginRequest{}

	// 2. Mengubah data JSON dari request body menjadi struct (objek Go).
	err := ctx.ShouldBindJSON(request)
	if err != nil {
		// Jika format JSON tidak sesuai, berikan respons error 400 (Bad Request).
		response.HttpResponse(response.ParamHTTPResp{
			Code: http.StatusBadRequest,
			Err:  err,
			Gin:  ctx,
		})

		return
	}

	// 3. Melakukan validasi isi data 'request', membuktikan apakah format
	// email sudah benar atau form tidak kosong.
	validate := validator.New()
	err = validate.Struct(request)
	if err != nil {
		// Jika data tidak valid, kembalikan error 422 (Unprocessable Entity).
		errmMessage := http.StatusText(http.StatusUnprocessableEntity)
		errResponse := errWrap.ErrValidationResponse(err)
		response.HttpResponse(response.ParamHTTPResp{
			Code:    http.StatusUnprocessableEntity,
			Message: &errmMessage,
			Data:    errResponse,
			Err:     err,
			Gin:     ctx,
		})
		return
	}
	// 4. Meneruskan data yang sudah divalidasi ke Service layer untuk mengecek
	// kebenaran akun ke database.
	user, err := u.service.GetUser().Login(ctx, request)
	if err != nil {
		// Jika password salah atau user tidak terdaftar, kembalikan error 400.
		response.HttpResponse(response.ParamHTTPResp{
			Code: http.StatusBadRequest,
			Err:  err,
			Gin:  ctx,
		})

		return
	}

	// 5. Apabila sukses, kembalikan respons 200 (OK) berisi data user berserta Token aksesnya.
	response.HttpResponse(response.ParamHTTPResp{
		Code:  http.StatusOK,
		Data:  user.User,
		Token: &user.Token,
		Gin:   ctx,
	})

}

// Register adalah fungsi untuk mendaftarkan akun pengguna baru.
func (u *UserController) Register(ctx *gin.Context) {
	// 1. Menyiapkan variabel 'request' untuk menampung data registrasi.
	request := &dto.RegisterRequest{}

	// 2. Mengikat (bind) JSON kiriman user dari body HTTP ke variabel 'request'.
	err := ctx.ShouldBindJSON(request)
	if err != nil {
		// Jika data yang dikirim tidak berbentuk JSON yang benar, tolak dengan error 400.
		response.HttpResponse(response.ParamHTTPResp{
			Code: http.StatusBadRequest,
			Err:  err,
			Gin:  ctx,
		})

		return
	}

	// 3. Melakukan pemeriksaan (validasi) kelengkapan aturan data.
	validate := validator.New()
	err = validate.Struct(request)
	if err != nil {
		// Apabila tidak memenuhi syarat, tolak dengan error validasi (422).
		errmMessage := http.StatusText(http.StatusUnprocessableEntity)
		errResponse := errWrap.ErrValidationResponse(err)
		response.HttpResponse(response.ParamHTTPResp{
			Code:    http.StatusUnprocessableEntity,
			Message: &errmMessage,
			Data:    errResponse,
			Err:     err,
			Gin:     ctx,
		})
		return
	}
	// 4. Meneruskan data pendaftaran yang sudah divalidasi ke Service layer
	// untuk disimpan di database.
	user, err := u.service.GetUser().Register(ctx, request)
	if err != nil {
		// Jika terjadi masalah dari basis data (misal email telah terdaftar), kembalikan error.
		response.HttpResponse(response.ParamHTTPResp{
			Code: http.StatusBadRequest,
			Err:  err,
			Gin:  ctx,
		})

		return
	}

	// 5. Jika sukses disimpan, kembalikan respons 200 (OK).
	response.HttpResponse(response.ParamHTTPResp{
		Code: http.StatusOK,
		Data: user.User,
		Gin:  ctx,
	})

}

// Update adalah fungsi untuk memperbarui profil data pengguna (contoh: mengganti nama).
func (u *UserController) Update(ctx *gin.Context) {
	// 1. Menyiapkan penampung untuk data update (perubahan baru).
	request := &dto.UpdateRequest{}
	// Menangkap ID (uuid) dari parameter URL rute (seperti /users/:uuid).
	uuid := ctx.Param("uuid")

	// 2. Menerjemahkan JSON ke dalam struct.
	err := ctx.ShouldBindJSON(request)
	if err != nil {
		// Jika JSON gagal dibaca.
		response.HttpResponse(response.ParamHTTPResp{
			Code: http.StatusBadRequest,
			Err:  err,
			Gin:  ctx,
		})
		return
	}

	// 3. Memvalidasi field-field di dalam data.
	validate := validator.New()
	err = validate.Struct(request)
	if err != nil {
		// Beri tahu klien bahwa data yang dikirim tidak lolos uji kelayakan (422).
		errMessage := http.StatusText(http.StatusUnprocessableEntity)
		errResponse := errWrap.ErrValidationResponse(err)
		response.HttpResponse(response.ParamHTTPResp{
			Code:    http.StatusUnprocessableEntity,
			Message: &errMessage,
			Data:    errResponse,
			Err:     err,
			Gin:     ctx,
		})
		return
	}

	// 4. Meneruskan data terbaru dan UUID sasaran ke service untuk diperbarui sistem databasenya.
	user, err := u.service.GetUser().Update(ctx, request, uuid)
	if err != nil {
		// Jika terjadi kegagalan dari sisi pembaruan data sistem (user tak ditemukan).
		response.HttpResponse(response.ParamHTTPResp{
			Code: http.StatusBadRequest,
			Err:  err,
			Gin:  ctx,
		})
		return
	}

	// 5. Jika sukses, kembali berikan response 200 (OK) menandakan bahwa proses telah beres.
	response.HttpResponse(response.ParamHTTPResp{
		Code: http.StatusOK,
		Data: user,
		Gin:  ctx,
	})
}

// GetUserLogin adalah fungsi untuk mengambil info rinci dari user yang sedang dalam state 'login'.
func (u *UserController) GetUserLogin(ctx *gin.Context) {
	// 1. Service akan memanggil fungsi GetUserLogin dengan meneruskan Context dari req (berguna buat ambil token yg lewat middleware).
	user, err := u.service.GetUser().GetUserLogin(ctx.Request.Context())
	if err != nil {
		// Kalau gagal, misalnya hak akses (tokennya) tidak berlaku, tampilkan pesan error.
		response.HttpResponse(response.ParamHTTPResp{
			Code: http.StatusBadRequest,
			Err:  err,
			Gin:  ctx,
		})
		return
	}

	// 2. Kalau berhasil terbaca, kembalikan respons 200 dengan data user rincinya.
	response.HttpResponse(response.ParamHTTPResp{
		Code: http.StatusOK,
		Data: user,
		Gin:  ctx,
	})
}

// GetUserByUUID adalah fungsi untuk menampilkan profil atau data pengguna tertentu berdasarkan ID (uuid) nya.
func (u *UserController) GetUserByUUID(ctx *gin.Context) {
	// 1. Meminta service mencarikan user yang UUID nya sama dengan nilai parameter rute "uuid".
	user, err := u.service.GetUser().GetUserByUUID(ctx.Request.Context(), ctx.Param("uuid"))
	if err != nil {
		// Jika salah format uuid, atau datanya hilang, tangkap errornya.
		response.HttpResponse(response.ParamHTTPResp{
			Code: http.StatusBadRequest,
			Err:  err,
			Gin:  ctx,
		})
		return
	}

	response.HttpResponse((response.ParamHTTPResp{
		Code: http.StatusOK,
		Data: user,
		Gin:  ctx,
	}))
}
