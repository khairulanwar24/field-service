package services

import (
	"context"
	"field-service/common/util"
	"field-service/constants"
	errFieldSchedule "field-service/constants/error/fieldschedule"
	"field-service/domain/dto"
	"field-service/domain/models"
	"field-service/repositories"
	"fmt"
	"github.com/google/uuid"
	"time"
)

// FieldScheduleService mengelola jadwal penggunaan lapangan.
// Jadwal ini berisi informasi kapan lapangan tersedia (available) atau sudah dipesan (booked).
type FieldScheduleService struct {
	repository repositories.IRepositoryRegistry
}

// IFieldScheduleService mendefinisikan kontrak untuk layanan jadwal lapangan.
type IFieldScheduleService interface {
	GetAllWithPagination(context.Context, *dto.FieldScheduleRequestParam) (*util.PaginationResult, error)
	GetAllByFieldIDAndDate(context.Context, string, string) ([]dto.FieldScheduleForBookingResponse, error)
	GetByUUID(context.Context, string) (*dto.FieldScheduleResponse, error)
	GenerateScheduleForOneMonth(context.Context, *dto.GenerateFieldScheduleForOneMonthRequest) error
	Create(context.Context, *dto.FieldScheduleRequest) error
	Update(context.Context, string, *dto.UpdateFieldScheduleRequest) (*dto.FieldScheduleResponse, error)
	UpdateStatus(context.Context, *dto.UpdateStatusFieldScheduleRequest) error
	Delete(context.Context, string) error
}

// NewFieldScheduleService membuat instance baru dari FieldScheduleService.
func NewFieldScheduleService(repository repositories.IRepositoryRegistry) IFieldScheduleService {
	return &FieldScheduleService{repository: repository}
}

// GetAllWithPagination mengambil semua jadwal dengan pagination.
func (f *FieldScheduleService) GetAllWithPagination(
	ctx context.Context,
	param *dto.FieldScheduleRequestParam,
) (*util.PaginationResult, error) {
	// 1. Ambil data jadwal dari repository.
	fieldSchedules, total, err := f.repository.GetFieldSchedule().FindAllWithPagination(ctx, param)
	if err != nil {
		return nil, err
	}

	// 2. Mapping ke DTO response.
	fieldScheduleResults := make([]dto.FieldScheduleResponse, 0, len(fieldSchedules))
	for _, schedule := range fieldSchedules {
		fieldScheduleResults = append(fieldScheduleResults, dto.FieldScheduleResponse{
			UUID:         schedule.UUID,
			FieldName:    schedule.Field.Name,
			Date:         schedule.Date.Format("2006-01-02"),
			PricePerHour: schedule.Field.PricePerHour,
			Status:       schedule.Status.GetStatusString(), // Mengonversi enum status ke string (Available/Booked).
			Time:         fmt.Sprintf("%s - %s", schedule.Time.StartTime, schedule.Time.EndTime),
			CreatedAt:    schedule.CreatedAt,
			UpdatedAt:    schedule.UpdatedAt,
		})
	}

	// 3. Siapkan data pagination.
	pagination := &util.PaginationParam{
		Count: total,
		Limit: param.Limit,
		Page:  param.Page,
		Data:  fieldScheduleResults,
	}

	response := util.GeneratePagination(*pagination)
	return &response, nil
}

// convertMonthName adalah helper untuk mengubah nama bulan dari format Inggris ke Indonesia (singkat).
func (f *FieldScheduleService) convertMonthName(inputDate string) string {
	date, err := time.Parse(time.DateOnly, inputDate)
	if err != nil {
		return ""
	}

	indonesiaMonth := map[string]string{
		"Jan": "Jan",
		"Feb": "Feb",
		"Mar": "Mar",
		"Apr": "Apr",
		"May": "Mei",
		"Jun": "Jun",
		"Jul": "Jul",
		"Aug": "Agu",
		"Sep": "Sep",
		"Oct": "Okt",
		"Nov": "Nov",
		"Dec": "Des",
	}

	formattedDate := date.Format("02 Jan")
	day := formattedDate[:3]
	month := formattedDate[3:]
	formattedDate = fmt.Sprintf("%s %s", day, indonesiaMonth[month])
	return formattedDate
}

// GetAllByFieldIDAndDate mengambil jadwal untuk satu lapangan tertentu pada tanggal tertentu.
// Digunakan biasanya untuk tampilan booking di sisi user.
func (f *FieldScheduleService) GetAllByFieldIDAndDate(
	ctx context.Context,
	uuid, date string,
) ([]dto.FieldScheduleForBookingResponse, error) {
	// 1. Cari dulu data lapangannya.
	field, err := f.repository.GetField().FindByUUID(ctx, uuid)
	if err != nil {
		return nil, err
	}

	// 2. Ambil semua jadwal berdasarkan ID lapangan (integer) dan tanggal.
	fieldSchedules, err := f.repository.GetFieldSchedule().FindAllByFieldIDAndDate(ctx, int(field.ID), date)
	if err != nil {
		return nil, err
	}

	// 3. Mapping hasil ke DTO khusus booking.
	fieldScheduleResults := make([]dto.FieldScheduleForBookingResponse, 0, len(fieldSchedules))
	for _, fieldSchedule := range fieldSchedules {
		pricePerHour := float64(fieldSchedule.Field.PricePerHour)
		startTime, _ := time.Parse("15:04:05", fieldSchedule.Time.StartTime)
		endTime, _ := time.Parse("15:04:05", fieldSchedule.Time.EndTime)
		
		fieldScheduleResults = append(fieldScheduleResults, dto.FieldScheduleForBookingResponse{
			UUID:         fieldSchedule.UUID,
			PricePerHour: util.RupiahFormat(&pricePerHour),
			Date:         f.convertMonthName(fieldSchedule.Date.Format("2006-01-02")),
			Status:       fieldSchedule.Status.GetStatusString(),
			Time:         fmt.Sprintf("%s - %s", startTime.Format("15:04"), endTime.Format("15:04")),
		})
	}

	return fieldScheduleResults, nil
}

// GetByUUID mengambil detail satu jadwal berdasarkan UUID.
func (f *FieldScheduleService) GetByUUID(ctx context.Context, uuid string) (*dto.FieldScheduleResponse, error) {
	fieldSchedule, err := f.repository.GetFieldSchedule().FindByUUID(ctx, uuid)
	if err != nil {
		return nil, err
	}

	response := dto.FieldScheduleResponse{
		UUID:         fieldSchedule.UUID,
		FieldName:    fieldSchedule.Field.Name,
		PricePerHour: fieldSchedule.Field.PricePerHour,
		Date:         fieldSchedule.Date.Format(time.DateOnly),
		Status:       fieldSchedule.Status.GetStatusString(),
		Time:         fmt.Sprintf("%s - %s", fieldSchedule.Time.StartTime, fieldSchedule.Time.EndTime),
		CreatedAt:    fieldSchedule.CreatedAt,
		UpdatedAt:    fieldSchedule.UpdatedAt,
	}
	return &response, nil
}

// Create menambahkan beberapa jadwal sekaligus untuk satu tanggal tertentu.
func (f *FieldScheduleService) Create(ctx context.Context, request *dto.FieldScheduleRequest) error {
	// 1. Pastikan lapangan ada.
	field, err := f.repository.GetField().FindByUUID(ctx, request.FieldID)
	if err != nil {
		return err
	}

	fieldSchedules := make([]models.FieldSchedule, 0, len(request.TimeIDs))
	dateParsed, _ := time.Parse(time.DateOnly, request.Date)
	
	// 2. Iterasi setiap ID jam yang dipilih.
	for _, timeID := range request.TimeIDs {
		// Cari jam yang sesuai.
		scheduleTime, err := f.repository.GetTime().FindByUUID(ctx, timeID)
		if err != nil {
			return err
		}

		// Cek apakah jadwal tersebut sudah ada di database (duplikat).
		schedule, err := f.repository.GetFieldSchedule().FindByDateAndTimeID(ctx, request.Date, int(scheduleTime.ID), int(field.ID))
		if err != nil {
			return err
		}

		if schedule != nil {
			return errFieldSchedule.ErrFieldScheduleIsExist
		}

		// Jika belum ada, tambahkan ke list untuk disimpan.
		fieldSchedules = append(fieldSchedules, models.FieldSchedule{
			UUID:    uuid.New(),
			FieldID: field.ID,
			TimeID:  scheduleTime.ID,
			Date:    dateParsed,
			Status:  constants.Available,
		})
	}

	// 3. Simpan semua jadwal sekaligus (bulk create).
	err = f.repository.GetFieldSchedule().Create(ctx, fieldSchedules)
	if err != nil {
		return err
	}

	return nil
}

// GenerateScheduleForOneMonth mengotomatiskan pembuatan jadwal untuk 30 hari ke depan.
func (f *FieldScheduleService) GenerateScheduleForOneMonth(
	ctx context.Context,
	request *dto.GenerateFieldScheduleForOneMonthRequest,
) error {
	// 1. Cari data lapangan.
	field, err := f.repository.GetField().FindByUUID(ctx, request.FieldID)
	if err != nil {
		return err
	}

	// 2. Ambil semua master jam yang tersedia.
	times, err := f.repository.GetTime().FindAll(ctx)
	if err != nil {
		return err
	}

	// 3. Tentukan jumlah hari (30 hari).
	numberOfDays := 30
	fieldSchedules := make([]models.FieldSchedule, 0, numberOfDays)
	// Mulai dari besok.
	now := time.Now().Add(time.Duration(1) * 24 * time.Hour)
	
	// 4. Loop harian selama 30 hari.
	for i := 0; i < numberOfDays; i++ {
		currentDate := now.AddDate(0, 0, i)
		// 5. Di setiap hari, loop untuk semua jam.
		for _, item := range times {
			// Cek jika sudah ada jadwalnya (untuk menghindari error duplikat).
			schedule, err := f.repository.GetFieldSchedule().FindByDateAndTimeID(
				ctx,
				currentDate.Format(time.DateOnly),
				int(item.ID),
				int(field.ID),
			)
			if err != nil {
				return err
			}

			if schedule != nil {
				// Lewati atau kembalikan error jika sudah ada.
				return errFieldSchedule.ErrFieldScheduleIsExist
			}

			// Tambahkan ke daftar jadwal yang akan dibuat.
			fieldSchedules = append(fieldSchedules, models.FieldSchedule{
				UUID:    uuid.New(),
				FieldID: field.ID,
				TimeID:  item.ID,
				Date:    currentDate,
				Status:  constants.Available,
			})
		}
	}

	// 6. Simpan semua jadwal yang sudah di-generate ke database.
	err = f.repository.GetFieldSchedule().Create(ctx, fieldSchedules)
	if err != nil {
		return err
	}

	return nil
}

// Update digunakan untuk mengubah data jam atau tanggal pada jadwal yang sudah ada.
func (f *FieldScheduleService) Update(
	ctx context.Context,
	uuidParam string,
	request *dto.UpdateFieldScheduleRequest,
) (*dto.FieldScheduleResponse, error) {
	// 1. Pastikan jadwal ada.
	fieldSchedule, err := f.repository.GetFieldSchedule().FindByUUID(ctx, uuidParam)
	if err != nil {
		return nil, err
	}
	
	// 2. Pastikan jam yang baru ada.
	scheduleTime, err := f.repository.GetTime().FindByUUID(ctx, request.TimeID)
	if err != nil {
		return nil, err
	}

	// 3. Validasi: Jangan sampai update ke waktu/tanggal yang sudah ditempati jadwal lain.
	isTimeExist, err := f.repository.GetFieldSchedule().FindByDateAndTimeID(
		ctx,
		request.Date,
		int(scheduleTime.ID),
		int(fieldSchedule.FieldID),
	)
	if err != nil {
		return nil, err
	}

	// Jika jadwal baru sudah ada dan bukan jadwal yang sedang kita edit saat ini.
	if isTimeExist != nil && request.Date != fieldSchedule.Date.Format(time.DateOnly) {
		checkDate, err := f.repository.GetFieldSchedule().FindByDateAndTimeID(
			ctx,
			request.Date,
			int(scheduleTime.ID),
			int(fieldSchedule.FieldID),
		)
		if err != nil {
			return nil, err
		}

		if checkDate != nil {
			return nil, errFieldSchedule.ErrFieldScheduleIsExist
		}
	}

	// 4. Eksekusi update.
	dateParsed, _ := time.Parse(time.DateOnly, request.Date)
	fieldResult, err := f.repository.GetFieldSchedule().Update(ctx, uuidParam, &models.FieldSchedule{
		Date:   dateParsed,
		TimeID: scheduleTime.ID,
	})
	if err != nil {
		return nil, err
	}

	// 5. Mapping ke response.
	response := dto.FieldScheduleResponse{
		UUID:         fieldResult.UUID,
		FieldName:    fieldResult.Field.Name,
		Date:         fieldResult.Date.Format(time.DateOnly),
		PricePerHour: fieldResult.Field.PricePerHour,
		Status:       fieldSchedule.Status.GetStatusString(),
		Time:         fmt.Sprintf("%s - %s", scheduleTime.StartTime, scheduleTime.EndTime),
		CreatedAt:    fieldResult.CreatedAt,
		UpdatedAt:    fieldResult.UpdatedAt,
	}
	return &response, nil
}

// UpdateStatus mengubah beberapa jadwal menjadi 'Booked' (Dipesan).
// Digunakan setelah proses pembayaran berhasil.
func (f *FieldScheduleService) UpdateStatus(
	ctx context.Context,
	request *dto.UpdateStatusFieldScheduleRequest,
) error {
	for _, item := range request.FieldScheduleIDs {
		// Cek keberadaan jadwal.
		_, err := f.repository.GetFieldSchedule().FindByUUID(ctx, item)
		if err != nil {
			return err
		}

		// Update statusnya.
		err = f.repository.GetFieldSchedule().UpdateStatus(ctx, constants.Booked, item)
		if err != nil {
			return err
		}
	}
	return nil
}

// Delete menghapus satu data jadwal.
func (f *FieldScheduleService) Delete(ctx context.Context, uuid string) error {
	// Pastikan ada.
	_, err := f.repository.GetFieldSchedule().FindByUUID(ctx, uuid)
	if err != nil {
		return err
	}

	// Hapus.
	err = f.repository.GetFieldSchedule().Delete(ctx, uuid)
	if err != nil {
		return err
	}

	return nil
}
