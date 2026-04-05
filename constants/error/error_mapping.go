package error

import (
	errField "field-service/constants/error/field"
	errFieldSchedule "field-service/constants/error/fieldschedule"
	errTime "field-service/constants/error/time"
)

// ErrMapping adalah fungsi sentral pendeteksi kebenaran sebuah nilai tipe pesan "error".
// Jika wujud error-nya merupakan bagian dari "error sah yang bisa ditebak/buatan kita pribadi di backend",
// Fungsi ini bernilai `true`. Apabila error benar-benar murni aneh akibat kegagalan server asli, maka ia bernilai `false`.
func ErrMapping(err error) bool {
	// 1. Kumpulkan seluruh "Array Penampung Error/Laci Error" milik masing-masing modul.
	var (
		GeneralErrors       = GeneralErrors
		FieldErrors         = errField.FieldErrors
		FieldScheduleErrors = errFieldSchedule.FieldScheduleErrors
		TimeErrors          = errTime.TimeErrors
	)

	// 2. Buat Array kosong baru ('allErrors') untuk menggabungkan semuanya jadi satu.
	allErrors := make([]error, 0)
	allErrors = append(allErrors, GeneralErrors...)
	allErrors = append(allErrors, FieldErrors...)
	allErrors = append(allErrors, FieldScheduleErrors...)
	allErrors = append(allErrors, TimeErrors...)

	// 3. Cek (loop/berputar) satu per satu pesan error tergabung di 'allErrors' ini.
	for _, item := range allErrors {
		// 4. Bandingkan! Apakah pesan teks error dari parameter ('err') SAMA dengan salah satu dari error buatan aseli kita ini?
		if err.Error() == item.Error() {
			// Jika sama (matched), beri kode 'true' agar penangkap error di response.go tahu kalau ini sah dan layak dikirim JSON ke Klien.
			return true
		}
	}

	// 5. Belum ditemukan di perulangan? Berarti errornya belum di-daftarkan developer (misterius). Kembalikan false.
	return false
}
