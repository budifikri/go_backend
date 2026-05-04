package request

type CreatePaketRequest struct {
	KodePaket string             `json:"kodepaket" validate:"required,min=1,max=50"`
	NmPaket   string             `json:"nm_paket" validate:"required,min=1,max=150"`
	Deskripsi string             `json:"deskripsi"`
	IsActive  *bool              `json:"is_active"`
	Items     []PaketItemRequest `json:"items" validate:"required,min=1"`
}

type PaketItemRequest struct {
	IDProduk string `json:"id_produk" validate:"required"`
}

type UpdatePaketRequest struct {
	KodePaket *string            `json:"kodepaket"`
	NmPaket   *string            `json:"nm_paket"`
	Deskripsi *string            `json:"deskripsi"`
	IsActive  *bool              `json:"is_active"`
	Items     []PaketItemRequest `json:"items"`
}
