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
    s.id,
    s.category_id,
    sc.name as category_name,
    s.name as service_name,
    s.price,
    s.estimation,
    s.min_quantity,
    s.unit,
    COALESCE(s.description, '') as description,
    s.is_active,
    s.sort_order,
    s.created_at,
    s.updated_at
FROM services s
         JOIN service_categories sc ON s.category_id = sc.id
WHERE s.is_active = true
  AND (sqlc.narg('category_id')::bigint IS NULL OR s.category_id = sqlc.narg('category_id')::bigint)
  AND (sqlc.narg('search')::text IS NULL OR s.name ILIKE '%' || sqlc.narg('search')::text || '%')
ORDER BY s.sort_order, s.name;

-- name: GetServiceById :one
SELECT
    s.id,
    s.category_id,
    sc.name as category_name,
    s.name as service_name,
    s.price,
    s.estimation,
    s.min_quantity,
    s.unit,
    COALESCE(s.description, '') as description,
    s.is_active,
    s.sort_order,
    s.created_at,
    s.updated_at
FROM services s
         JOIN service_categories sc ON s.category_id = sc.id
WHERE s.id = $1 AND s.is_active = true;

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
SET
    category_id = sqlc.arg('category_id')::bigint,
    name = sqlc.arg('name')::text,
    price = sqlc.arg('price')::numeric,
    estimation = sqlc.arg('estimation')::text,
    min_quantity = sqlc.arg('min_quantity')::numeric,
    unit = sqlc.arg('unit')::text,
    description = COALESCE(sqlc.narg('description')::text, description),
    sort_order = sqlc.arg('sort_order')::int,
    updated_at = NOW()
WHERE id = sqlc.arg('id')
    RETURNING *;

-- name: SoftDeleteService :one
UPDATE services
SET is_active = false, updated_at = NOW()
WHERE id = $1 AND is_active = true
    RETURNING *;

-- ==================== INCOME / TRANSACTIONS ====================

-- name: CreateTransaction :one
INSERT INTO transactions (
    invoice_no, transaction_type, user_id, customer_id,
    payment_status, is_delivery, total_amount, notes
)
VALUES ($1, 'income', $2, $3, 'unpaid', $4, $5, $6)
    RETURNING *;

-- name: CreateTransactionItem :one
INSERT INTO transaction_items (
    transaction_id, service_id, qty, unit, unit_price, notes
)
SELECT $1, $2, $3, s.unit, s.price, $4
FROM services s
WHERE s.id = $2
    RETURNING *;

-- name: GetTransactionById :one
SELECT
    t.*,
    ti.id as item_id,
    ti.service_id,
    ti.item_name,
    ti.qty,
    ti.unit,
    ti.unit_price,
    ti.subtotal,
    ti.notes as item_notes,
    s.name as service_name
FROM transactions t
         LEFT JOIN transaction_items ti ON t.id = ti.transaction_id
         LEFT JOIN services s ON ti.service_id = s.id
WHERE t.id = $1;

-- name: ListTransactions :many
SELECT * FROM transactions
WHERE ($1::date IS NULL OR DATE(transaction_date) >= $1::date)
  AND ($2::date IS NULL OR DATE(transaction_date) <= $2::date)
  AND ($3::text IS NULL OR transaction_type = $3::text)
ORDER BY transaction_date DESC;

-- ==================== EXPENSE ====================

-- name: CreateExpense :one
INSERT INTO transactions (
    invoice_no, transaction_type, user_id, supplier,
    expense_category, paid_amount, total_amount, payment_status, notes
)
VALUES ($1, 'expenditure', $2, $3, $4, $5, $6, $7, $8)
    RETURNING *;

-- name: CreateExpenseItem :one
INSERT INTO transaction_items (
    transaction_id, item_name, qty, unit_price, notes
)
VALUES ($1, $2, $3, $4, $5)
    RETURNING *;

-- name: ListExpenses :many
SELECT * FROM transactions
WHERE transaction_type = 'expenditure'
  AND (sqlc.arg('start_date')::date IS NULL OR DATE(transaction_date) >= sqlc.arg('start_date')::date)
  AND (sqlc.arg('end_date')::date IS NULL OR DATE(transaction_date) <= sqlc.arg('end_date')::date)
  AND is_deleted = false
ORDER BY transaction_date DESC;

-- name: GetExpenseById :one
SELECT * FROM transactions
WHERE id = $1 AND transaction_type = 'expenditure';

-- update and sofdelete expense income
-- ==================== UPDATE INCOME ====================
-- name: UpdateIncome :one
UPDATE transactions
SET
    customer_id = COALESCE(sqlc.narg('customer_id')::bigint, customer_id),
    payment_status = COALESCE(sqlc.narg('payment_status')::text, payment_status),
    is_delivery = COALESCE(sqlc.narg('is_delivery')::bool, is_delivery),
    total_amount = COALESCE(sqlc.narg('total_amount')::numeric, total_amount),
    notes = COALESCE(sqlc.narg('notes')::text, notes),
    updated_at = NOW()
WHERE id = sqlc.arg('id')
  AND transaction_type = 'income'
  AND is_deleted = false
    RETURNING *;

-- name: UpdateIncomeItem :one
UPDATE transaction_items
SET
    qty = COALESCE(sqlc.narg('qty')::numeric, qty),
    unit_price = COALESCE(sqlc.narg('unit_price')::numeric, unit_price),
    notes = COALESCE(sqlc.narg('notes')::text, notes)
WHERE id = sqlc.arg('id')
    RETURNING *;

-- ==================== UPDATE EXPENSE ====================
-- name: UpdateExpense :one
UPDATE transactions
SET
    supplier = COALESCE(sqlc.narg('supplier')::text, supplier),
    expense_category = COALESCE(sqlc.narg('expense_category')::text, expense_category),
    total_amount = COALESCE(sqlc.narg('total_amount')::numeric, total_amount),
    paid_amount = COALESCE(sqlc.narg('paid_amount')::numeric, paid_amount),
    notes = COALESCE(sqlc.narg('notes')::text, notes),
    updated_at = NOW()
WHERE id = sqlc.arg('id')
  AND transaction_type = 'expenditure'
  AND is_deleted = false
    RETURNING *;

-- name: UpdateExpenseItem :one
UPDATE transaction_items
SET
    item_name = COALESCE(sqlc.narg('item_name')::text, item_name),
    qty = COALESCE(sqlc.narg('qty')::numeric, qty),
    unit_price = COALESCE(sqlc.narg('unit_price')::numeric, unit_price),
    notes = COALESCE(sqlc.narg('notes')::text, notes)
WHERE id = sqlc.arg('id')
    RETURNING *;

-- name: SoftDeleteExpense :one
UPDATE transactions
SET
    is_deleted = true,
    deleted_at = NOW(),
    updated_at = NOW()
WHERE id = $1
  AND transaction_type = 'expenditure'
  AND is_deleted = false
    RETURNING *;



-- ==================== SEARCH BY DATE RANGE ====================
-- name: ListTransactionsByDateRange :many
SELECT * FROM transactions
WHERE ($1::date IS NULL OR DATE(transaction_date) >= $1::date)
  AND ($2::date IS NULL OR DATE(transaction_date) <= $2::date)
  AND ($3::text IS NULL OR transaction_type = $3::text)
  AND is_deleted = false
ORDER BY transaction_date DESC;



--====================TRANSACTION REPORT =====================
-- name: GetTransactionReport :many
SELECT
    t.id,
    t.invoice_no,
    t.transaction_type,
    t.user_id,
    t.customer_id,
    t.payment_status,
    t.total_amount,
    t.paid_amount,
    t.notes,
    t.transaction_date,
    COALESCE(c.name, '') as customer_name,
    COALESCE(t.supplier, '') as supplier,
    COALESCE(t.expense_category, '') as expense_category
FROM transactions t
         LEFT JOIN customers c ON t.customer_id = c.id
WHERE (sqlc.arg('start_date')::date IS NULL OR DATE(t.transaction_date) >= sqlc.arg('start_date')::date)
  AND (sqlc.arg('end_date')::date IS NULL OR DATE(t.transaction_date) <= sqlc.arg('end_date')::date)
  AND (sqlc.arg('type')::text IS NULL OR t.transaction_type = sqlc.arg('type')::text)
  AND t.is_deleted = false
ORDER BY t.transaction_date DESC;
-- name: GetTransactionSummary :one
SELECT
    COALESCE(SUM(CASE WHEN transaction_type = 'income' THEN total_amount ELSE 0 END), 0) as total_income,
    COALESCE(SUM(CASE WHEN transaction_type = 'expenditure' THEN total_amount ELSE 0 END), 0) as total_expense,
    COALESCE(SUM(CASE WHEN transaction_type = 'income' THEN total_amount ELSE -total_amount END), 0) as net_profit
FROM transactions
WHERE (sqlc.arg('start_date')::date IS NULL OR DATE(transaction_date) >= sqlc.arg('start_date')::date)
  AND (sqlc.arg('end_date')::date IS NULL OR DATE(transaction_date) <= sqlc.arg('end_date')::date)
  AND is_deleted = false;
-- ==================== PAYMENTS ====================

-- name: AddPayment :one
INSERT INTO transaction_payments (
    transaction_id, payment_method_id, amount, payment_type, received_by, notes
)
VALUES ($1, $2, $3, $4, $5, $6)
    RETURNING *;

-- name: UpdateTransactionPaymentStatus :one
UPDATE transactions
SET
    payment_status = $2,
    paid_amount = paid_amount + $3,
    payment_method_id = $4,
    updated_at = NOW()
WHERE id = $1
    RETURNING *;