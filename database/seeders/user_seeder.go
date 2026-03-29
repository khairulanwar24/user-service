package seeders

import (
	"user-service/constants"
	"user-service/domain/models"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// RunUserSeeder berfungsi untuk memasukkan data awal (seed data)
// ke dalam tabel users di database.
func RunUserSeeder(db *gorm.DB) {
	// Menghasilkan hash password dari nilai plain text "admin" menggunakan bcrypt
	// dengan tingkat keamanan (cost) bawaan (DefaultCost).
	password, _ := bcrypt.GenerateFromPassword([]byte("admin"), bcrypt.DefaultCost)

	// Membuat representasi objek User untuk dideklarasikan sebagai data awal (admin).
	user := models.User{
		UUID:        uuid.New(),
		Name:        "Administrator",
		Username:    "admin",
		Password:    string(password), // menyimpan password yang sudah di-hash
		PhoneNumber: "08123456789",
		Email:       "admin@gmail.com",
		RoleID:      constants.Admin,
	}

	// FirstOrCreate akan mencari data user berdasarkan kriteria pencarian (Username).
	// Jika user tersebut belum ada di database, GORM akan menyimpannya sebagai baris baru.
	err := db.FirstOrCreate(&user, models.User{Username: user.Username}).Error
	if err != nil {
		// Log error jika proses seeding user gagal dan menghentikan eksekusi (panic).
		logrus.Errorf("Failed to seed user: %v", err)
		panic(err)
	}

	// Memberikan pesan log informasi jika user berhasil di-seed (atau sudah diverifikasi ada).
	logrus.Infof("user %s successfully seeded", user.Username)

}
