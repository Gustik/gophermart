-- Откат миграции - удаление таблиц в обратном порядке
DROP INDEX IF EXISTS idx_withdrawals_processed_at;
DROP INDEX IF EXISTS idx_withdrawals_user_id;
DROP TABLE IF EXISTS withdrawals;

DROP TABLE IF EXISTS balance;

DROP INDEX IF EXISTS idx_orders_uploaded_at;
DROP INDEX IF EXISTS idx_orders_status;
DROP INDEX IF EXISTS idx_orders_user_id;
DROP TABLE IF EXISTS orders;

DROP TABLE IF EXISTS users;