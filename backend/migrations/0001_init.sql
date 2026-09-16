-- 0001_init.sql
-- Initial schema for the job management SaaS.

CREATE TABLE businesses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(150) NOT NULL,
    phone VARCHAR(20),
    email VARCHAR(150),
    address TEXT,
    timezone VARCHAR(50) NOT NULL DEFAULT 'Asia/Kolkata',
    currency VARCHAR(3) NOT NULL DEFAULT 'INR',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    business_id UUID NOT NULL REFERENCES businesses(id),
    name VARCHAR(150) NOT NULL,
    phone VARCHAR(20) NOT NULL,
    email VARCHAR(150),
    password_hash TEXT NOT NULL,
    role VARCHAR(20) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    last_login_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT users_role_check CHECK (role IN ('OWNER', 'ADMIN', 'TECHNICIAN'))
);

CREATE UNIQUE INDEX users_business_phone_unique ON users(business_id, phone);

CREATE TABLE customers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    business_id UUID NOT NULL REFERENCES businesses(id),
    name VARCHAR(150) NOT NULL,
    phone VARCHAR(20) NOT NULL,
    email VARCHAR(150),
    address TEXT,
    latitude NUMERIC(10,7),
    longitude NUMERIC(10,7),
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX customers_business_idx ON customers(business_id);
CREATE INDEX customers_business_phone_idx ON customers(business_id, phone);

CREATE TABLE technicians (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    business_id UUID NOT NULL REFERENCES businesses(id),
    user_id UUID NOT NULL UNIQUE REFERENCES users(id),
    status VARCHAR(20) NOT NULL DEFAULT 'AVAILABLE',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT technician_status_check CHECK (status IN ('AVAILABLE', 'BUSY', 'INACTIVE'))
);

CREATE TABLE jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    business_id UUID NOT NULL REFERENCES businesses(id),
    customer_id UUID NOT NULL REFERENCES customers(id),
    technician_id UUID REFERENCES technicians(id),

    job_number VARCHAR(30) NOT NULL,

    service_type VARCHAR(150) NOT NULL,
    problem_description TEXT,

    address TEXT,
    latitude NUMERIC(10,7),
    longitude NUMERIC(10,7),

    status VARCHAR(30) NOT NULL DEFAULT 'PENDING',

    estimated_amount NUMERIC(12,2) NOT NULL DEFAULT 0,
    final_amount NUMERIC(12,2) NOT NULL DEFAULT 0,

    scheduled_at TIMESTAMPTZ,

    accepted_at TIMESTAMPTZ,
    on_the_way_at TIMESTAMPTZ,
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,

    notes TEXT,

    created_by UUID NOT NULL REFERENCES users(id),

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT jobs_status_check CHECK (
        status IN (
            'PENDING',
            'ACCEPTED',
            'ON_THE_WAY',
            'STARTED',
            'COMPLETED',
            'CANCELLED'
        )
    )
);

CREATE UNIQUE INDEX jobs_business_number_unique ON jobs(business_id, job_number);
CREATE INDEX jobs_business_status_idx ON jobs(business_id, status);
CREATE INDEX jobs_business_technician_idx ON jobs(business_id, technician_id);
CREATE INDEX jobs_business_customer_idx ON jobs(business_id, customer_id);
CREATE INDEX jobs_business_scheduled_idx ON jobs(business_id, scheduled_at);
CREATE INDEX jobs_business_created_idx ON jobs(business_id, created_at);

CREATE TABLE job_status_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    business_id UUID NOT NULL REFERENCES businesses(id),
    job_id UUID NOT NULL REFERENCES jobs(id) ON DELETE CASCADE,
    old_status VARCHAR(30),
    new_status VARCHAR(30) NOT NULL,
    changed_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX job_status_history_job_idx ON job_status_history(job_id);
CREATE INDEX job_status_history_business_idx ON job_status_history(business_id);

CREATE TABLE job_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    business_id UUID NOT NULL REFERENCES businesses(id),
    job_id UUID NOT NULL REFERENCES jobs(id) ON DELETE CASCADE,
    description VARCHAR(255) NOT NULL,
    quantity NUMERIC(10,2) NOT NULL DEFAULT 1,
    unit_price NUMERIC(12,2) NOT NULL DEFAULT 0,
    total_price NUMERIC(12,2) NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX job_items_job_idx ON job_items(job_id);

CREATE TABLE payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    business_id UUID NOT NULL REFERENCES businesses(id),
    job_id UUID NOT NULL REFERENCES jobs(id),
    amount NUMERIC(12,2) NOT NULL,
    method VARCHAR(30) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'PAID',
    transaction_reference VARCHAR(150),
    paid_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT payment_method_check CHECK (
        method IN ('CASH', 'UPI', 'CARD', 'BANK_TRANSFER', 'OTHER')
    ),
    CONSTRAINT payment_status_check CHECK (
        status IN ('PAID', 'REFUNDED')
    )
);

CREATE INDEX payments_business_idx ON payments(business_id);
CREATE INDEX payments_job_idx ON payments(job_id);
CREATE INDEX payments_paid_at_idx ON payments(business_id, paid_at);

CREATE TABLE receipts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    business_id UUID NOT NULL REFERENCES businesses(id),
    job_id UUID NOT NULL REFERENCES jobs(id),
    receipt_number VARCHAR(50) NOT NULL,
    public_token VARCHAR(100) NOT NULL UNIQUE,
    issued_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX receipts_business_number_unique ON receipts(business_id, receipt_number);

-- Sequences for human-friendly sequential numbers.
CREATE SEQUENCE job_number_seq START 1000;
CREATE SEQUENCE receipt_number_seq START 1000;