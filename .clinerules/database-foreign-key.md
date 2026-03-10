## Brief overview
Pedoman untuk menangani foreign key constraint dan relasi database pada proyek Go backend dengan GORM.

## Database modeling
- Selalu definisikan foreign key constraint secara eksplisit pada model GORM
- Gunakan tag `references` untuk menentukan tabel dan kolom yang direferensikan
- Tambahkan tag `constraint:OnDelete:CASCADE` untuk menangani penghapusan data terkait

## Foreign key constraint
- Jangan hanya menggunakan `index` tanpa foreign key constraint
- Format: `gorm:"column:field_name;type:uuid;notNull;index;references:parent_table(id)"`
- Contoh: `CompanyID uuid.UUID \`gorm:"column:company_id;type:uuid;notNull;index;references:companies(id)" json:"company_id"\``

## Relationship definition
- Tambahkan struct relationship untuk navigasi antar tabel
- Format: `Company Company \`gorm:"foreignKey:CompanyID;constraint:OnDelete:CASCADE;" json:"-"\``
- Gunakan `json:"-"` untuk menghindari circular reference pada JSON response

## Error handling
- Tangani foreign key constraint violation dengan error message yang jelas
- Validasi keberadaan data parent sebelum membuat data child
- Gunakan transaction untuk operasi yang melibatkan multiple tables

## Service layer validation
- Selalu validasi keberadaan foreign key reference sebelum operasi CRUD
- Return error message yang spesifik untuk foreign key violation
- Contoh: "Company dengan ID tersebut tidak ditemukan"

## Migration strategy
- Gunakan AutoMigrate dengan hati-hati untuk foreign key constraint
- Pastikan constraint sudah ada di database sebelum menjalankan AutoMigrate
- Gunakan raw SQL untuk menambahkan constraint jika diperlukan

## Testing
- Buat test case untuk foreign key constraint violation
- Test scenario: create child tanpa parent, delete parent dengan child
- Test transaction rollback pada foreign key violation