package request

// CreateJadwalDokterRequest represents the request body for creating jadwal dokter
type CreateJadwalDokterRequest struct {
	DokterID   string `json:"dokter_id" validate:"required,uuid4"`
	Hari       string `json:"hari" validate:"required,oneof=Senin Selasa Rabu Kamis Jumat Sabtu Minggu"`
	JamMulai   string `json:"jam_mulai" validate:"required,datetime=15:04"`
	JamSelesai string `json:"jam_selesai" validate:"required,datetime=15:04"`
}

// UpdateJadwalDokterRequest represents the request body for updating jadwal dokter
type UpdateJadwalDokterRequest struct {
	DokterID   *string `json:"dokter_id" validate:"omitempty,uuid4"`
	Hari       *string `json:"hari" validate:"omitempty,oneof=Senin Selasa Rabu Kamis Jumat Sabtu Minggu"`
	JamMulai   *string `json:"jam_mulai" validate:"omitempty,datetime=15:04"`
	JamSelesai *string `json:"jam_selesai" validate:"omitempty,datetime=15:04"`
	IsActive   *bool   `json:"is_active"`
}
