package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/pos-retail/go_backend/internal/models"
	"github.com/pos-retail/go_backend/internal/repository"
	"github.com/pos-retail/go_backend/internal/types/request"
	"github.com/pos-retail/go_backend/internal/types/response"
	"gorm.io/gorm"
)

func getApiKey() string {
	return os.Getenv("API_KEY_TELEGRAM")
}

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
		CompanyID:               companyID,
		TelegramIDPenjualan:     input.TelegramIDPenjualan,
		TelegramIDPembelian:     input.TelegramIDPembelian,
		TelegramIDStockOpname:   input.TelegramIDStockOpname,
		TelegramIDClosingDrawer: input.TelegramIDClosingDrawer,
		NotifyPenjualan:         input.NotifyPenjualan,
		NotifyPembelian:         input.NotifyPembelian,
		NotifyStockOpname:       input.NotifyStockOpname,
		NotifyClosingDrawer:     input.NotifyClosingDrawer,
		IsActive:                input.IsActive,
	}

	saved, err := s.telegramRepo.SaveConfig(config)
	if err != nil {
		return response.NewErrorResponse(err.Error())
	}

	return response.NewSuccessResponse(saved, "")
}

func (s *TelegramService) TestConnection(input request.TestTelegramRequest) response.ApiResponse {
	apiKey := getApiKey()
	if apiKey == "" {
		return response.NewErrorResponse("API Key tidak dikonfigurasi di server environment")
	}

	message := s.formatTestMessage(input.Type)
	err := s.sendMessage(input.TelegramID, apiKey, message)
	if err != nil {
		return response.NewErrorResponse(fmt.Sprintf("Koneksi gagal: %s", err.Error()))
	}

	return response.NewSuccessResponse(map[string]string{"status": "success", "message": "Koneksi berhasil!"}, "")
}

func (s *TelegramService) formatTestMessage(notifType string) string {
	switch notifType {
	case "penjualan":
		return "*PREVIEW: PENJUALAN BARU*\n\n" +
			"🕒 Waktu: 18 Apr 2026, 10:30\n" +
			"💰 Total: Rp 150.000\n" +
			"👤 Kasir: Admin\n" +
			"🏷️ No: SL/2026/0001\n\n" +
			"📦 *Detail Produk:*\n" +
			"━━━━━━━━━━━━━━━━━━━━━━━\n" +
			"• Produk A × 2 = Rp 100.000\n" +
			"• Produk B × 1 = Rp 50.000\n" +
			"━━━━━━━━━━━━━━━━━━━━━━━\n" +
			"📊 Total Item: 2 | Qty: 3"
	case "pembelian":
		return "*PREVIEW: PEMBELIAN BARU*\n\n" +
			"🏢 Supplier: Supplier ABC\n" +
			"📅 Tanggal: 18 Apr 2026\n" +
			"💰 Total: Rp 500.000\n" +
			"📄 No: PO/2026/0001\n\n" +
			"📦 *Detail Produk:*\n" +
			"━━━━━━━━━━━━━━━━━━━━━━━\n" +
			"• Produk A × 10 = Rp 300.000\n" +
			"• Produk B × 5 = Rp 200.000\n" +
			"━━━━━━━━━━━━━━━━━━━━━━━\n" +
			"📊 Total Item: 2 | Qty: 15"
	case "stock_opname":
		return "*PREVIEW: STOCK OPNAME SELESAI*\n\n" +
			"🏢 Warehouse: Gudang Utama\n" +
			"📅 Tanggal: 18 Apr 2026\n" +
			"✅ Status: Completed\n\n" +
			"📊 *Ringkasan:*\n" +
			"━━━━━━━━━━━━━━━━━━━━━━━\n" +
			"📋 Total SKU: 100\n" +
			"✅ Sesuai: 95\n" +
			"⚠️ Selisih: 5 item"
	case "closing_drawer":
		return "*PREVIEW: CLOSING DRAWER*\n\n" +
			"🕒 Waktu: 18 Apr 2026, 22:00\n" +
			"👤 Kasir: Admin\n" +
			"🏷️ No: CD/2026/0001\n\n" +
			"💰 *Setoran:*\n" +
			"━━━━━━━━━━━━━━━━━━━━━━━\n" +
			"💵 Saldo Penutupan: Rp 1.000.000\n" +
			"📊 Ekspektasi: Rp 980.000\n" +
			"⚖️ Selisih: Rp 20.000\n" +
			"━━━━━━━━━━━━━━━━━━━━━━━\n" +
			"\n✅ Penutupan laci telah dilakukan!"
	default:
		return "✅ Koneksi berhasil! Bot Telegram terhubung dengan benar."
	}
}

func (s *TelegramService) SendNotification(telegramID, message string) error {
	apiKey := getApiKey()
	if apiKey == "" {
		return fmt.Errorf("API Key tidak dikonfigurasi")
	}
	log.Printf("[TELEGRAM] Sending to ID: %s, Message length: %d", telegramID, len(message))
	err := s.sendMessage(telegramID, apiKey, message)
	if err != nil {
		log.Printf("[TELEGRAM] SendMessage error: %v", err)
		return err
	}
	log.Printf("[TELEGRAM] SendMessage success")
	return nil
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

func (s *TelegramService) FormatPenjualanMessageRow(sale *repository.SaleWithNames, items []repository.SaleItemWithProfit) string {
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

func (s *TelegramService) FormatPembelianReceiveMessage(po *repository.PurchaseOrderRow, items []repository.PurchaseOrderItemRow, totalReceive int) string {
	var buffer bytes.Buffer

	supplierName := "Unknown"
	if po.SupplierName != nil {
		supplierName = *po.SupplierName
	}

	buffer.WriteString("*📦 BARANG DITERIMA*\n\n")
	buffer.WriteString(fmt.Sprintf("🏢 Supplier: %s\n", supplierName))
	buffer.WriteString(fmt.Sprintf("📅 Tanggal: %s\n", po.CreatedAt))
	buffer.WriteString(fmt.Sprintf("🏷️ No PO: %s\n", po.PoNumber))

	if len(items) > 0 {
		buffer.WriteString("\n📦 *Detail Diterima:*\n")
		buffer.WriteString("━━━━━━━━━━━━━━━━━━━━━━━\n")

		for _, item := range items {
			if item.QtyReceive > 0 {
				productName := "Unknown"
				if item.ProductName != nil {
					productName = *item.ProductName
				}
				buffer.WriteString(fmt.Sprintf("• %s: %d pcs\n", productName, item.QtyReceive))
			}
		}

		buffer.WriteString("━━━━━━━━━━━━━━━━━━━━━━━\n")
		buffer.WriteString(fmt.Sprintf("📊 Total Diterima: %d pcs\n", totalReceive))
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

func (s *TelegramService) FormatClosingDrawerMessage(drawer *models.CashDrawer, closingBy string) string {
	var buffer bytes.Buffer

	buffer.WriteString("*CLOSING DRAWER*\n\n")
	buffer.WriteString(fmt.Sprintf("🕒 Waktu: %s\n", drawer.ClosedAt.Format("02 Jan 2006, 15:04")))
	buffer.WriteString(fmt.Sprintf("👤 Kasir: %s\n", closingBy))
	buffer.WriteString(fmt.Sprintf("🏷️ No: %s\n", drawer.DrawerNumber))

	buffer.WriteString("\n💰 *Setoran:*\n")
	buffer.WriteString("━━━━━━━━━━━━━━━━━━━━━━━\n")
	closingBalance := 0.0
	if drawer.ClosingBalance != nil {
		closingBalance = *drawer.ClosingBalance
	}
	buffer.WriteString(fmt.Sprintf("💵 Saldo Penutupan: Rp %.0f\n", closingBalance))
	buffer.WriteString(fmt.Sprintf("📊 Ekspektasi: Rp %.0f\n", drawer.ExpectedBalance))
	if drawer.Variance != nil {
		buffer.WriteString(fmt.Sprintf("⚖️ Selisih: Rp %.0f\n", *drawer.Variance))
	}
	buffer.WriteString("━━━━━━━━━━━━━━━━━━━━━━━\n")
	buffer.WriteString("\n✅ Penutupan laci telah dilakukan!")

	return buffer.String()
}
