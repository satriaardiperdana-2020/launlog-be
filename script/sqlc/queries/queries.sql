-- ==================== AUTH & USERS ====================
-- name: CreateUser :one
INSERT INTO users (name, email, password_hash, role_id, is_active)
VALUES ($1, $2, $3, (SELECT id FROM roles WHERE name = 'kasir' LIMIT 1), true)
RETURNING id, name, email, created_at;

-- name: GetUserByID :one
SELECT id, name, email, role_id, is_active, created_at
FROM users
WHERE id = $1 AND is_active = true;

-- name: GetUserByEmail :one
SELECT id, name, email, password_hash, role_id, is_active, created_at
FROM users
WHERE email = $1 AND is_active = true;

-- ==================== BLACKLIST TOKENS ====================
-- name: AddTokenToBlacklist :exec
INSERT INTO blacklisted_tokens (jti, expires_at)
VALUES ($1, $2);

-- name: IsTokenBlacklisted :one
SELECT EXISTS(SELECT 1 FROM blacklisted_tokens WHERE jti = $1) AS blacklisted;

-- ==================== CUSTOMERS ====================
-- name: CreateCustomer :one
INSERT INTO customers (name, phone, address)
VALUES ($1, $2, $3)
    RETURNING *;


-- name: ListCustomers :many
SELECT * FROM customers
WHERE is_active = true
  AND ($1::text = '' OR name ILIKE '%' || $1 || '%' OR phone ILIKE '%' || $1 || '%')
ORDER BY name;

-- name: GetCustomerById :one
SELECT * FROM customers WHERE id = $1 AND is_active = true;

-- name: UpdateCustomer :one
UPDATE customers
SET name = $2,
    phone = $3,
    address = $4,
    is_active = $5,
    updated_at = NOW()
WHERE id = $1
    RETURNING *;

-- name: SoftDeleteCustomer :one
UPDATE customers
SET
    is_active = false,
    updated_at = NOW()
WHERE id = $1
    RETURNING *;

-- ==================== SERVICE CATEGORIES ====================
-- name: CreateServiceCategory :one
INSERT INTO service_categories (name, description)
VALUES ($1, $2)
    RETURNING *;

-- name: ListServiceCategories :many
select
    id,
    name,
    description
from
    service_categories
where
    is_active = true
order by
    sort_order,
    name;

-- name: GetServiceCategoryByID :one
SELECT id,
       name,
       description
FROM service_categories
WHERE id = $1 AND is_active = true;

-- name: UpdateServiceCategory :one
UPDATE service_categories
SET name = $2, description = $3, sort_order = $4, updated_at = NOW()
WHERE id = $1
    RETURNING *;

-- name: SoftDeleteServiceCategory :one
UPDATE service_categories
SET is_active = false, updated_at = NOW()
WHERE id = $1 AND is_active = true
    RETURNING *;

-- ==================== SERVICES ====================
-- name: CreateService :one
INSERT INTO services (category_id, name, price, estimation, min_quantity, unit, description)
VALUES ($1, $2, $3, $4, $5, $6, $7)
    RETURNING *;

-- name: ListServices :many
SELECT
    s.id, sc.name as category_name, s.name as service_name, s.price, s.estimation, s.min_quantity, s.unit, s.description, s.is_active
FROM services s
         JOIN service_categories sc ON s.category_id = sc.id
WHERE s.is_active = true
ORDER BY s.sort_order, s.name;

-- name: GetServiceByID :one
SELECT
    s.id, sc.name as category_name, s.name as service_name, s.price, s.estimation, s.min_quantity, s.unit, s.description, s.is_active
FROM services s
         JOIN service_categories sc ON s.category_id = sc.id
WHERE  s.id = $1 AND s.is_active = true;

-- name: GetServiceDetail :one
SELECT
    s.id,
    s.category_id,
    sc.name as category_name,
    s.name,
    s.price,
    s.estimation,
    s.min_quantity,
    s.unit,
    COALESCE(s.description, '') as description,
    s.is_active,
    s.created_at,
    s.updated_at
FROM services s
JOIN service_categories sc ON s.category_id = sc.id
WHERE s.id = $1;

-- name: UpdateService :one
UPDATE services
SET category_id = $2, name = $3, price = $4, estimation = $5,
    min_quantity = $6, unit = $7, description = $8,
    updated_at = NOW()
WHERE id = $1
    RETURNING *;

-- name: SoftDeleteService :one
UPDATE services
SET is_active = false, updated_at = NOW()
WHERE id = $1 AND is_active = true
    RETURNING *;

-- ==================== TRANSACTIONS ====================
-- name: CreateTransaction :one
INSERT INTO transactions (invoice_no, type, user_id, customer_id, payment_status, is_delivery, total_amount, notes)
VALUES ($1, 'income', $2, $3, 'unpaid', $4, $5, $6)
    RETURNING *;

-- name: CreateTransactionItem :one
INSERT INTO transaction_items (transaction_id, service_id, qty, unit, unit_price, notes)
SELECT $1, $2, $3, s.unit, s.price, $4
FROM services s WHERE s.id = $2
    RETURNING *;

-- name: GetTodayIncomeExpense :one
SELECT
    COALESCE(SUM(CASE WHEN type='income' THEN total_amount ELSE 0 END),0) as today_income,
    COALESCE(SUM(CASE WHEN type='expenditure' THEN total_amount ELSE 0 END),0) as today_expense
FROM transactions
WHERE DATE(transaction_date) = CURRENT_DATE;