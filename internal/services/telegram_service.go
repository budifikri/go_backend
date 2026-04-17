package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/pos-retail/go_backend/internal/models"
	"github.com/pos-retail/go_backend/internal/repository"
	"github.com/pos-retail/go_backend/internal/types/request"
	"github.com/pos-retail/go_backend/internal/types/response"
	"gorm.io/gorm"
)

type TelegramService struct {
	db           *gorm.DB
	telegramRepo *repository.TelegramRepository
}

func NewTelegramService(db *gorm.DB, telegramRepo *repository.TelegramRepository) *TelegramService {
	return &TelegramService{db: db, telegramRepo: telegramRepo}
}

func (s *TelegramService) GetConfigByCompany(companyID uuid.UUID) (*models.TelegramConfig, error) {
	return s.telegramRepo.GetConfigByCompany(companyID)
}

func (s *TelegramService) SaveConfig(companyID uuid.UUID, input request.CreateTelegramRequest) response.ApiResponse {
	config := &models.TelegramConfig{
		CompanyID:             companyID,
		APIKey:                input.APIKey,
		TelegramIDPenjualan:   input.TelegramIDPenjualan,
		TelegramIDPembelian:   input.TelegramIDPembelian,
		TelegramIDStockOpname: input.TelegramIDStockOpname,
		NotifyPenjualan:       input.NotifyPenjualan,
		NotifyPembelian:       input.NotifyPembelian,
		NotifyStockOpname:     input.NotifyStockOpname,
		IsActive:              input.IsActive,
	}

	saved, err := s.telegramRepo.SaveConfig(config)
	if err != nil {
		return response.NewErrorResponse(err.Error())
	}

	return response.NewSuccessResponse(saved, "")
}

func (s *TelegramService) TestConnection(input request.TestTelegramRequest) response.ApiResponse {
	err := s.sendMessage(input.TelegramID, input.APIKey, "✅ Koneksi berhasil! Bot Telegram terhubung dengan benar.")
	if err != nil {
		return response.NewErrorResponse(fmt.Sprintf("Koneksi gagal: %s", err.Error()))
	}

	return response.NewSuccessResponse(map[string]string{"status": "success", "message": "Koneksi berhasil!"}, "")
}

func (s *TelegramService) SendNotification(telegramID, apiKey, message string) error {
	return s.sendMessage(telegramID, apiKey, message)
}

func (s *TelegramService) sendMessage(telegramID, apiKey, message string) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", apiKey)

	payload := map[string]interface{}{
		"chat_id":    telegramID,
		"text":       message,
		"parse_mode": "Markdown",
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("telegram API returned status %d", resp.StatusCode)
	}

	return nil
}

func (s *TelegramService) FormatPenjualanMessageRow(sale *repository.SaleWithNames, items []repository.SaleItemWithProduct) string {
	var buffer bytes.Buffer

	buffer.WriteString("*PENJUALAN BARU*\n\n")
	buffer.WriteString(fmt.Sprintf("🕒 Waktu: %s\n", sale.CreatedAt))
	buffer.WriteString(fmt.Sprintf("💰 Total: Rp %.0f\n", sale.TotalAmount))
	buffer.WriteString(fmt.Sprintf("👤 Kasir: %s\n", sale.CashierName))
	buffer.WriteString(fmt.Sprintf("🏷️ No: %s\n", sale.SaleNumber))

	if len(items) > 0 {
		buffer.WriteString("\n📦 *Detail Produk:*\n")
		buffer.WriteString("━━━━━━━━━━━━━━━━━━━━━━━\n")

		maxItems := 5
		if len(items) > maxItems {
			for i := 0; i < maxItems; i++ {
				lineTotal := items[i].UnitPrice * float64(items[i].Quantity)
				buffer.WriteString(fmt.Sprintf("• %s × %d = Rp %.0f\n", items[i].ProductName, items[i].Quantity, lineTotal))
			}
			buffer.WriteString(fmt.Sprintf("\n...dan %d item lainnya\n", len(items)-maxItems))
		} else {
			for _, item := range items {
				lineTotal := item.UnitPrice * float64(item.Quantity)
				buffer.WriteString(fmt.Sprintf("• %s × %d = Rp %.0f\n", item.ProductName, item.Quantity, lineTotal))
			}
		}

		buffer.WriteString("━━━━━━━━━━━━━━━━━━━━━━━\n")
		totalQty := 0
		for _, item := range items {
			totalQty += item.Quantity
		}
		buffer.WriteString(fmt.Sprintf("📊 Total Item: %d | Qty: %d\n", len(items), totalQty))
	}

	return buffer.String()
}

func (s *TelegramService) FormatPembelianMessageRow(po *repository.PurchaseOrderRow, items []repository.PurchaseOrderItemRow) string {
	var buffer bytes.Buffer

	supplierName := "Unknown"
	if po.SupplierName != nil {
		supplierName = *po.SupplierName
	}

	totalAmount, _ := parseAmount(po.TotalAmount)

	buffer.WriteString("*PEMBELIAN BARU*\n\n")
	buffer.WriteString(fmt.Sprintf("🏢 Supplier: %s\n", supplierName))
	buffer.WriteString(fmt.Sprintf("📅 Tanggal: %s\n", po.CreatedAt))
	buffer.WriteString(fmt.Sprintf("💰 Total: Rp %.0f\n", totalAmount))
	buffer.WriteString(fmt.Sprintf("📄 No: %s\n", po.PoNumber))

	if len(items) > 0 {
		buffer.WriteString("\n📦 *Detail Produk:*\n")
		buffer.WriteString("━━━━━━━━━━━━━━━━━━━━━━━\n")

		maxItems := 5
		if len(items) > maxItems {
			for i := 0; i < maxItems; i++ {
				lineTotal, _ := parseAmount(items[i].LineTotal)
				productName := "Unknown"
				if items[i].ProductName != nil {
					productName = *items[i].ProductName
				}
				buffer.WriteString(fmt.Sprintf("• %s × %d = Rp %.0f\n", productName, items[i].QtyPo, lineTotal))
			}
			buffer.WriteString(fmt.Sprintf("\n...dan %d item lainnya\n", len(items)-maxItems))
		} else {
			for _, item := range items {
				lineTotal, _ := parseAmount(item.LineTotal)
				productName := "Unknown"
				if item.ProductName != nil {
					productName = *item.ProductName
				}
				buffer.WriteString(fmt.Sprintf("• %s × %d = Rp %.0f\n", productName, item.QtyPo, lineTotal))
			}
		}

		buffer.WriteString("━━━━━━━━━━━━━━━━━━━━━━━\n")
		totalQty := 0
		for _, item := range items {
			totalQty += item.QtyPo
		}
		buffer.WriteString(fmt.Sprintf("📊 Total Item: %d | Qty: %d\n", len(items), totalQty))
	}

	return buffer.String()
}

func (s *TelegramService) FormatStockOpnameMessage(opname *models.StockOpname, warehouseName string, discrepancies int) string {
	var buffer bytes.Buffer

	buffer.WriteString("*STOCK OPNAME SELESAI*\n\n")
	buffer.WriteString(fmt.Sprintf("🏢 Warehouse: %s\n", warehouseName))
	buffer.WriteString(fmt.Sprintf("📅 Tanggal: %s\n", opname.CreatedAt.Format("02 Jan 2006")))
	buffer.WriteString(fmt.Sprintf("✅ Status: %s\n", opname.Status))

	buffer.WriteString("\n📊 *Ringkasan:*\n")
	buffer.WriteString("━━━━━━━━━━━━━━━━━━━━━━━\n")
	buffer.WriteString(fmt.Sprintf("📋 Total SKU: -\n"))
	buffer.WriteString(fmt.Sprintf("✅ Sesuai: -\n"))
	buffer.WriteString(fmt.Sprintf("⚠️ Selisih: %d item\n", discrepancies))
	buffer.WriteString("━━━━━━━━━━━━━━━━━━━━━━━\n")

	return buffer.String()
}

func parseAmount(s string) (float64, error) {
	if s == "" {
		return 0, nil
	}
	return strconv.ParseFloat(s, 64)
}
