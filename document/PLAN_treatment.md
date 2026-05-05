# PLAN: Master Data Treatment Implementation

## Overview
Implementasi fitur Master Data Treatment untuk sistem POS Retail (fokus klinik/beauty clinic). Mengikuti pola desain yang sudah ada (Product, Customer).

---

## A. Backend (Go/Fiber) - STATUS: ✅ COMPLETED

### 1. Database Migration
- **File:** `internal/models/treatment.go` ✅
- **Tables:** 
  - `treatments` (id uuid, name, company_id, duration, price, description, is_active, created_at)
  - `treatment_tags` (id uuid, name)
  - `treatment_tag_relations` (treatment_id, tag_id)
- **AutoMigrate:** Sudah ditambahkan di `cmd/server/main.go`

### 2. Repository Layer
- **File:** `internal/repository/treatment_repository.go` ✅
- **Methods:**
  - `FindAll()` - list treatments dengan filter (search, tag, status)
  - `FindByID()` - get treatment by ID dengan preload tags
  - `Create()` - create treatment
  - `Update()` - update treatment
  - `Delete()` - delete treatment
  - Tag methods: `FindAllTags()`, `CreateTag()`, `UpdateTag()`, `DeleteTag()`
  - Relation methods: `DeleteTagRelationsByTreatmentID()`, `CreateTagRelation()`

### 3. Service Layer
- **File:** `internal/services/treatment_service.go` ✅
- **Response structs:** `TreatmentListResponse`, `TreatmentDetailResponse`, `TagResponse`
- **Methods:** 
  - `GetTreatments()` - pagination + filtering
  - `GetTreatmentByID()` - detail with tags
  - `CreateTreatment()` - create + attach tags
  - `UpdateTreatment()` - update + sync tags
  - `DeleteTreatment()` - soft/hard delete
  - Tag methods: `GetTags()`, `CreateTag()`, `UpdateTag()`, `DeleteTag()`

### 4. Handler Layer
- **File:** `internal/handlers/treatment_handler.go` ✅
- **Endpoints:**
  - `GET /api/treatments` - List treatments (filter: search, tag_id, is_active, pagination)
  - `GET /api/treatments/:id` - Get treatment detail
  - `POST /api/treatments` - Create treatment (body: name, duration, price, description, tag_ids[])
  - `PUT /api/treatments/:id` - Update treatment
  - `DELETE /api/treatments/:id` - Delete treatment
  - `GET /api/treatment-tags` - List all tags
  - `POST /api/treatment-tags` - Create tag (body: name)
  - `PUT /api/treatment-tags/:id` - Update tag
  - `DELETE /api/treatment-tags/:id` - Delete tag

### 5. Route Registration
- **File:** `cmd/server/main.go` ✅
- Treatment routes dan TreatmentTag routes sudah didaftarkan di Fiber app

---

## B. Frontend (React) - STATUS: ❌ PENDING

### 1. API Layer
- **File:** `src/features/master/treatment/treatment.api.js`
- **Functions:**
  ```javascript
  // Treatments
  listTreatments(token, params) // search, tag_id, is_active, limit, offset
  createTreatment(token, input) // name, duration, price, description, tag_ids[]
  updateTreatment(token, id, input)
  deleteTreatment(token, id)
  
  // Treatment Tags
  listTreatmentTags(token)
  createTreatmentTag(token, name)
  updateTreatmentTag(token, id, name)
  deleteTreatmentTag(token, id)
  ```

### 2. Main Component
- **File:** `src/components/ToolbarItem/master/Treatment.jsx`
- **Pattern:** Copy dari `Product.jsx` atau `Customer.jsx`
- **Table Columns:**
  ```javascript
  [
    { key: 'no', label: 'NO' },
    { key: 'name', label: 'NAME' },
    { key: 'duration', label: 'DURASI (min)' },
    { key: 'price', label: 'HARGA' },
    { key: 'tags', label: 'TAGS' },
    { key: 'is_active', label: 'STATUS' },
  ]
  ```
- **Form Fields:**
  - Nama Treatment (text input)
  - Durasi (number input, dalam menit)
  - Harga (number input)
  - Deskripsi (textarea)
  - Tags (multi-select dengan checkbox, + tombol manage tags)
- **Dummy Data:** 2-3 treatment untuk offline mode

### 3. Tags Management Modal
- Modal untuk CRUD treatment tags
- Panggil dari form treatment (tombol "Manage Tags")
- Table: Name, Created Date, Actions (Edit/Delete)
- Form: New tag name + Add button

### 4. Integration Points

#### a. Update `src/data/toolbarItems.js`
Tambahkan di bagian `master` array (setelah `paket`):
```javascript
{ key: 'treatment', label: 'Treatment', mark: 'T', tone: 'teal', 
  filter: { businessType: ['clinic'], moduleCodes: ['clinic_core'] } },
```

#### b. Update `src/App.jsx`
Tambahkan `'treatment'` ke `IMPLEMENTED_TOOLS` Set:
```javascript
const IMPLEMENTED_TOOLS = new Set(['warehouse', 'satuan', 'categori', 'product', 
  'customer', 'supplier', 'dokter', 'jadwal_dokter', 'paket', 'company', 
  'theme', 'user', 'lapstok', 'laphargagrosir', 'lapjual', 'lapbeli', 
  'opname', 'beli', 'receive', 'retur', 'promotion', 'lapcashdrawer', 
  'report_setting', 'backup', 'telegram', 'module', 'business_type', 
  'module_package', 'appointment', 'treatment'])
```

#### c. Update `src/components/Dashboard/DashboardCanvas.jsx`
Tambahkan mapping untuk `'treatment'` key ke `<Treatment />` component.

---

## C. Visual Design Sketch

### Treatment List View (Mengikuti Product)
```
+-------------------------------------------------------+
| Treatment                              [+ F1: Add New] |
+-------------------------------------------------------+
| Search: [_____________]  Status: [v Active ▼]          |
+-------------------------------------------------------+
| NO | NAME        | DURASI  | HARGA    | TAGS   | STATUS |
|----|-------------|---------|----------|--------|--------|
| 1  | Facial      | 60 min  | Rp 150K  | wajah  | Active |
| 2  | Massage    | 90 min  | Rp 200K  | badan  | Active |
| 3  | Refleksi   | 45 min  | Rp 120K  | badan  | Inactive|
+-------------------------------------------------------+
| [< Prev] Page 1 of 5 [Next >]  Total: 48 records    |
+-------------------------------------------------------+
```

### Treatment Form Modal
```
+-------------------------------------------------------+
| Treatment Form                                        |
+-------------------------------------------------------+
| Nama Treatment: [________________________]                |
| Durasi (menit): [____]                                |
| Harga:          [__________]                            |
| Deskripsi:      [________________________]                |
|                  [________________________]                |
| Tags:           [v Select Multiple  ] [+ Manage Tags]   |
|                  ☑ Tag1  ☑ Tag2  ☑ Tag3               |
+-------------------------------------------------------+
|                    [Batal] [Simpan]                     |
+-------------------------------------------------------+
```

### Treatment Tags Management Modal
```
+-------------------------------------------------------+
| Treatment Tags Management                              |
+-------------------------------------------------------+
| +---------+-----------+----------+                    |
| | Name    | Created   | Action   |                    |
| +---------+-----------+----------+                    |
| | Facial  | 2026-01-01| Edit Del |                    |
| | Massage | 2026-01-02| Edit Del |                    |
| +---------+-----------+----------+                    |
| New Tag: [_____________] [+ Add]                      |
+-------------------------------------------------------+
```

---

## D. Keyboard Shortcuts (Same as Product)
- **F1 / +** : Add New Treatment
- **F2** : Edit Selected Treatment
- **Delete** : Delete Selected Treatment
- **Ctrl + ←** : Previous Record (dalam form)
- **Ctrl + →** : Next Record (dalam form)
- **Escape** : Exit/Close Form

---

## E. Risk Analysis & Mitigation

| Risk | Mitigation |
|------|-----------|
| Migration fails | Backup database before migration, test in dev environment |
| Orphaned tag relations | Use transaction when creating/updating treatment with tags |
| Inconsistent frontend | Copy-paste pattern from Product.jsx, change fields only |
| Module access control | Ensure filter in toolbarItems is correct (clinic only) |
| Backend build error | Run `go mod tidy` before build |

---

## F. Next Steps

1. ✅ Backend: Verify build (`cd go_backend && go build ./...`)
2. ❌ Frontend: Create `treatment.api.js`
3. ❌ Frontend: Create `Treatment.jsx` component
4. ❌ Frontend: Update `toolbarItems.js`
5. ❌ Frontend: Update `App.jsx` and `DashboardCanvas.jsx`
6. ❌ Testing: `npm run lint` and `npm run build`

---

## G. Database Schema Summary

```sql
-- Treatments table
CREATE TABLE treatments (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  name text NOT NULL,
  company_id uuid,
  duration integer DEFAULT 60, -- in minutes
  price numeric DEFAULT 0,
  description text,
  is_active boolean DEFAULT true,
  created_at timestamp DEFAULT CURRENT_TIMESTAMP
);

-- Treatment tags table
CREATE TABLE treatment_tags (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  name varchar(50) UNIQUE NOT NULL
);

-- Treatment tag relations (many-to-many)
CREATE TABLE treatment_tag_relations (
  treatment_id uuid REFERENCES treatments(id),
  tag_id uuid REFERENCES treatment_tags(id),
  PRIMARY KEY (treatment_id, tag_id)
);
```

---

**Created:** 2026-05-05  
**Last Updated:** 2026-05-05  
**Status:** Backend Complete, Frontend Pending
