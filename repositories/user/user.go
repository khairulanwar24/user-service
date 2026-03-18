package repositories

import (
	"context"
	"errors"
	errWrap "user-service/common/error"
	errConstant "user-service/constants/error"
	"user-service/domain/dto"
	"user-service/domain/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserRepository adalah struct yang menyimpan referensi ke koneksi database (db).
// Objek dengan tipe ini akan kita gunakan untuk berinteraksi dengan database untuk tabel users.
type UserRepository struct {
	db *gorm.DB
}

// IUserRepository mendefinisikan apa saja aksi (fungsi) yang bisa/harus dilakukan terhadap data User.
type IUserRepository interface {
	Register(context.Context, *dto.RegisterRequest) (*models.User, error)
	Update(context.Context, *dto.UpdateRequest, string) (*models.User, error)
	FindByUsername(context.Context, string) (*models.User, error)
	FindByEmail(context.Context, string) (*models.User, error)
	FindByUUID(context.Context, string) (*models.User, error)
}

// NewUserRepository digunakan untuk membuat instance baru dari UserRepository.
func NewUserRepository(db *gorm.DB) IUserRepository {
	return &UserRepository{db: db}
}

// Register berfungsi untuk menyimpan data pengguna baru ke dalam database.
func (r *UserRepository) Register(ctx context.Context, req *dto.RegisterRequest) (*models.User, error) {
	// 1. Mempersiapkan struktur object User baru menggunakan data dari input eksternal (req)
	user := models.User{
		UUID:        uuid.New(),
		Name:        req.Name,
		Username:    req.Username,
		Password:    req.Password,
		PhoneNumber: req.PhoneNumber,
		Email:       req.Email,
		RoleID:      req.RoleID,
	}

	// 2. Memerintahkan database (melalui GORM Create) untuk menyimpan data `user` tersebut
	err := r.db.WithContext(ctx).Create(&user).Error
	if err != nil {
		// Jika terjadi masalah saat menyimpan, kita rangkum errornya dan diberikan pesan ErrSQLError
		return nil, errWrap.WrapError(errConstant.ErrSQLError)
	}

	// 3. Mengembalikan data user yang baru saja berhasil di-simpan dan tanda tidak ada error (nil)
	return &user, nil
}

// Update digunakan untuk memperbarui atau mengubah data user yang sudah ada berdasarkan kolom UUID mereka.
func (r *UserRepository) Update(ctx context.Context, req *dto.UpdateRequest, uuid string) (*models.User, error) {
	// 1. Inisiasi data terbaru yang akan di-update (ditimpa) ke database
	user := models.User{
		Name:        req.Name,
		Username:    req.Username,
		Password:    *req.Password,
		PhoneNumber: req.PhoneNumber,
		Email:       req.Email,
	}

	// 2. Menjalankan perintah sql perbaruan: "cari data yang uuid-nya = `uuid`, lalu update pakai `user`"
	err := r.db.WithContext(ctx).
		Where("uuid = ?", uuid).
		Updates(&user).Error
	if err != nil {
		return nil, errWrap.WrapError(errConstant.ErrSQLError)
	}

	return &user, nil
}

// FindByUsername berfungsi untuk mencari satu baris (data) user di database berdasarkan username-nya.
func (r *UserRepository) FindByUsername(ctx context.Context, username string) (*models.User, error) {
	var user models.User
	// First: Mencari satu user pencocokan pertama dengan kondisi: username = ?
	// Preload: Mengambil (Join) tabel terkait "Role" agar role si pemain langsung otomatis kebaca
	err := r.db.WithContext(ctx).
		Preload("Role").
		Where("username = ?", username).
		First(&user).Error
	if err != nil {
		// Jika error disebabkan tidak temunya rekor (datanya tidak ada di database)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errConstant.ErrUserNotFound
		}
		// Jika ada error kesalahan teknis database lainnya
		return nil, errWrap.WrapError(errConstant.ErrSQLError)
	}
	return &user, nil // Berhasil, kembalikan user yang dicari
}

// FindByEmail berfungsi mencari data user di tabel database berdasarkan alamat email mereka.
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).
		Preload("Role").
		Where("email = ?", email).
		First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errConstant.ErrUserNotFound
		}
		return nil, errWrap.WrapError(errConstant.ErrSQLError)
	}
	return &user, nil
}

// FindByUUID berfungsi mencari satu pengguna berdasarkan kode Identifier uniknya (UUID), bukan ID numerik (auto increment).
func (r *UserRepository) FindByUUID(ctx context.Context, uuid string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).
		Preload("Role").
		Where("uuid = ?", uuid).
		First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errConstant.ErrUserNotFound
		}
		return nil, errWrap.WrapError(errConstant.ErrSQLError)
	}
	return &user, nil
}
