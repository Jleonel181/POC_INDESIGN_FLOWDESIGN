-- ============================================
-- Reset completo de la base de datos designflow_db
-- Borra todos los datos y reinicia los autoincrementables a 1.
-- Ejecutar con: psql -U <user> -d designflow_db -f reset.sql
-- ============================================

TRUNCATE TABLE pautas, pages, editions RESTART IDENTITY CASCADE;
