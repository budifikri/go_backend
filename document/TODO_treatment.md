# TODO: Master Data Treatment Implementation

## Status Legend
- ✅ Completed
- ❌ Pending
- 🔄 In Progress

---

## Backend Tasks

### Database & Models
- [✅] Buat model treatment (`internal/models/treatment.go`)
  - Treatment struct (id, name, company_id, duration, price, description, is_active, created_at)
  - TreatmentTag struct (id, name)
  - TreatmentTagRelation struct (treatment_id, tag_id)
  - TableName() methods
  - BeforeCreate() hooks for UUID generation

### Repository Layer
- [✅] Buat repository (`internal/repository/treatment_repository.go`)
  - FindAll() with filters (search, tag_id, is_active, company_id)
  - FindByID() with Preload Tags
  - Create(), Update(), Delete()
  - Tag methods: FindAllTags(), CreateTag(), UpdateTag(), DeleteTag()
  - Relation methods: DeleteTagRelationsByTreatmentID(), CreateTagRelation()

### Service Layer
- [✅] Buat service (`internal/services/treatment_service.go`)
  - TreatmentListResponse & TreatmentDetailResponse structs
  - TagResponse struct
  - GetTreatments() - pagination + filtering
  - GetTreatmentByID() - with tags
  - CreateTreatment() - with tag relations
  - UpdateTreatment() - sync tags
  - DeleteTreatment()
  - GetTags(), CreateTag(), UpdateTag(), DeleteTag()

### Handler Layer
- [✅] Buat handler (`internal/handlers/treatment_handler.go`)
  - GetTreatments() - GET /api/treatments
  - GetTreatment() - GET /api/treatments/:id
  - CreateTreatment() - POST /api/treatments
  - UpdateTreatment() - PUT /api/treatments/:id
  - DeleteTreatment() - DELETE /api/treatments/:id
  - GetTags() - GET /api/treatment-tags
  - CreateTag() - POST /api/treatment-tags
  - UpdateTag() - PUT /api/treatment-tags/:id
  - DeleteTag() - DELETE /api/treatment-tags/:id

### Main Integration
- [✅] Update `cmd/server/main.go`
  - Add models to AutoMigrate (Treatment, TreatmentTag, TreatmentTagRelation)
  - Initialize treatment repository, service, handler
  - Register routes group `/treatments` and `/treatment-tags`

### Backend Verification
- [❌] Run `go mod tidy` to ensure dependencies
- [❌] Build backend: `cd go_backend && go build ./...`
- [❌] Test migration: Start server and verify tables created
- [❌] Test API endpoints (Postman/curl):
  - GET /api/treatments
  - POST /api/treatments
  - GET /api/treatments/:id
  - PUT /api/treatments/:id
  - DELETE /api/treatments/:id
  - GET /api/treatment-tags
  - POST /api/treatment-tags

---

## Frontend Tasks

### API Layer
- [❌] Buat `src/features/master/treatment/treatment.api.js`
  - listTreatments(token, params) function
  - createTreatment(token, input) function
  - updateTreatment(token, id, input) function
  - deleteTreatment(token, id) function
  - listTreatmentTags(token) function
  - createTreatmentTag(token, name) function
  - updateTreatmentTag(token, id, name) function
  - deleteTreatmentTag(token, id) function
  - Pattern: Copy from `product.api.js` or `customer.api.js`

### Components
- [❌] Buat `src/components/ToolbarItem/master/Treatment.jsx`
  - Import patterns from Product.jsx or Customer.jsx
  - State management (data, pagination, form, selectedId, etc.)
  - Table with columns: NO, NAME, DURATION, PRICE, TAGS, STATUS
  - Form fields: name, duration, price, description, tags (multi-select)
  - Dummy data for offline mode
  - Keyboard navigation (F1, F2, Delete, Ctrl+←/→, Escape)
  - Search, filter by status
  - Toggle status (MasterStatusToggle)
  - Delete confirmation modal
  - Export to Excel, Import from Excel
  - Print report

### Tags Management
- [❌] Buat Tags Modal dalam Treatment.jsx atau separate component
  - State: tags list, newTag form, editTag form
  - Table: Name, Created Date, Actions
  - CRUD operations for tags
  - Integrate with treatment form (multi-select)

### Integration

#### Toolbar Items
- [❌] Update `src/data/toolbarItems.js`
  - Add to `master` array:
    ```javascript
    { key: 'treatment', label: 'Treatment', mark: 'T', tone: 'teal', 
      filter: { businessType: ['clinic'], moduleCodes: ['clinic_core'] } },
    ```

#### App.jsx
- [❌] Update `src/App.jsx`
  - Add 'treatment' to IMPLEMENTED_TOOLS Set

#### Dashboard Canvas
- [❌] Update `src/components/Dashboard/DashboardCanvas.jsx`
  - Add case for 'treatment' key
  - Import Treatment component
  - Render `<Treatment onExit={handleExit} />`

---

## Testing Tasks

### Backend Testing
- [❌] Unit tests for treatment service (if applicable)
- [❌] Integration tests for treatment API endpoints
- [❌] Test tag relations (create treatment with tags, update tags)

### Frontend Testing
- [❌] Run `npm run lint` - check for errors
- [❌] Run `npm run build` - verify no build errors
- [❌] Manual testing:
  - Open Treatment master via menu
  - Add new treatment with tags
  - Edit treatment (change tags)
  - Delete treatment
  - Manage tags (add, edit, delete)
  - Search and filter
  - Export/Import Excel
  - Print report

---

## Documentation Tasks
- [✅] Buat PLAN_treatment.md (this document's companion)
- [✅] Buat TODO_treatment.md (this document)
- [❌] Update `AGENTS.md` if new patterns discovered
- [❌] Update API documentation (if using Swagger/OpenAPI)

---

## Deployment Tasks
- [❌] Commit backend changes
- [❌] Commit frontend changes
- [❌] Push to branch
- [❌] Create Pull Request
- [❌] Deploy to staging environment
- [❌] Verify in staging
- [❌] Deploy to production

---

## Progress Summary

| Category | Total | Completed | Pending | Progress |
|----------|-------|-----------|---------|----------|
| Backend Models | 1 | 1 | 0 | 100% |
| Backend Repository | 1 | 1 | 0 | 100% |
| Backend Service | 1 | 1 | 0 | 100% |
| Backend Handler | 1 | 1 | 0 | 100% |
| Backend Routes | 1 | 1 | 0 | 100% |
| Backend Build | 1 | 0 | 1 | 0% |
| Frontend API | 1 | 0 | 1 | 0% |
| Frontend Component | 1 | 0 | 1 | 0% |
| Frontend Integration | 3 | 0 | 3 | 0% |
| Testing | 4 | 0 | 4 | 0% |
| **Total** | **15** | **5** | **10** | **33%** |

---

**Created:** 2026-05-05  
**Last Updated:** 2026-05-05  
**Overall Progress:** 33% (Backend Done, Frontend Pending)
