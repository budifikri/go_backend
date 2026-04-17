package request

type CreateTelegramRequest struct {
	APIKey                string `json:"api_key" validate:"required"`
	TelegramIDPenjualan   string `json:"telegram_id_penjualan"`
	TelegramIDPembelian   string `json:"telegram_id_pembelian"`
	TelegramIDStockOpname string `json:"telegram_id_stock_opname"`
	NotifyPenjualan       bool   `json:"notify_penjualan"`
	NotifyPembelian       bool   `json:"notify_pembelian"`
	NotifyStockOpname     bool   `json:"notify_stock_opname"`
	IsActive              bool   `json:"is_active"`
}

type TestTelegramRequest struct {
	TelegramID string `json:"telegram_id" validate:"required"`
	APIKey     string `json:"api_key" validate:"required"`
}
