package error

import "errors"

// Kumpulan variabel pesan error standar yang umum (general) terjadi di semua jenis aplikasi web.
var (
	ErrInternalServerError = errors.New("internal server error")                   // Error jika server mogok/crash.
	ErrSQLError            = errors.New("database server failed to execute query") // Error jika database SQL mogok bekerja.
	ErrTooManyRequests     = errors.New("too many requests")                       // Error jika pengguna (client) melampaui batas kecepatan request / ngespam serangan.
	ErrUnauthorized        = errors.New("unauthorized")                            // Error jika klien sama sekali tidak memberikan kunci tiket masuk yang diakui.
	ErrInvalidToken        = errors.New("invalid token")                           // Error jika tiketnya (misal token JWT) kadaluarsa/rusak.
	ErrInvalidUploadFile   = errors.New("invalid upload file")                     // Error jika file yang dicoba upload tidak dikenali/rusak formatnya.
	ErrSizeTooBig          = errors.New("size too big")                            // Error jika ukuran file terlalu besar melampaui batas maksimal.
	ErrForbidden           = errors.New("forbidden")                               // Error jika user berhak login tapi level-nya (Role Admin/User) melarang ia masuk ke tempat tertentu.
)

// GeneralErrors adalah laci wujud array tempat menyimpan gabungan daftar jenis pesan error di atas.
// Array ini nanti bakal di-looping untuk diperiksa/dicocokkan ketika muncul sebuat masalah di sistem.
var GeneralErrors = []error{
	ErrInternalServerError,
	ErrSQLError,
	ErrTooManyRequests,
	ErrUnauthorized,
	ErrInvalidToken,
	ErrForbidden,
}
