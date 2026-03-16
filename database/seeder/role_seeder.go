package seeder

import (
	"user-service/domain/models"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// RunRoleSeeder berfungsi untuk memasukkan data awal (seed data)
// ke dalam tabel roles di database.
func RunRoleSeeder(db *gorm.DB) {
	// Mendefinisikan daftar role (peran) yang akan dimasukkan ke database
	roles := []models.Role{
		{
			Code: "ADMIN",
			Name: "Administrator",
		},
		{
			Code: "CUSTOMER",
			Name: "Customer",
		},
	}

	// Melakukan perulangan untuk setiap role yang sudah didefinisikan
	for _, role := range roles {
		// FirstOrCreate akan mencari role di database berdasarkan Code.
		// Jika tidak ada, maka GORM akan menyimpannya sebagai baris baru.
		err := db.FirstOrCreate(&role, models.Role{Code: role.Code}).Error
		if err != nil {
			// Jika terjadi error saat proses seed, log error dan hentikan eksekusi
			logrus.Errorf("Failed to seed role: %v", err)
			panic(err)
		}
		logrus.Infof("role %s successfully seeded", role.Code)
	}
}
