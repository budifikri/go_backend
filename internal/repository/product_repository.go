package repository

import (
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/pos-retail/go_backend/internal/models"
	"gorm.io/gorm"
)

type ProductHppTraceRow struct {
	Seq             int       `json:"seq"`
	EventDate       time.Time `json:"event_date"`
	EventType       string    `json:"event_type"`
	ReferenceID     uuid.UUID `json:"reference_id"`
	ReferenceNumber string    `json:"reference_number"`
	WarehouseID     uuid.UUID `json:"warehouse_id"`
	WarehouseName   string    `json:"warehouse_name"`
	Qty             float64   `json:"qty"`
	UnitCost        float64   `json:"unit_cost"`
	Hpp             float64   `json:"hpp"`
	Notes           string    `json:"notes"`
}

type ProductRepository struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

func (r *ProductRepository) FindAll(filters map[string]interface{}, limit, offset int) ([]models.Product, int64, error) {
	var products []models.Product
	var total int64

	query := r.db.Model(&models.Product{})

	// If is_active is not provided: include both active and inactive.
	if v, ok := filters["is_active"].(bool); ok {
		query = query.Where("is_active = ?", v)
	}
	if categoryID, ok := filters["category_id"].(string); ok && categoryID != "" {
		query = query.Where("category_id = ?", categoryID)
	}
	if companyID, ok := filters["company_id"].(string); ok && companyID != "" {
		query = query.Where("company_id = ?", companyID)
	}
	if search, ok := filters["search"].(string); ok && search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where("name ILIKE ? OR sku ILIKE ? OR barcode ILIKE ?", searchPattern, searchPattern, searchPattern)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Preload("Category").Preload("Unit").Limit(limit).Offset(offset).Order("created_at DESC").Find(&products).Error; err != nil {
		return nil, 0, err
	}

	return products, total, nil
}

func (r *ProductRepository) FindByID(id uuid.UUID) (*models.Product, error) {
	var product models.Product
	if err := r.db.Preload("Category").Preload("Unit").First(&product, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &product, nil
}

func (r *ProductRepository) FindBySKU(sku string) (*models.Product, error) {
	var product models.Product
	if err := r.db.Where("sku = ?", sku).First(&product).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &product, nil
}

func (r *ProductRepository) FindByBarcode(barcode string) (*models.Product, error) {
	var product models.Product
	if err := r.db.Where("barcode = ?", barcode).First(&product).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &product, nil
}

func (r *ProductRepository) Create(product *models.Product) error {
	return r.db.Create(product).Error
}

func (r *ProductRepository) Update(product *models.Product) error {
	log.Printf("[DEBUG] ProductRepo.Update: id=%s, UnitID=%s", product.ID, product.UnitID)
	err := r.db.Model(product).Select("sku", "barcode", "unit_id", "category_id", "name", "description", "cost_price", "retail_price", "tax_rate", "reorder_point", "is_active", "status", "updated_at").Updates(product).Error
	if err != nil {
		log.Printf("[DEBUG] ProductRepo.Update: error=%v", err)
	}
	return err
}

func (r *ProductRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Product{}, "id = ?", id).Error
}

func (r *ProductRepository) FindPriceTiers(productID uuid.UUID) ([]models.PriceTier, error) {
	var tiers []models.PriceTier
	if err := r.db.Where("product_id = ?", productID).Order("min_quantity ASC").Find(&tiers).Error; err != nil {
		return nil, err
	}
	return tiers, nil
}

func (r *ProductRepository) CreatePriceTier(tier *models.PriceTier) error {
	return r.db.Create(tier).Error
}

func (r *ProductRepository) DeletePriceTier(id uuid.UUID) error {
	return r.db.Delete(&models.PriceTier{}, "id = ?", id).Error
}

func (r *ProductRepository) FindOpenedProductIDs(productIDs []uuid.UUID) (map[uuid.UUID]bool, error) {
	opened := make(map[uuid.UUID]bool, len(productIDs))
	if len(productIDs) == 0 {
		return opened, nil
	}

	var rows []struct {
		ProductID uuid.UUID `gorm:"column:product_id"`
	}
	err := r.db.Table("stock_opname_items soi").
		Select("DISTINCT soi.product_id").
		Joins("JOIN stock_opnames so ON so.id = soi.opname_id").
		Where("soi.product_id IN ?", productIDs).
		Where("so.is_opening = ?", true).
		Where("LOWER(so.status) IN ?", []string{"approved", "posted", "completed"}).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	for _, row := range rows {
		opened[row.ProductID] = true
	}

	return opened, nil
}

func (r *ProductRepository) GetHppTrace(productID uuid.UUID) ([]ProductHppTraceRow, error) {
	const query = `
WITH RECURSIVE base_events AS (
	SELECT
		soi.product_id,
		COALESCE(so.opname_date, so.created_at) AS event_date,
		'OPENING_STOCK' AS event_type,
		so.id AS reference_id,
		so.opname_number AS reference_number,
		so.warehouse_id,
		COALESCE(w.name, '-') AS warehouse_name,
		soi.actual_quantity::numeric AS qty,
		soi.cost_price::numeric AS unit_cost,
		COALESCE(NULLIF(so.notes, ''), 'Opening stock') AS notes,
		1 AS source_priority
	FROM stock_opname_items soi
	JOIN stock_opnames so ON so.id = soi.opname_id
	LEFT JOIN warehouses w ON w.id = so.warehouse_id
	WHERE soi.product_id = ?
	  AND so.is_opening = true
	  AND LOWER(so.status) IN ('posted', 'completed')

	UNION ALL

	SELECT
		poi.product_id,
		COALESCE(po.receive_date, po.updated_at, po.created_at) AS event_date,
		'PURCHASE_RECEIVE' AS event_type,
		po.id AS reference_id,
		po.po_number AS reference_number,
		po.warehouse_id,
		COALESCE(w.name, '-') AS warehouse_name,
		poi.qty_receive::numeric AS qty,
		poi.unit_price::numeric AS unit_cost,
		COALESCE(NULLIF(po.note_receive, ''), 'Receive from PO') AS notes,
		2 AS source_priority
	FROM purchase_order_items poi
	JOIN purchase_orders po ON po.id = poi.po_id
	LEFT JOIN warehouses w ON w.id = po.warehouse_id
	WHERE poi.product_id = ?
	  AND COALESCE(poi.qty_receive, 0) > 0
	  AND LOWER(po.status_receive) = 'receive'
), ordered_events AS (
	SELECT
		*,
		ROW_NUMBER() OVER (
			ORDER BY event_date, source_priority, reference_number, reference_id
		) AS seq
	FROM base_events
), running_hpp AS (
	SELECT
		seq,
		event_date,
		event_type,
		reference_id,
		reference_number,
		warehouse_id,
		warehouse_name,
		qty,
		unit_cost,
		notes,
		qty::numeric(15,2) AS running_qty_after,
		(qty * unit_cost)::numeric(15,2) AS running_value_after,
		CASE
			WHEN qty <= 0 THEN unit_cost::numeric(15,2)
			ELSE unit_cost::numeric(15,2)
		END::numeric(15,2) AS running_hpp_after
	FROM ordered_events
	WHERE seq = 1

	UNION ALL

	SELECT
		e.seq,
		e.event_date,
		e.event_type,
		e.reference_id,
		e.reference_number,
		e.warehouse_id,
		e.warehouse_name,
		e.qty,
		e.unit_cost,
		e.notes,
		(r.running_qty_after + e.qty)::numeric(15,2) AS running_qty_after,
		(r.running_value_after + (e.qty * e.unit_cost))::numeric(15,2) AS running_value_after,
		CASE
			WHEN (r.running_qty_after + e.qty) <= 0 THEN e.unit_cost::numeric(15,2)
			ELSE ((r.running_value_after + (e.qty * e.unit_cost)) / NULLIF((r.running_qty_after + e.qty), 0))::numeric(15,2)
		END::numeric(15,2) AS running_hpp_after
	FROM running_hpp r
	JOIN ordered_events e ON e.seq = r.seq + 1
)
SELECT
	seq,
	event_date,
	event_type,
	reference_id,
	reference_number,
	warehouse_id,
	warehouse_name,
	qty,
	unit_cost,
	running_hpp_after AS hpp,
	notes
FROM running_hpp
ORDER BY seq;
`

	var rows []ProductHppTraceRow
	if err := r.db.Raw(query, productID, productID).Scan(&rows).Error; err != nil {
		return nil, err
	}

	return rows, nil
}
