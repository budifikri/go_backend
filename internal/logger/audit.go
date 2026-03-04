package logger

import (
	"fmt"

	"gorm.io/gorm"
)

func AuditCreate(db *gorm.DB, logger *Logger, table, recordID, userID, companyID string, createFunc func() error) error {
	if logger == nil {
		return createFunc()
	}

	if err := createFunc(); err != nil {
		logger.LogError(ActionCreate, table, userID, companyID, recordID, err)
		return err
	}

	var newData map[string]interface{}
	if err := db.Table(table).Where("id = ?", recordID).Take(&newData).Error; err != nil {
		logger.LogError(ActionCreate, table, userID, companyID, recordID, err)
		return nil
	}

	logger.Log(ActionCreate, table, userID, companyID, recordID, nil, newData)
	return nil
}

func AuditUpdate(
	db *gorm.DB,
	logger *Logger,
	table string,
	recordID string,
	userID string,
	companyID string,
	updateFunc func() error,
) error {
	if logger == nil {
		return updateFunc()
	}

	var oldData map[string]interface{}
	if err := db.Table(table).Where("id = ?", recordID).Take(&oldData).Error; err != nil {
		return err
	}

	if err := updateFunc(); err != nil {
		logger.LogError(ActionUpdate, table, userID, companyID, recordID, err)
		return err
	}

	var newData map[string]interface{}
	if err := db.Table(table).Where("id = ?", recordID).Take(&newData).Error; err != nil {
		logger.LogError(ActionUpdate, table, userID, companyID, recordID, err)
		return nil
	}

	logger.Log(ActionUpdate, table, userID, companyID, recordID, oldData, newData)
	return nil
}

func AuditDelete(db *gorm.DB, logger *Logger, table, recordID, userID, companyID string, deleteFunc func() error) error {
	if logger == nil {
		return deleteFunc()
	}

	var oldData map[string]interface{}
	if err := db.Table(table).Where("id = ?", recordID).Take(&oldData).Error; err != nil {
		return err
	}

	if err := deleteFunc(); err != nil {
		logger.LogError(ActionDelete, table, userID, companyID, recordID, err)
		return err
	}

	logger.Log(ActionDelete, table, userID, companyID, recordID, oldData, nil)
	return nil
}

func MustID(recordID string) string {
	if recordID == "" {
		return "-"
	}
	return recordID
}

func WrapError(msg string, err error) error {
	if err == nil {
		return fmt.Errorf(msg)
	}
	return fmt.Errorf("%s: %w", msg, err)
}
