package request

import "github.com/pos-retail/go_backend/internal/models"

type CreateDokterRequest struct {
	Nama         string              `json:"nama" validate:"required"`
	JenisKelamin models.JenisKelamin `json:"jenis_kelamin" validate:"required,oneof=L P"`
	TempatLahir  string              `json:"tempat_lahir" validate:"required"`
	TanggalLahir string              `json:"tanggal_lahir" validate:"required"`
	Alamat       string              `json:"alamat" validate:"required"`
	NoTelp       string              `json:"no_telp" validate:"required"`
	Email        string              `json:"email" validate:"required,email"`
	Tipe         models.TipeDokter   `json:"tipe" validate:"required,oneof=Dokter Beautician"`
	Active       *bool               `json:"active" validate:"omitempty"`
}

type UpdateDokterRequest struct {
	Nama         *string              `json:"nama" validate:"omitempty"`
	JenisKelamin *models.JenisKelamin `json:"jenis_kelamin" validate:"omitempty,oneof=L P"`
	TempatLahir  *string              `json:"tempat_lahir" validate:"omitempty"`
	TanggalLahir *string              `json:"tanggal_lahir" validate:"omitempty"`
	Alamat       *string              `json:"alamat" validate:"omitempty"`
	NoTelp       *string              `json:"no_telp" validate:"omitempty"`
	Email        *string              `json:"email" validate:"omitempty,email"`
	Tipe         *models.TipeDokter   `json:"tipe" validate:"omitempty,oneof=Dokter Beautician"`
	Active       *bool                `json:"active" validate:"omitempty"`
}
