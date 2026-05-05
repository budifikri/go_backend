-- Cek constraint foreign key yang ada
SELECT 
    tc.constraint_name, 
    kcu.column_name,
    ccu.table_name AS foreign_table_name,
    ccu.column_name AS foreign_column_name
FROM information_schema.table_constraints AS tc
JOIN information_schema.key_column_usage AS kcu
  ON tc.constraint_name = kcu.constraint_name
JOIN information_schema.constraint_column_usage AS ccu
  ON tc.constraint_name = ccu.constraint_name
WHERE tc.table_name = 'appointments'
  AND tc.constraint_type = 'FOREIGN KEY'
  AND tc.table_schema = CURRENT_SCHEMA();

-- Drop constraint lama jika mengacu ke products
ALTER TABLE appointments DROP CONSTRAINT IF EXISTS fk_appointments_treatment;
ALTER TABLE appointments DROP CONSTRAINT IF EXISTS appointments_treatment_id_fkey;

-- Pastikan tabel treatments sudah ada
SELECT COUNT(*) FROM treatments;

-- Insert sample treatments jika belum ada (ganti company_id dengan yang valid)
-- INSERT INTO treatments (id, name, duration, price, is_active, created_at, updated_at)
-- VALUES 
--   (gen_random_uuid(), 'Facial Treatment', 60, 150000, true, NOW(), NOW()),
--   (gen_random_uuid(), 'Massage Therapy', 90, 200000, true, NOW(), NOW()),
--   (gen_random_uuid(), 'Hair Treatment', 45, 100000, true, NOW(), NOW());

-- Test insert appointment dengan treatment_id yang valid
-- Pastikan treatment_id berasal dari tabel treatments
