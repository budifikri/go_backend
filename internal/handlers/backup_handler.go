package handlers

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/pos-retail/go_backend/internal/models"
	"github.com/pos-retail/go_backend/internal/services"
	"github.com/pos-retail/go_backend/internal/utils"
)

type BackupHandler struct {
	backupService *services.BackupService
}

func NewBackupHandler(backupService *services.BackupService) *BackupHandler {
	return &BackupHandler{
		backupService: backupService,
	}
}

type UserPayload struct {
	UserID    string
	CompanyID uuid.UUID
	Role      string
}

func GetUserFromContext(c *fiber.Ctx) *UserPayload {
	user, ok := c.Locals("user").(*utils.JWTPayload)
	if !ok {
		return nil
	}

	companyID, err := uuid.Parse(user.CompanyID)
	if err != nil {
		return nil
	}

	return &UserPayload{
		UserID:    user.UserID,
		CompanyID: companyID,
		Role:      user.Role,
	}
}

func (h *BackupHandler) CreateBackup(c *fiber.Ctx) error {
	user := GetUserFromContext(c)
	if user == nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "Unauthorized",
		})
	}

	if user.Role != "admin" && user.Role != "superadmin" {
		return c.Status(http.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"error":   "Only admin can create backup",
		})
	}

	var req models.CreateBackupRequest
	if err := c.BodyParser(&req); err != nil {
		req.IsAuto = false
	}

	backup, err := h.backupService.CreateBackup(user.CompanyID, "", user.UserID, req.IsAuto)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"id":                  backup.ID,
			"filename":            backup.Filename,
			"file_size":           backup.FileSize,
			"file_size_formatted": formatFileSize(backup.FileSize),
			"status":              backup.Status,
			"table_count":         backup.TableCount,
			"row_count":           backup.RowCount,
			"created_at":          backup.CreatedAt,
			"is_auto":             backup.IsAuto,
		},
	})
}

func (h *BackupHandler) ListBackups(c *fiber.Ctx) error {
	user := GetUserFromContext(c)
	if user == nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "Unauthorized",
		})
	}

	backups, err := h.backupService.ListBackups(user.CompanyID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	var response []fiber.Map
	for _, backup := range backups {
		response = append(response, fiber.Map{
			"id":                  backup.ID,
			"company_id":          backup.CompanyID,
			"filename":            backup.Filename,
			"file_path":           backup.FilePath,
			"file_size":           backup.FileSize,
			"file_size_formatted": formatFileSize(backup.FileSize),
			"status":              backup.Status,
			"created_by":          backup.CreatedBy,
			"created_at":          backup.CreatedAt,
			"is_auto":             backup.IsAuto,
			"table_count":         backup.TableCount,
			"row_count":           backup.RowCount,
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    response,
	})
}

func (h *BackupHandler) DownloadBackup(c *fiber.Ctx) error {
	user := GetUserFromContext(c)
	if user == nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "Unauthorized",
		})
	}

	filename := c.Params("filename")
	if filename == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Filename required",
		})
	}

	filePath, err := h.backupService.GetBackupFilePath(user.CompanyID, filename)
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   "Backup file not found",
		})
	}

	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	return c.SendFile(filePath)
}

func (h *BackupHandler) DeleteBackup(c *fiber.Ctx) error {
	user := GetUserFromContext(c)
	if user == nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "Unauthorized",
		})
	}

	if user.Role != "admin" && user.Role != "superadmin" {
		return c.Status(http.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"error":   "Only admin can delete backup",
		})
	}

	filename := c.Params("filename")
	if filename == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Filename required",
		})
	}

	err := h.backupService.DeleteBackup(user.CompanyID, filename)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Backup deleted successfully",
	})
}

func (h *BackupHandler) GetSchedule(c *fiber.Ctx) error {
	user := GetUserFromContext(c)
	if user == nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "Unauthorized",
		})
	}

	schedule, err := h.backupService.GetSchedule(user.CompanyID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    schedule,
	})
}

func (h *BackupHandler) UpdateSchedule(c *fiber.Ctx) error {
	user := GetUserFromContext(c)
	if user == nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "Unauthorized",
		})
	}

	if user.Role != "admin" && user.Role != "superadmin" {
		return c.Status(http.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"error":   "Only admin can update schedule",
		})
	}

	var req models.UpdateScheduleRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body",
		})
	}

	err := h.backupService.UpdateSchedule(user.CompanyID, &req)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	schedule, _ := h.backupService.GetSchedule(user.CompanyID)

	return c.JSON(fiber.Map{
		"success": true,
		"data":    schedule,
		"message": "Schedule updated successfully",
	})
}

func (h *BackupHandler) ValidateRestore(c *fiber.Ctx) error {
	user := GetUserFromContext(c)
	if user == nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "Unauthorized",
		})
	}

	filename := c.Query("filename")
	if filename == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Filename required",
		})
	}

	validation, err := h.backupService.ValidateRestore(user.CompanyID, filename)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    validation,
	})
}

func (h *BackupHandler) RestoreBackup(c *fiber.Ctx) error {
	user := GetUserFromContext(c)
	if user == nil {
		fmt.Println("[DEBUG] RestoreBackup: User is nil!")
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "Unauthorized",
		})
	}

	fmt.Printf("[DEBUG] RestoreBackup: User authenticated - companyID: %s, role: %s\n", user.CompanyID, user.Role)

	if user.Role != "admin" && user.Role != "superadmin" {
		fmt.Println("[DEBUG] RestoreBackup: User is not admin!")
		return c.Status(http.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"error":   "Only admin can restore backup",
		})
	}

	var req models.RestoreRequest
	if err := c.BodyParser(&req); err != nil {
		fmt.Printf("[DEBUG] RestoreBackup: BodyParser error - %v\n", err)
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body",
		})
	}

	fmt.Printf("[DEBUG] RestoreBackup: Request - filename: %s, confirm: %v\n", req.Filename, req.Confirm)

	if req.Filename == "" {
		fmt.Println("[DEBUG] RestoreBackup: Filename is empty!")
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Filename required",
		})
	}

	if !req.Confirm {
		fmt.Println("[DEBUG] RestoreBackup: Confirm is false!")
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Confirmation required",
		})
	}

	fmt.Println("[DEBUG] RestoreBackup: Calling backupService.RestoreBackup...")
	result, err := h.backupService.RestoreBackup(user.CompanyID, "", req.Filename, req.Confirm)
	if err != nil {
		fmt.Printf("[DEBUG] RestoreBackup: Service returned error - %v\n", err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	fmt.Printf("[DEBUG] RestoreBackup: Success - result: %+v\n", result)
	return c.JSON(fiber.Map{
		"success": true,
		"data":    result,
		"message": "Restore completed successfully",
	})
}

func (h *BackupHandler) RestoreProgress(c *fiber.Ctx) error {
	user := GetUserFromContext(c)
	if user == nil {
		fmt.Println("[DEBUG] RestoreProgress: User is nil!")
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "Unauthorized",
		})
	}

	fmt.Printf("[DEBUG] RestoreProgress: Connected for companyID: %s\n", user.CompanyID)

	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("X-Accel-Buffering", "no")
	c.Status(200)

	ch, done := h.backupService.SubscribeProgress(user.CompanyID)

	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		fmt.Println("[DEBUG] RestoreProgress: Stream writer started")
		for {
			select {
			case progress, ok := <-ch:
				if !ok {
					fmt.Println("[DEBUG] RestoreProgress: Channel closed")
					return
				}
				fmt.Printf("[DEBUG] RestoreProgress: Sending - Stage: %s, Progress: %.2f\n", progress.Stage, progress.Progress)
				if strings.HasPrefix(progress.Message, "[") {
					fmt.Fprintf(w, "event: complete\n")
					fmt.Fprintf(w, "data: %s\n\n", progress.Message)
				} else {
					data, _ := json.Marshal(progress)
					fmt.Fprintf(w, "data: %s\n\n", data)
				}
				w.Flush()
			case <-c.Context().Done():
				fmt.Println("[DEBUG] RestoreProgress: Context done")
				done()
				return
			}
		}
	})

	<-c.Context().Done()
	done()
	return nil
}

func (h *BackupHandler) DeleteData(c *fiber.Ctx) error {
	user := GetUserFromContext(c)
	if user == nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "Unauthorized",
		})
	}

	if user.Role != "admin" && user.Role != "superadmin" {
		return c.Status(http.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"error":   "Only admin can delete data",
		})
	}

	var req models.DeleteDataRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body",
		})
	}

	if req.Scope == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Scope is required (all, master, transaction)",
		})
	}

	if req.Scope != "all" && req.Scope != "master" && req.Scope != "transaction" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid scope. Must be: all, master, or transaction",
		})
	}

	if !req.Backuped {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Harap buat backup sebelum menghapus data",
		})
	}

	result, err := h.backupService.DeleteData(user.CompanyID, req.Scope)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    result,
		"message": fmt.Sprintf("Berhasil menghapus %d data dari scope %s", result.TotalRecords, req.Scope),
	})
}

func (h *BackupHandler) GetTableCounts(c *fiber.Ctx) error {
	user := GetUserFromContext(c)
	if user == nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "Unauthorized",
		})
	}

	scope := c.Params("scope")
	if scope == "" {
		scope = "all"
	}

	if scope != "all" && scope != "master" && scope != "transaction" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid scope. Must be: all, master, or transaction",
		})
	}

	result, err := h.backupService.GetTableCounts(user.CompanyID, scope)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    result,
	})
}

func formatFileSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func formatDateTime(t time.Time) string {
	return t.Format("02 Jan 2006, 15:04")
}

func formatDate(t time.Time) string {
	return t.Format("02 Jan 2006")
}

func formatNumber(n int64) string {
	return strconv.FormatInt(n, 10)
}
