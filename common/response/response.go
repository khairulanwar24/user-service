package response

import (
	"net/http"                                 // digunakan untuk mengambil text dari HTTP status code (misalnya 200 -> "OK")
	"user-service/constants"                   // berisi constant seperti Success dan Error
	errConstant "user-service/constants/error" // alias import untuk constant error

	"github.com/gin-gonic/gin" // framework web Gin
)

// Struct Response digunakan sebagai format response JSON yang akan dikirim ke client
type Response struct {
	Status  string      `json:"status"`          // status response (biasanya "success" atau "error")
	Message any         `json:"message"`         // pesan response (bisa string, object, dll)
	Data    interface{} `json:"data"`            // data yang dikirim ke client
	Token   *string     `json:"token,omitempty"` // token optional, jika nil tidak akan muncul di JSON
}

// Struct ini digunakan untuk mengirim parameter ke function HttpResponse
type ParamHTTPResp struct {
	Code    int          // HTTP status code (200, 400, 500 dll)
	Err     error        // error dari proses sebelumnya
	Message *string      // custom message jika ingin mengganti pesan default
	Gin     *gin.Context // context dari Gin untuk mengirim response
	Data    interface{}  // data yang akan dikirim ke client
	Token   *string      // token jika ada (biasanya untuk login)
}

// Function utama untuk mengirim HTTP response
func HttpResponse(param ParamHTTPResp) {

	// Jika tidak ada error
	if param.Err == nil {

		// kirim response success dalam bentuk JSON
		param.Gin.JSON(param.Code, Response{
			Status:  constants.Success,              // status success
			Message: http.StatusText(http.StatusOK), // mengambil text dari HTTP status (200 -> OK)
			Data:    param.Data,                     // data yang dikirim
			Token:   param.Token,                    // token jika ada
		})
	}

	// default message jika terjadi error
	message := errConstant.ErrInternalServerError.Error()

	// jika ada custom message dari parameter
	if param.Message != nil {
		message = *param.Message

		// jika ada error dari proses sebelumnya
	} else if param.Err != nil {

		// cek apakah error termasuk dalam mapping error yang sudah dibuat
		if errConstant.ErrMapping(param.Err) {

			// jika iya gunakan message dari error tersebut
			message = param.Err.Error()
		}
	}

	// kirim response error ke client
	param.Gin.JSON(param.Code, Response{
		Status:  constants.Error, // status error
		Message: message,         // pesan error
		Data:    param.Data,      // data tambahan jika ada
	})

	// mengakhiri function
	return
}
