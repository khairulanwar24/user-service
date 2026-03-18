package dto

import (
	"github.com/google/uuid"
)

// LoginRequest adalah format data yang kita harapkan saat user mengirim permintaan Login.
// Tag `json:"..."` berarti data ini akan dikirim dalam format JSON.
// Tag `validate:"required"` artinya field ini wajib diisi, tidak boleh kosong.
type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// UserResponse adalah format data profil user yang akan kita kembalikan (tampilkan) ke pengguna.
// Sengaja dibuat terpisah agar password dan data rahasia lainnya tidak ikut terkirim ke publik.
type UserResponse struct {
	UUID        uuid.UUID `json:"uuid"`        // ID unik dan aman dari pengguna
	Name        string    `json:"name"`        // Nama lengkap pengguna
	Username    string    `json:"username"`    // Username pengguna
	Email       string    `json:"email"`       // Email pengguna
	Role        string    `json:"role"`        // Nama perannya (misal: "Admin")
	PhoneNumber string    `json:"phoneNumber"` // Nomor HP pengguna
}

// LoginResponse adalah bentuk data atau respon yang akan diterima sistem/aplikasi setelah loginnya berhasil.
type LoginResponse struct {
	User  UserResponse `json:"user"`  // Berisi data profil singkat dari user yang berhasil login
	Token string       `json:"token"` // Berisi token JWT sebagai kunci rahasia untuk mengakses fitur lainnya
}

// RegisterRequest adalah data yang kita butuhkan dari form saat user ingin mendaftar (buat akun baru).
type RegisterRequest struct {
	Name            string `json:"name" validate:"required"`
	Username        string `json:"username" validate:"required"`
	Password        string `json:"password" validate:"required"`
	ConfirmPassword string `json:"confirmPassword" validate:"required"`
	Email           string `json:"email" validate:"required,email"`
	PhoneNumber     string `json:"phoneNumber" validate:"required"`
	RoleID          uint
}

type RegisterResponse struct {
	User UserResponse `json:"user"` // Mengembalikan informasi profil pengguna hasil registrasi
}

// UpdateRequest adalah format data yang dibutuhkan saat user ingin mengubah data profilnya.
type UpdateRequest struct {
	Name            string  `json:"name" validate:"required"`
	Username        string  `json:"username" validate:"required"`
	Password        *string `json:"password,omitempty"`
	ConfirmPassword *string `json:"confirmPassword,omitempty"`
	Email           string  `json:"email" validate:"required,email"`
	PhoneNumber     string  `json:"phoneNumber" validate:"required"`
	RoleID          uint
}
