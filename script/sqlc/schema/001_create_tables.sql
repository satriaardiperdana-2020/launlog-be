-- ========================
-- ROLES (menggunakan BIGSERIAL)
-- ========================
CREATE TABLE roles (
id          BIGSERIAL PRIMARY KEY,
name        VARCHAR(50)  NOT NULL UNIQUE,
description VARCHAR(255),
is_active   BOOLEAN      NOT NULL DEFAULT true,
created_at  timestamptz  NOT NULL DEFAULT NOW(),
updated_at  timestamptz  NOT NULL DEFAULT NOW()
);

-- user pos bisa awner role admin atau kasir tidak semua menu dikasih - > untuk login
CREATE TABLE users (
id            BIGSERIAL    PRIMARY KEY,
name          VARCHAR(255),
email         VARCHAR(255) NOT NULL UNIQUE,
phone         VARCHAR(20),
password_hash TEXT         NOT NULL,
role_id       BIGINT       NOT NULL REFERENCES roles(id),
is_active     BOOLEAN      NOT NULL DEFAULT true,
created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- ========================
-- MENUS (menggunakan BIGSERIAL)
-- ========================
CREATE TABLE menus (
id         BIGSERIAL PRIMARY KEY,
name       VARCHAR(100) NOT NULL,
url        VARCHAR(255),
icon       VARCHAR(100),
parent_id  BIGINT REFERENCES menus(id) ON DELETE SET NULL,
sort_order INT          NOT NULL DEFAULT 0,
is_active  BOOLEAN      NOT NULL DEFAULT true,
assigned_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ========================
-- ROLE_MENUS (hanya can_view, tanpa id terpisah)
-- ========================
CREATE TABLE role_menus (
id          BIGSERIAL PRIMARY KEY,
role_id     BIGINT NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
menu_id     BIGINT NOT NULL REFERENCES menus(id) ON DELETE CASCADE,
can_view    BOOLEAN NOT NULL DEFAULT false,
can_create  BOOLEAN NOT NULL DEFAULT false,
can_edit    BOOLEAN NOT NULL DEFAULT false,
can_delete  BOOLEAN NOT NULL DEFAULT false,
assigned_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
UNIQUE (role_id, menu_id)
);

-- ========================
-- CUSTOMERS (pelanggan) TAK ADA LOGIN POS
-- ========================
CREATE TABLE customers (
id         BIGSERIAL PRIMARY KEY,
name       VARCHAR(150) NOT NULL,
phone      VARCHAR(20),
address    TEXT,
is_active   BOOLEAN NOT NULL DEFAULT true,
created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ========================
-- SERVICE_CATEGORIES
-- ========================
CREATE TABLE service_categories (
id          BIGSERIAL    PRIMARY KEY,
name        VARCHAR(50)  NOT NULL UNIQUE,
description TEXT,
is_active   BOOLEAN      NOT NULL DEFAULT true,
sort_order  INT          NOT NULL DEFAULT 0,
updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- ========================
-- SERVICES
-- ========================
CREATE TABLE services (
id            BIGSERIAL     PRIMARY KEY,
category_id   BIGINT        NOT NULL REFERENCES service_categories(id) ON DELETE RESTRICT,
name          VARCHAR(150)  NOT NULL,
price         NUMERIC(12,2) NOT NULL DEFAULT 0,
estimation    VARCHAR(50)   NOT NULL,
min_quantity  NUMERIC(10,2) NOT NULL DEFAULT 1,
unit          VARCHAR(20)   NOT NULL DEFAULT 'kg',
description   TEXT,
is_active     BOOLEAN       NOT NULL DEFAULT true,
sort_order    INT           NOT NULL DEFAULT 0,
created_at    TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
updated_at    TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

-- ========================
-- PAYMENT METHODS (metode pembayaran)
-- ========================
CREATE TABLE payment_methods (
id         BIGSERIAL    PRIMARY KEY,
name       VARCHAR(100) NOT NULL UNIQUE,
is_active  BOOLEAN      NOT NULL DEFAULT true,
sort_order INT          NOT NULL DEFAULT 0,
created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
updated_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- ========================
-- TRANSACTIONS
-- ========================
CREATE TABLE transactions (
id                BIGSERIAL     PRIMARY KEY,
invoice_no        VARCHAR(50)   NOT NULL UNIQUE,
type              VARCHAR(20)   NOT NULL DEFAULT 'income'
CHECK (type IN ('income', 'expenditure')),
user_id           BIGINT        NOT NULL REFERENCES users(id),
customer_id       BIGINT        REFERENCES customers(id),
payment_method_id BIGINT        REFERENCES payment_methods(id),
payment_status    VARCHAR(20)   NOT NULL DEFAULT 'unpaid'
CHECK (payment_status IN ('unpaid', 'dp', 'paid')),
is_delivery       BOOLEAN       NOT NULL DEFAULT false,
paid_amount       NUMERIC(12,2) NOT NULL DEFAULT 0,
supplier          VARCHAR(150),
total_amount      NUMERIC(12,2) NOT NULL DEFAULT 0,
notes             TEXT,
transaction_date  TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
created_at        TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
updated_at        TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

-- ========================
-- TRANSACTION_ITEMS
-- ========================
CREATE TABLE transaction_items (
id             BIGSERIAL     PRIMARY KEY,
transaction_id BIGINT        NOT NULL REFERENCES transactions(id) ON DELETE CASCADE,
service_id     BIGINT        REFERENCES services(id),
item_name      VARCHAR(255),
qty            NUMERIC(10,2) NOT NULL DEFAULT 1,
unit           VARCHAR(20),
unit_price     NUMERIC(12,2) NOT NULL DEFAULT 0,
subtotal       NUMERIC(12,2) GENERATED ALWAYS AS (qty * unit_price) STORED,
notes          TEXT
);

-- ========================
-- TRANSACTION_PAYMENTS
-- ========================
CREATE TABLE transaction_payments (
id                BIGSERIAL     PRIMARY KEY,
transaction_id    BIGINT        NOT NULL REFERENCES transactions(id) ON DELETE CASCADE,
payment_method_id BIGINT        NOT NULL REFERENCES payment_methods(id),
amount            NUMERIC(12,2) NOT NULL CHECK (amount > 0),
payment_type      VARCHAR(20)   NOT NULL DEFAULT 'dp'
CHECK (payment_type IN ('dp', 'settlement')),
paid_at           TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
received_by       BIGINT        NOT NULL REFERENCES users(id),
notes             TEXT,
created_at        TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

-- ========================
-- TRANSACTION_DELIVERIES
-- ========================
CREATE TABLE transaction_deliveries (
id             BIGSERIAL     PRIMARY KEY,
transaction_id BIGINT        NOT NULL REFERENCES transactions(id) ON DELETE CASCADE,
type           VARCHAR(20)   NOT NULL CHECK (type IN ('pickup', 'dropoff')),
status         VARCHAR(30)   NOT NULL DEFAULT 'pending'
CHECK (status IN ('pending','scheduled','on_the_way','done','cancelled')),
address        TEXT          NOT NULL,
scheduled_at   TIMESTAMPTZ,
completed_at   TIMESTAMPTZ,
courier_name   VARCHAR(100),
delivery_fee   NUMERIC(12,2) NOT NULL DEFAULT 0,
notes          TEXT,
created_at     TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
updated_at     TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
UNIQUE (transaction_id, type)
);