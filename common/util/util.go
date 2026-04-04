package util

import (
	"os"
	"reflect"
	"strconv"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	_ "github.com/spf13/viper/remote"
)

// BindFromJSON adalah fungsi untuk membaca file konfigurasi berformat JSON
// dan memasukannya (binding) ke dalam sebuah variabel/struct (dest).
// Parameter:
// - dest: variabel tujuan tempat menyimpan hasil bacaan konfigurasi.
// - filename: nama file konfigurasi (tanpa ekstensi .json).
// - path: lokasi/direktori tempat mencari file konfigurasi.
func BindFromJSON(dest any, filename, path string) error {
	// Membuat instance (objek) baru dari Viper untuk menangani konfigurasi.
	v := viper.New()

	// Memberi tahu Viper bahwa tipe file yang akan dibaca adalah JSON.
	v.SetConfigType("json")
	// Menambahkan direktori tempat file konfigurasi dicari.
	v.AddConfigPath(path)
	// Menentukan nama file konfigurasi yang ingin dibaca.
	v.SetConfigName(filename)

	// Mulai mencari dan membaca isi file konfigurasi.
	err := v.ReadInConfig()
	if err != nil {
		// Jika terjadi error saat membaca (misal file tidak ditemukan), kembalikan error tersebut.
		return err
	}

	// Memasukkan isi konfigurasi yang sudah dibaca ke dalam variabel 'dest'.
	err = v.Unmarshal(&dest)
	if err != nil {
		// Jika gagal, catat pesan error (log) lalu kembalikan error-nya.
		logrus.Errorf("failed to unmarshal: %v", err)
		return err
	}

	// Jika semua sukses diproses, kembalikan nilai nil (berarti tidak ada error).
	return nil
}

// SetEnvFromConsulKV adalah fungsi untuk mengambil nilai-nilai konfigurasi (key-value)
// dari Consul dan menyimpannya ke dalam Environment Variable sistem operasi (OS env).
// Parameter:
// - v: pointer ke instance Viper yang sudah berisi data konfigurasi.
func SetEnvFromConsulKV(v *viper.Viper) error {
	// Menyiapkan map kosong untuk menyimpan pasangan kunci (string) dan nilai (bebas/any).
	env := make(map[string]any)

	// Menguraikan (unmarshal) isi konfigurasi ke dalam wadah 'env'.
	err := v.Unmarshal(&env)
	if err != nil {
		// Jika gagal mengurai data, catat error (log) dan kembalikan error tersebut.
		logrus.Errorf("failed to unmarshal: %v", err)
		return err
	}

	// Melakukan perulangan untuk setiap data di dalam 'env'
	// 'k' adalah kunci (key), dan 'v' adalah nilainya (value).
	for k, v := range env {
		var (
			// reflect.ValueOf digunakan untuk mengecek tipe data asli dari 'v'
			valOf = reflect.ValueOf(v)
			// val adalah variabel penampung untuk menyimpan data setelah diubah menjadi teks (string)
			val string
		)

		// Memeriksa tipe data dari 'v' dan mengubahnya menjadi format string
		// agar bisa disimpan sebagai OS Environment Variable.
		switch valOf.Kind() {
		case reflect.String:
			// Jika tipenya sudah string, ambil nilainya langsung.
			val = valOf.String()
		case reflect.Int:
			// Jika tipenya angka bulat (integer), ubah jadi teks.
			val = strconv.Itoa(int(valOf.Int()))
		case reflect.Uint:
			// Jika tipenya integer tak bertanda (positif), ubah jadi teks.
			val = strconv.Itoa(int(valOf.Uint()))
		case reflect.Float32:
			// Jika tipenya bilangan pecahan (float), ubah jadi angka bulat lalu ke teks.
			val = strconv.Itoa(int(valOf.Float()))
		case reflect.Bool:
			// Jika tipenya boolean (true/false), ubah jadi teks "true" atau "false".
			val = strconv.FormatBool(valOf.Bool())
		}

		// Menyimpan pasangan nilai 'k' (kunci) dan 'val' (nilai string) ke dalam OS Environment Variable.
		err = os.Setenv(k, val)
		if err != nil {
			// Jika gagal menyimpan, catat error (log) lalu lapor.
			logrus.Errorf("failed to set env: %v", err)
			return err
		}
	}

	// Jika sukses, kembalikan 'nil' (tidak ada error).
	return nil
}

// BindFromConsul adalah fungsi untuk membaca konfigurasi dari server Consul secara remote.
// Parameter:
// - dest: variabel tempat menyimpan hasil konfigurasi.
// - endPoint: alamat/URL server Consul (misal: "localhost:8500").
// - path: path/kunci tempat data konfigurasi disimpan di Consul.
func BindFromConsul(dest any, endPoint, path string) error {
	// Membuat objek (instance) baru Viper.
	v := viper.New()
	// Menentukan bahwa format konfigurasi yang akan dibaca adalah JSON.
	v.SetConfigType("json")

	// Menambahkan provider (penyedia) konfigurasi remote, yaitu "consul".
	err := v.AddRemoteProvider("consul", endPoint, path)
	if err != nil {
		// Jika gagak menambahkan provider, catat error (log) dan return.
		logrus.Errorf("failed to add remote provider: %v", err)
		return err
	}

	// Mengambil/membaca konfigurasi dari server remote (Consul).
	err = v.ReadRemoteConfig()
	if err != nil {
		// Jika gagal mengambil dari remote, catat error (log) dan return.
		logrus.Errorf("failed to read remote config: %v", err)
		return err
	}

	// Memasukkan (unmarshal) hasil konfigurasi yang dibaca ke dalam variabel 'dest'.
	err = v.Unmarshal(&dest)
	if err != nil {
		// Jika gagal melakukan unmarshal, catat error (log) dan return.
		logrus.Errorf("failed to unmarshal: %v", err)
		return err
	}

	// Memanggil fungsi SetEnvFromConsulKV untuk menyimpan nilai konfigurasi
	// tersebut ke dalam Environment Variables sistem operasi.
	err = SetEnvFromConsulKV(v)
	if err != nil {
		// Jika gagal menyimpan ke env, catat error (log) dan return.
		logrus.Errorf("failed to set env from consul kv: %v", err)
		return err
	}

	// Jika sukses semua langkahnya, kembalikan 'nil' (berhasil/tanpa error).
	return nil
}
