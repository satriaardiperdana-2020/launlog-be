-- name: CreateUser :one
INSERT INTO users (name, email, password_hash, role_id, is_active)
VALUES ($1, $2, $3, (SELECT id FROM roles WHERE name = 'kasir' LIMIT 1), true)
RETURNING id, name, email, created_at;

-- name: GetUserByEmail :one
SELECT id, name, email, password_hash, role_id, is_active, created_at
FROM users
WHERE email = $1 AND is_active = true;

-- name: CreateCustomer :one
INSERT INTO customers (name, phone, address)
VALUES ($1, $2, $3)
    RETURNING *;

-- name: ListCustomers :many
SELECT * FROM customers
WHERE ($1::text = '' OR name ILIKE '%' || $1 || '%' OR phone ILIKE '%' || $1 || '%')
ORDER BY name;

-- name: GetCustomerById :one
SELECT * FROM customers WHERE id = $1;

-- name: GetServiceById :one
SELECT * FROM services WHERE id = $1;

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