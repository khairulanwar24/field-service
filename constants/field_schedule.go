package constants

// Mendefinisikan tipe data alias 'FieldScheduleStatusName' untuk menampung teks berwujud string.
type FieldScheduleStatusName string
// Mendefinisikan tipe data alias 'FieldScheduleStatus' untuk menampung angka berwujud integer (int).
type FieldScheduleStatus int

const (
	// Available mewakili status angka 100 (Jadwal tersedia).
	Available FieldScheduleStatus = 100
	// Booked mewakili status angka 200 (Jadwal sudah dipesan).
	Booked    FieldScheduleStatus = 200

	// AvailableString adalah perwakilan teks untuk status jadwal tersedia.
	AvailableString FieldScheduleStatusName = "Available"
	// BookedString adalah perwakilan teks untuk status jadwal yang dibooking.
	BookedString    FieldScheduleStatusName = "Booked"
)

// mapFieldScheduleStatusIntToString adalah sebuah pemetaan (dictionary/map).
// Fungsinya: Jika kita masukkan angka status (contoh: 100), kita akan dapat teksnya ("Available").
var mapFieldScheduleStatusIntToString = map[FieldScheduleStatus]FieldScheduleStatusName{
	Available: AvailableString,
	Booked:    BookedString,
}

// mapFieldScheduleStatusStringToInt adalah kebalikan dari yang di atas.
// Fungsinya: Jika kita masukkan teks ("Available"), kita akan dapat angka statusnya (100).
var mapFieldScheduleStatusStringToInt = map[FieldScheduleStatusName]FieldScheduleStatus{
	AvailableString: Available,
	BookedString:    Booked,
}

// GetStatusString adalah sebuah "method" (fungsi yang menempel pada tipe data FieldScheduleStatus/integer).
// Cara kerjanya, mengambil teks string berdasarkan nilai integer yang dimiliki.
func (f FieldScheduleStatus) GetStatusString() FieldScheduleStatusName {
	return mapFieldScheduleStatusIntToString[f]
}

// GetStatusInt adalah "method" (fungsi yang menempel pada tipe data FieldScheduleStatusName/string).
// Cara kerjanya, mengambil angka integer berdasarkan teks string yang dimiliki.
func (f FieldScheduleStatusName) GetStatusInt() FieldScheduleStatus {
	return mapFieldScheduleStatusStringToInt[f]
}
