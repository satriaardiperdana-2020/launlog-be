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
    name          VARCHAR(150),
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
-- ==================== TRANSACTIONS TABLE ====================
-- Menyimpan semua transaksi (income/pemasukan dan expenditure/pengeluaran)

CREATE TABLE transactions (
    -- Primary Key
                              id BIGSERIAL PRIMARY KEY,
    -- Identifikasi Transaksi
                              invoice_no VARCHAR(50) NOT NULL UNIQUE,
                              transaction_type VARCHAR(20) DEFAULT 'income' NOT NULL , -- 'income' atau 'expense'
    -- Relasi ke User (kasir/admin)
                              user_id BIGINT NOT NULL REFERENCES users(id),
    -- Untuk Income (transaksi dengan pelanggan)
                              customer_id BIGINT NULL REFERENCES customers(id),
                              payment_method_id BIGINT NULL REFERENCES payment_methods(id),
                              payment_status VARCHAR(20) NOT NULL DEFAULT 'unpaid', -- 'unpaid', 'dp', 'paid'
                              is_delivery BOOLEAN NOT NULL DEFAULT FALSE,
    -- Untuk expense (pengeluaran)
                              supplier VARCHAR(150) NULL,
                              expense_category VARCHAR(50) NULL,
    -- Nilai Transaksi
                              total_amount NUMERIC(12,2) NOT NULL DEFAULT 0,
                              paid_amount NUMERIC(12,2) NOT NULL DEFAULT 0,
    -- Catatan
                              notes TEXT NULL,
    -- Timestamp
                              transaction_date TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                              created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                              updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    -- Constraints
                              CONSTRAINT transactions_type_check CHECK (transaction_type IN ('income', 'expenditure')),
                              CONSTRAINT transactions_payment_status_check CHECK (payment_status IN ('unpaid', 'dp', 'paid'))
);

-- ==================== INDEX (Opsional, untuk performa) ====================
CREATE INDEX idx_transactions_type ON transactions(transaction_type);
CREATE INDEX idx_transactions_date ON transactions(transaction_date);
CREATE INDEX idx_transactions_customer ON transactions(customer_id);
CREATE INDEX idx_transactions_user ON transactions(user_id);


-- ========================
-- TRANSACTION_ITEMS
-- ========================
CREATE TABLE public.transaction_items (
                                          id BIGSERIAL NOT NULL,
                                          transaction_id BIGINT NOT NULL,
                                          service_id BIGINT NULL,
                                          item_name varchar(255) NULL,
                                          qty numeric(10, 2) DEFAULT 1 NOT NULL,
                                          unit varchar(20) NULL,
                                          unit_price numeric(12, 2) DEFAULT 0 NOT NULL,
                                          subtotal numeric(12, 2) GENERATED ALWAYS AS ((qty * unit_price)) STORED NULL,
                                          notes text NULL,
                                          CONSTRAINT transaction_items_pkey PRIMARY KEY (id),
                                          CONSTRAINT transaction_items_service_id_fkey foreign key (service_id) references services(id),
                                          CONSTRAINT transaction_items_transaction_id_fkey foreign key (transaction_id) references transactions(id)
                                              on  delete cascade
);
-- ========================
-- TRANSACTION_PAYMENTS
-- ========================
CREATE TABLE public.transaction_payments (
                                             id bigserial NOT NULL,
                                             transaction_id BIGINT NOT NULL,
                                             payment_method_id BIGINT NOT NULL,
                                             amount numeric(12, 2) NOT NULL,
                                             payment_type varchar(20) DEFAULT 'dp' NOT NULL,
                                             paid_at timestamptz DEFAULT now() NOT NULL,
                                             received_by BIGINT NOT NULL,
                                             notes text NULL,
                                             created_at timestamptz DEFAULT now() NOT NULL,
                                             CONSTRAINT transaction_payments_amount_check CHECK ((amount > (0))),
                                             CONSTRAINT transaction_payments_payment_type_check CHECK (((payment_type) = ANY ((ARRAY['dp', 'settlement'])))),
                                             CONSTRAINT transaction_payments_pkey PRIMARY KEY (id)
);
-- ========================
-- TRANSACTION_DELIVERIES
-- ========================
CREATE TABLE public.transaction_deliveries (
                                               id              BIGSERIAL       NOT NULL,
                                               transaction_id  BIGINT          NOT NULL,
                                               td_type         VARCHAR(20)     NOT NULL DEFAULT 'income',
                                               status          VARCHAR(30)     NOT NULL DEFAULT 'pending',
                                               address         TEXT            NOT NULL,
                                               scheduled_at    TIMESTAMPTZ     NULL,
                                               completed_at    TIMESTAMPTZ     NULL,
                                               courier_name    VARCHAR(100)    NULL,
                                               delivery_fee    NUMERIC(12, 2)  NOT NULL DEFAULT 0,
                                               notes           TEXT            NULL,
                                               created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
                                               updated_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    -- Primary Key
                                               CONSTRAINT transaction_deliveries_pkey
                                                   PRIMARY KEY (id),
    -- Unique
                                               CONSTRAINT transaction_deliveries_transaction_id_type_key
                                                   UNIQUE (transaction_id, td_type),
    -- Check: td_type
                                               CONSTRAINT transaction_deliveries_type_check
                                                   CHECK (td_type = ANY (ARRAY['pickup', 'dropoff'])),
    -- Check: status
                                               CONSTRAINT transaction_deliveries_status_check
                                                   CHECK (status = ANY (ARRAY['pending', 'scheduled', 'on_the_way', 'done', 'cancelled']))
);
ALTER TABLE transactions
    ADD COLUMN IF NOT EXISTS is_deleted BOOLEAN DEFAULT false,
    ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ NULL;

-- Create index for soft delete
CREATE INDEX idx_transactions_is_deleted ON transactions(is_deleted);
CREATE INDEX idx_transactions_deleted_at ON transactions(deleted_at);
CREATE TABLE blacklisted_tokens (
    jti  TEXT PRIMARY KEY,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);