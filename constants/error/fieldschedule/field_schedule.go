package error

import "errors"

// Kumpulan variabel pesan error khusus untuk urusan Jadwal Lapangan (Field Schedule).
var (
	ErrFieldScheduleNotFound = errors.New("field schedule not found") // Error jika data jadwal lapangan yang dicari tidak ada di database.
	ErrFieldScheduleIsExist  = errors.New("field schedule already exist") // Error pencegah data kembar jika jadwal tersebut ternyata sudah pernah dibuat.
)

// FieldScheduleErrors array penampung kumpulan error Jadwal Lapangan untuk dicocokkan nanti oleh error_mapping.
var FieldScheduleErrors = []error{
	ErrFieldScheduleNotFound,
	ErrFieldScheduleIsExist,
}
