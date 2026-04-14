package services

import (
	"bytes"
	"context"
	"field-service/common/gcs"
	"field-service/common/util"
	errConstant "field-service/constants/error"
	"field-service/domain/dto"
	"field-service/domain/models"
	"field-service/repositories"
	"fmt"
	"io"
	"mime/multipart"
	"path"
	"time"

	"github.com/google/uuid"
)

// FieldService adalah struct yang mengimplementasikan logika bisnis untuk fitur Lapangan (Field).
// Struct ini membutuhkan repository untuk akses database dan gcs client untuk upload gambar.
type FieldService struct {
	repository repositories.IRepositoryRegistry
	gcs        gcs.IGCSClient
}

// IFieldService adalah interface yang mendefinisikan kontrak (metode apa saja yang harus ada)
// untuk layanan Lapangan. Ini memudahkan dalam testing (mocking).
type IFieldService interface {
	GetAllWithPagination(context.Context, *dto.FieldRequestParam) (*util.PaginationResult, error)
	GetAllWithoutPagination(context.Context) ([]dto.FieldResponse, error)
	GetByUUID(context.Context, string) (*dto.FieldResponse, error)
	Create(context.Context, *dto.FieldRequest) (*dto.FieldResponse, error)
	Update(context.Context, string, *dto.UpdateFieldRequest) (*dto.FieldResponse, error)
	Delete(context.Context, string) error
}

// NewFieldService adalah function constructor untuk membuat instance baru dari FieldService.
// Mengembalikan interface IFieldService.
func NewFieldService(repository repositories.IRepositoryRegistry, gcs gcs.IGCSClient) IFieldService {
	return &FieldService{repository: repository, gcs: gcs}
}

// GetAllWithPagination mengambil semua data lapangan dengan sistem pagination (halaman).
func (f *FieldService) GetAllWithPagination(
	ctx context.Context,
	param *dto.FieldRequestParam,
) (*util.PaginationResult, error) {
	// 1. Ambil data dari repository dengan parameter pagination yang dikirim user.
	fields, total, err := f.repository.GetField().FindAllWithPagination(ctx, param)
	if err != nil {
		return nil, err
	}

	// 2. Mapping dari model database ke DTO (Data Transfer Object) untuk response.
	fieldResults := make([]dto.FieldResponse, 0, len(fields))
	for _, field := range fields {
		fieldResults = append(fieldResults, dto.FieldResponse{
			UUID:         field.UUID,
			Code:         field.Code,
			Name:         field.Name,
			PricePerHour: field.PricePerHour,
			Images:       field.Images,
			CreatedAt:    field.CreatedAt,
			UpdatedAt:    field.UpdatedAt,
		})
	}

	// 3. Bungkus hasil ke dalam struktur pagination agar user tahu total data dan halaman saat ini.
	pagination := &util.PaginationParam{
		Count: total,
		Page:  param.Page,
		Limit: param.Limit,
		Data:  fieldResults,
	}

	response := util.GeneratePagination(*pagination)
	return &response, nil
}

// GetAllWithoutPagination mengambil semua data lapangan tanpa pembatasan halaman.
func (f *FieldService) GetAllWithoutPagination(ctx context.Context) ([]dto.FieldResponse, error) {
	// 1. Ambil semua data langsung dari repository.
	fields, err := f.repository.GetField().FindAllWithoutPagination(ctx)
	if err != nil {
		return nil, err
	}

	// 2. Mapping ke DTO response.
	fieldResults := make([]dto.FieldResponse, 0, len(fields))
	for _, field := range fields {
		fieldResults = append(fieldResults, dto.FieldResponse{
			UUID:         field.UUID,
			Name:         field.Name,
			PricePerHour: field.PricePerHour,
			Images:       field.Images,
		})
	}

	return fieldResults, nil
}

// GetByUUID mengambil satu data lapangan berdasarkan UUID yang unik.
func (f *FieldService) GetByUUID(ctx context.Context, uuid string) (*dto.FieldResponse, error) {
	// 1. Cari data di repository berdasarkan UUID.
	field, err := f.repository.GetField().FindByUUID(ctx, uuid)
	if err != nil {
		return nil, err
	}

	// 2. Format harga ke Rupiah sebelum dikembalikan ke user.
	pricePerHour := float64(field.PricePerHour)
	fieldResult := dto.FieldResponse{
		UUID:         field.UUID,
		Code:         field.Code,
		Name:         field.Name,
		PricePerHour: util.RupiahFormat(&pricePerHour),
		Images:       field.Images,
		CreatedAt:    field.CreatedAt,
		UpdatedAt:    field.UpdatedAt,
	}

	return &fieldResult, nil
}

// validateUpload adalah fungsi internal (helper) untuk memvalidasi file gambar yang diupload.
func (f *FieldService) validateUpload(images []multipart.FileHeader) error {
	// Pastikan ada file yang dikirim.
	if images == nil || len(images) == 0 {
		return errConstant.ErrInvalidUploadFile
	}

	// Cek ukuran file, maksimal 5MB.
	for _, image := range images {
		if image.Size > 5*1024*1024 {
			return errConstant.ErrSizeTooBig
		}
	}
	return nil
}

// processAndUploadImage adalah fungsi internal untuk memproses satu file dan menguploadnya ke GCS.
func (f *FieldService) processAndUploadImage(ctx context.Context, image multipart.FileHeader) (string, error) {
	// 1. Buka file.
	file, err := image.Open()
	if err != nil {
		return "", err
	}
	defer file.Close()

	// 2. Baca isi file ke buffer.
	buffer := new(bytes.Buffer)
	_, err = io.Copy(buffer, file)
	if err != nil {
		return "", err
	}

	// 3. Buat nama file yang unik berdasarkan waktu agar tidak duplikat.
	filename := fmt.Sprintf("images/%s-%s-%s", time.Now().Format("20060102150405"), image.Filename, path.Ext(image.Filename))
	
	// 4. Upload ke Google Cloud Storage.
	url, err := f.gcs.UploadFile(ctx, filename, buffer.Bytes())
	if err != nil {
		return "", err
	}
	return url, nil
}

// uploadImage adalah fungsi internal untuk mengelola proses upload beberapa gambar sekaligus.
func (f *FieldService) uploadImage(ctx context.Context, images []multipart.FileHeader) ([]string, error) {
	// 1. Validasi dulu semua gambar.
	err := f.validateUpload(images)
	if err != nil {
		return nil, err
	}

	// 2. Loop dan upload satu per satu.
	urls := make([]string, 0, len(images))
	for _, image := range images {
		url, err := f.processAndUploadImage(ctx, image)
		if err != nil {
			return nil, err
		}
		urls = append(urls, url)
	}
	return urls, nil
}

// Create digunakan untuk menambahkan data lapangan baru beserta gambar-gambarnya.
func (f *FieldService) Create(ctx context.Context, request *dto.FieldRequest) (*dto.FieldResponse, error) {
	// 1. Upload gambar terlebih dahulu.
	imageUrl, err := f.uploadImage(ctx, request.Images)
	if err != nil {
		return nil, err
	}

	// 2. Simpan data lapangan ke database.
	field, err := f.repository.GetField().Create(ctx, &models.Field{
		Code:         request.Code,
		Name:         request.Name,
		PricePerHour: request.PricePerHour,
		Images:       imageUrl,
	})
	if err != nil {
		return nil, err
	}

	// 3. Mapping hasil simpan ke response DTO.
	response := dto.FieldResponse{
		UUID:         field.UUID,
		Code:         field.Code,
		Name:         field.Name,
		PricePerHour: field.PricePerHour,
		Images:       field.Images,
		CreatedAt:    field.CreatedAt,
		UpdatedAt:    field.UpdatedAt,
	}
	return &response, nil
}

// Update digunakan untuk memperbarui data lapangan yang sudah ada.
func (f *FieldService) Update(ctx context.Context, uuidParam string, request *dto.UpdateFieldRequest) (*dto.FieldResponse, error) {
	// 1. Pastikan data yang mau diupdate memang ada.
	field, err := f.repository.GetField().FindByUUID(ctx, uuidParam)
	if err != nil {
		return nil, err
	}

	// 2. Cek apakah user mengupload gambar baru atau tetap pakai yang lama.
	var imageUrls []string
	if request.Images == nil {
		imageUrls = field.Images
	} else {
		imageUrls, err = f.uploadImage(ctx, request.Images)
		if err != nil {
			return nil, err
		}
	}

	// 3. Lakukan update di repository.
	fieldResult, err := f.repository.GetField().Update(ctx, uuidParam, &models.Field{
		Code:         request.Code,
		Name:         request.Name,
		PricePerHour: request.PricePerHour,
		Images:       imageUrls,
	})

	// 4. Mapping ke response.
	uuidParsed, _ := uuid.Parse(uuidParam)
	response := dto.FieldResponse{
		UUID:         uuidParsed,
		Code:         fieldResult.Code,
		Name:         fieldResult.Name,
		PricePerHour: fieldResult.PricePerHour,
		Images:       fieldResult.Images,
		CreatedAt:    fieldResult.CreatedAt,
		UpdatedAt:    fieldResult.UpdatedAt,
	}
	return &response, nil
}

// Delete digunakan untuk menghapus data lapangan berdasarkan UUID.
func (f *FieldService) Delete(ctx context.Context, uuid string) error {
	// 1. Pastikan data ada sebelum dihapus.
	_, err := f.repository.GetField().FindByUUID(ctx, uuid)
	if err != nil {
		return err
	}

	// 2. Hapus data dari database.
	err = f.repository.GetField().Delete(ctx, uuid)
	if err != nil {
		return err
	}
	return nil
}
