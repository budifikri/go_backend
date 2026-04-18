package request

type CreateTelegramRequest struct {
	TelegramIDPenjualan     string `json:"telegram_id_penjualan"`
	TelegramIDPembelian     string `json:"telegram_id_pembelian"`
	TelegramIDStockOpname   string `json:"telegram_id_stock_opname"`
	TelegramIDClosingDrawer string `json:"telegram_id_closing_drawer"`
	NotifyPenjualan         bool   `json:"notify_penjualan"`
	NotifyPembelian         bool   `json:"notify_pembelian"`
	NotifyStockOpname       bool   `json:"notify_stock_opname"`
	NotifyClosingDrawer     bool   `json:"notify_closing_drawer"`
	IsActive                bool   `json:"is_active"`
}

type TestTelegramRequest struct {
	TelegramID string `json:"telegram_id" validate:"required"`
	Type       string `json:"type" validate:"required"`
}
