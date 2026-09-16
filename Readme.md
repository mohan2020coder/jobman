# Service Business Job Management SaaS — MVP

## 1. Project Goal

Build a production-ready MVP for a simple job-management SaaS for small service businesses.

Target businesses:

* AC repair/service
* Plumbing
* Electrical services
* RO service
* Appliance repair
* Washing machine repair
* Refrigerator repair
* CCTV installation/service
* Pest control
* Other local field-service businesses

The application allows a business owner to:

1. Manage customers
2. Manage technicians
3. Create service jobs
4. Assign jobs to technicians
5. Track job status
6. Track payments
7. Generate receipts
8. View today's jobs and collection

Technicians use a React Native mobile application to:

1. View assigned jobs
2. Accept jobs
3. Mark "On the way"
4. Start work
5. Add work notes
6. Add job items/charges
7. Record payment
8. Complete the job

Customers do NOT need an account or mobile application in MVP.

Customers receive a public receipt URL that can be shared through WhatsApp/SMS.

Do NOT use an LLM or AI API in this MVP.

---

# 2. Technology Stack

## Backend

Use:

* Go
* REST API
* PostgreSQL
* JWT authentication
* bcrypt/Argon2 password hashing
* database/sql or pgx
* migrations
* environment-based configuration

Recommended Go structure:

```text
backend/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── auth/
│   ├── business/
│   ├── users/
│   ├── customers/
│   ├── technicians/
│   ├── jobs/
│   ├── payments/
│   ├── receipts/
│   ├── dashboard/
│   └── notifications/
├── middleware/
├── database/
├── migrations/
├── pkg/
├── config/
└── go.mod
```

Use a modular monolith.

Do NOT create microservices.

## Owner Web Application

Use:

* React
* TypeScript
* Vite
* React Router
* TanStack Query
* Axios/fetch
* simple component library or Bootstrap/Tailwind

The owner needs a desktop/web dashboard.

## Technician Application

Use:

* React Native
* TypeScript
* React Navigation
* TanStack Query
* secure token storage

The technician application must be mobile-first and simple.

---

# 3. Multi-Tenant Architecture

This is a SaaS application.

Multiple businesses will use the same application/database.

Every business-owned record must be associated with:

```text
business_id
```

Example:

```text
Business A
 ├── Users
 ├── Technicians
 ├── Customers
 ├── Jobs
 └── Payments

Business B
 ├── Users
 ├── Technicians
 ├── Customers
 ├── Jobs
 └── Payments
```

NEVER allow one business to access another business's data.

Every repository query involving business-owned data must filter by:

```sql
WHERE business_id = $1
```

Do not rely only on frontend filtering.

Tenant isolation must be enforced in the backend.

---

# 4. User Roles

Initially support:

```text
OWNER
ADMIN
TECHNICIAN
```

Permissions:

## OWNER

Can:

* manage business
* manage users
* manage technicians
* manage customers
* create jobs
* assign jobs
* update jobs
* view payments
* view dashboard
* view reports

## ADMIN

Can:

* manage customers
* create jobs
* assign jobs
* view jobs
* view payments
* view dashboard

Cannot:

* delete business
* manage owner

## TECHNICIAN

Can:

* view assigned jobs
* accept assigned jobs
* update assigned job status
* add work notes
* add job items
* record payment
* complete assigned jobs

Technicians cannot access jobs belonging to another technician unless explicitly allowed by an owner/admin endpoint.

---

# 5. Database Design

Use PostgreSQL.

Use UUIDs for primary keys.

Use:

```sql
uuid
varchar
text
numeric(12,2)
boolean
timestamp with time zone
```

Use `created_at` and `updated_at` on major tables.

Use foreign keys.

Use indexes for:

* business_id
* customer phone
* job status
* technician_id
* scheduled_at
* created_at

---

# 6. businesses

```sql
CREATE TABLE businesses (
    id UUID PRIMARY KEY,
    name VARCHAR(150) NOT NULL,
    phone VARCHAR(20),
    email VARCHAR(150),
    address TEXT,
    timezone VARCHAR(50) NOT NULL DEFAULT 'Asia/Kolkata',
    currency VARCHAR(3) NOT NULL DEFAULT 'INR',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

---

# 7. users

Users belong to a business.

```sql
CREATE TABLE users (
    id UUID PRIMARY KEY,
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

    CONSTRAINT users_role_check
        CHECK (role IN ('OWNER', 'ADMIN', 'TECHNICIAN'))
);

CREATE UNIQUE INDEX users_business_phone_unique
ON users(business_id, phone);
```

---

# 8. customers

```sql
CREATE TABLE customers (
    id UUID PRIMARY KEY,
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

CREATE INDEX customers_business_idx
ON customers(business_id);

CREATE INDEX customers_business_phone_idx
ON customers(business_id, phone);
```

Latitude/longitude are optional in MVP.

Do not implement live GPS tracking.

---

# 9. technicians

A technician is a user with technician-specific information.

```sql
CREATE TABLE technicians (
    id UUID PRIMARY KEY,
    business_id UUID NOT NULL REFERENCES businesses(id),
    user_id UUID NOT NULL UNIQUE REFERENCES users(id),
    status VARCHAR(20) NOT NULL DEFAULT 'AVAILABLE',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT technician_status_check
        CHECK (status IN ('AVAILABLE', 'BUSY', 'INACTIVE'))
);
```

---

# 10. jobs

This is the main business table.

```sql
CREATE TABLE jobs (
    id UUID PRIMARY KEY,
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

    CONSTRAINT jobs_status_check
        CHECK (
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

CREATE UNIQUE INDEX jobs_business_number_unique
ON jobs(business_id, job_number);

CREATE INDEX jobs_business_status_idx
ON jobs(business_id, status);

CREATE INDEX jobs_business_technician_idx
ON jobs(business_id, technician_id);

CREATE INDEX jobs_business_customer_idx
ON jobs(business_id, customer_id);

CREATE INDEX jobs_business_scheduled_idx
ON jobs(business_id, scheduled_at);
```

---

# 11. job_status_history

Every status transition must be recorded.

```sql
CREATE TABLE job_status_history (
    id UUID PRIMARY KEY,
    business_id UUID NOT NULL REFERENCES businesses(id),
    job_id UUID NOT NULL REFERENCES jobs(id) ON DELETE CASCADE,

    old_status VARCHAR(30),
    new_status VARCHAR(30) NOT NULL,

    changed_by UUID NOT NULL REFERENCES users(id),

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX job_status_history_job_idx
ON job_status_history(job_id);

CREATE INDEX job_status_history_business_idx
ON job_status_history(business_id);
```

Example:

```text
PENDING
   ↓
ACCEPTED
   ↓
ON_THE_WAY
   ↓
STARTED
   ↓
COMPLETED
```

---

# 12. job_items

A job can contain multiple charges/items.

Example:

```text
Gas refill        ₹1,000
Service charge      ₹500
```

Database:

```sql
CREATE TABLE job_items (
    id UUID PRIMARY KEY,
    business_id UUID NOT NULL REFERENCES businesses(id),
    job_id UUID NOT NULL REFERENCES jobs(id) ON DELETE CASCADE,

    description VARCHAR(255) NOT NULL,
    quantity NUMERIC(10,2) NOT NULL DEFAULT 1,
    unit_price NUMERIC(12,2) NOT NULL DEFAULT 0,
    total_price NUMERIC(12,2) NOT NULL DEFAULT 0,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX job_items_job_idx
ON job_items(job_id);
```

The backend must calculate:

```text
total_price = quantity × unit_price
```

Do not blindly trust the client-provided total.

---

# 13. payments

Keep payments separate from job status.

```sql
CREATE TABLE payments (
    id UUID PRIMARY KEY,
    business_id UUID NOT NULL REFERENCES businesses(id),
    job_id UUID NOT NULL REFERENCES jobs(id),

    amount NUMERIC(12,2) NOT NULL,

    method VARCHAR(30) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'PAID',

    transaction_reference VARCHAR(150),

    paid_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    created_by UUID NOT NULL REFERENCES users(id),

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT payment_method_check
        CHECK (
            method IN (
                'CASH',
                'UPI',
                'CARD',
                'BANK_TRANSFER',
                'OTHER'
            )
        ),

    CONSTRAINT payment_status_check
        CHECK (
            status IN (
                'PAID',
                'REFUNDED'
            )
        )
);

CREATE INDEX payments_business_idx
ON payments(business_id);

CREATE INDEX payments_job_idx
ON payments(job_id);

CREATE INDEX payments_paid_at_idx
ON payments(business_id, paid_at);
```

Allow multiple payments for one job.

Example:

```text
Job total: ₹2,000

Payment 1: ₹1,000 CASH
Payment 2: ₹1,000 UPI
```

---

# 14. receipts

```sql
CREATE TABLE receipts (
    id UUID PRIMARY KEY,
    business_id UUID NOT NULL REFERENCES businesses(id),
    job_id UUID NOT NULL REFERENCES jobs(id),

    receipt_number VARCHAR(50) NOT NULL,
    public_token VARCHAR(100) NOT NULL UNIQUE,

    issued_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX receipts_business_number_unique
ON receipts(business_id, receipt_number);
```

Customer should access a receipt using:

```text
/r/{public_token}
```

Do not expose internal UUIDs in public receipt URLs.

---

# 15. Database relationships

```text
businesses
    │
    ├── users
    │     │
    │     └── technicians
    │
    ├── customers
    │
    └── jobs
          │
          ├── job_status_history
          │
          ├── job_items
          │
          ├── payments
          │
          └── receipts
```

---

# 16. Job State Machine

Implement the following transitions only.

```text
PENDING
   ↓
ACCEPTED
   ↓
ON_THE_WAY
   ↓
STARTED
   ↓
COMPLETED
```

Cancellation:

```text
PENDING → CANCELLED
ACCEPTED → CANCELLED
ON_THE_WAY → CANCELLED
```

Do not allow:

```text
PENDING → COMPLETED
PENDING → STARTED
COMPLETED → STARTED
COMPLETED → PENDING
```

Status transitions must be validated server-side.

---

# 17. Important Job Rules

When assigning a technician:

```text
PENDING
+
technician_id
```

When technician accepts:

```text
status = ACCEPTED
accepted_at = NOW()
```

When technician marks on the way:

```text
status = ON_THE_WAY
on_the_way_at = NOW()
```

When technician starts:

```text
status = STARTED
started_at = NOW()
```

When completing:

```text
status = COMPLETED
completed_at = NOW()
```

Every status change must create a `job_status_history` record.

Use a database transaction for:

```text
update job
+
insert status history
```

---

# 18. Payment Rules

For a job:

```text
total_paid = SUM(payments.amount WHERE status = PAID)
```

Payment status should be calculated as:

```text
total_paid = 0
    → UNPAID

total_paid < final_amount
    → PARTIAL

total_paid >= final_amount
    → PAID
```

Do not store unnecessary duplicated payment status in the jobs table.

---

# 19. API Design

Base URL:

```text
/api/v1
```

All authenticated APIs require:

```http
Authorization: Bearer <token>
```

---

# 20. Authentication APIs

### Register business

```http
POST /api/v1/auth/register
```

Request:

```json
{
    "business_name": "ABC AC Services",
    "owner_name": "Rajesh",
    "phone": "9876543210",
    "email": "owner@example.com",
    "password": "password"
}
```

Response:

```json
{
    "token": "...",
    "user": {
        "id": "...",
        "name": "Rajesh",
        "role": "OWNER"
    },
    "business": {
        "id": "...",
        "name": "ABC AC Services"
    }
}
```

---

### Login

```http
POST /api/v1/auth/login
```

Request:

```json
{
    "phone": "9876543210",
    "password": "password"
}
```

Response:

```json
{
    "token": "...",
    "user": {
        "id": "...",
        "name": "Rajesh",
        "role": "OWNER"
    },
    "business": {
        "id": "...",
        "name": "ABC AC Services"
    }
}
```

---

### Current user

```http
GET /api/v1/auth/me
```

---

### Logout

```http
POST /api/v1/auth/logout
```

If using stateless JWT, implement token expiration appropriately.

Do not create unnecessary complexity for MVP.

---

# 21. Customer APIs

```http
GET /api/v1/customers
POST /api/v1/customers
GET /api/v1/customers/:id
PUT /api/v1/customers/:id
DELETE /api/v1/customers/:id
GET /api/v1/customers/:id/jobs
```

Customer create:

```json
{
    "name": "Ravi Kumar",
    "phone": "9876543210",
    "email": "ravi@example.com",
    "address": "12 MG Road",
    "notes": "Usually available after 6 PM"
}
```

Customer list must support:

```text
search
page
limit
```

Search by:

```text
name
phone
```

---

# 22. Technician APIs

```http
GET /api/v1/technicians
POST /api/v1/technicians
GET /api/v1/technicians/:id
PUT /api/v1/technicians/:id
DELETE /api/v1/technicians/:id
GET /api/v1/technicians/:id/jobs
```

Technician creation:

```json
{
    "name": "Kumar",
    "phone": "9876543211",
    "password": "temporary-password"
}
```

Only OWNER/ADMIN can create technicians.

---

# 23. Job APIs

```http
GET /api/v1/jobs
POST /api/v1/jobs
GET /api/v1/jobs/:id
PUT /api/v1/jobs/:id
DELETE /api/v1/jobs/:id
```

Job creation:

```json
{
    "customer_id": "...",
    "technician_id": "...",
    "service_type": "AC Repair",
    "problem_description": "AC not cooling",
    "address": "12 MG Road",
    "estimated_amount": 1500,
    "scheduled_at": "2026-09-15T14:00:00+05:30"
}
```

Backend generates:

```text
job_number
```

Do not allow the client to choose job numbers.

---

# 24. Job Action APIs

Do NOT allow arbitrary status updates from technicians.

Use dedicated endpoints.

```http
POST /api/v1/jobs/:id/accept
POST /api/v1/jobs/:id/on-the-way
POST /api/v1/jobs/:id/start
POST /api/v1/jobs/:id/complete
POST /api/v1/jobs/:id/cancel
```

---

# 25. Accept Job

```http
POST /api/v1/jobs/:id/accept
```

Only the assigned technician can accept the job.

Validate:

```text
current status = PENDING
current technician = authenticated technician
```

---

# 26. On The Way

```http
POST /api/v1/jobs/:id/on-the-way
```

Validate:

```text
current status = ACCEPTED
current technician = authenticated technician
```

---

# 27. Start Job

```http
POST /api/v1/jobs/:id/start
```

Validate:

```text
current status = ON_THE_WAY
current technician = authenticated technician
```

---

# 28. Complete Job

```http
POST /api/v1/jobs/:id/complete
```

Request:

```json
{
    "notes": "Gas refill completed",
    "items": [
        {
            "description": "Gas refill",
            "quantity": 1,
            "unit_price": 1000
        },
        {
            "description": "Service charge",
            "quantity": 1,
            "unit_price": 500
        }
    ],
    "payment": {
        "amount": 1500,
        "method": "UPI",
        "transaction_reference": "UPI123456"
    }
}
```

Backend must calculate:

```text
1000 + 500 = 1500
```

Do not trust client-calculated totals.

The complete operation should be transactional:

```text
BEGIN

insert job items

calculate final amount

update job

insert payment

insert job status history

create receipt

COMMIT
```

---

# 29. Payment APIs

```http
GET /api/v1/jobs/:id/payments
POST /api/v1/jobs/:id/payments
```

Payment request:

```json
{
    "amount": 500,
    "method": "CASH",
    "transaction_reference": null
}
```

Backend must validate:

```text
amount > 0
amount <= remaining balance
```

unless an admin explicitly allows overpayment.

---

# 30. Receipt API

Authenticated:

```http
GET /api/v1/jobs/:id/receipt
```

Public:

```http
GET /api/v1/public/receipts/:token
```

Public receipt must not require authentication.

It should expose only safe information:

```json
{
    "business": {
        "name": "ABC AC Services",
        "phone": "9876543210"
    },
    "receipt_number": "RCP-000124",
    "customer_name": "Ravi Kumar",
    "service": "AC Repair",
    "items": [
        {
            "description": "Gas refill",
            "amount": 1000
        },
        {
            "description": "Service charge",
            "amount": 500
        }
    ],
    "total": 1500,
    "paid": 1500,
    "payment_status": "PAID",
    "payment_method": "UPI",
    "completed_at": "..."
}
```

Never expose:

```text
password
password_hash
internal user IDs
business database information
private notes
JWT tokens
```

---

# 31. Dashboard APIs

```http
GET /api/v1/dashboard/today
```

Response:

```json
{
    "date": "2026-09-15",
    "jobs": {
        "total": 18,
        "pending": 5,
        "accepted": 1,
        "on_the_way": 1,
        "started": 3,
        "completed": 13,
        "cancelled": 0
    },
    "collection": {
        "total": 24500,
        "cash": 8500,
        "upi": 14000,
        "card": 2000,
        "bank_transfer": 0
    }
}
```

---

# 32. Job List API

```http
GET /api/v1/jobs
```

Support:

```text
status
technician_id
customer_id
date
search
page
limit
```

Example:

```text
GET /api/v1/jobs?status=COMPLETED&date=2026-09-15
```

Response:

```json
{
    "data": [],
    "pagination": {
        "page": 1,
        "limit": 20,
        "total": 100
    }
}
```

---

# 33. Technician "My Jobs"

```http
GET /api/v1/technician/jobs
```

Return only jobs belonging to the authenticated technician.

Support:

```text
today
upcoming
completed
```

Example:

```text
GET /api/v1/technician/jobs?date=today
```

---

# 34. API Error Format

Use a consistent format:

```json
{
    "error": {
        "code": "JOB_INVALID_STATUS",
        "message": "Job cannot be started from its current status."
    }
}
```

Examples:

```text
AUTH_INVALID_CREDENTIALS
AUTH_UNAUTHORIZED
VALIDATION_ERROR
RESOURCE_NOT_FOUND
FORBIDDEN
JOB_INVALID_STATUS
JOB_NOT_ASSIGNED
PAYMENT_EXCEEDS_BALANCE
BUSINESS_ACCESS_DENIED
```

Use correct HTTP status codes:

```text
400 validation
401 authentication
403 authorization
404 not found
409 conflict
422 business validation
500 internal error
```

---

# 35. Security Requirements

Implement:

* password hashing
* JWT authentication
* authentication middleware
* role-based authorization
* business/tenant isolation
* request validation
* SQL parameterization
* no SQL string concatenation
* rate limiting on login
* secure CORS
* reasonable request body limits
* structured logging
* never log passwords or tokens

Technician authorization is critical.

A technician must not be able to modify another technician's job by changing a job ID.

Every job action must verify:

```text
job.business_id == authenticated_user.business_id
```

and for technician actions:

```text
job.technician_id == authenticated_technician.id
```

---

# 36. Frontend Owner Dashboard

Create these pages:

```text
/login

/dashboard

/jobs
/jobs/new
/jobs/:id

/customers
/customers/:id

/technicians
/technicians/:id

/payments
```

Dashboard:

```text
Today's Jobs
Completed
Pending
Collection

Today's Job List
```

---

# 37. React Native Technician App

Screens:

```text
Login
Home
My Jobs
Job Details
Work Details
Payment
Profile
```

Home:

```text
Good morning Kumar

Today's Jobs: 5

Pending: 2
In Progress: 1
Completed: 2
```

Job card:

```text
Ravi Kumar

AC Repair

AC not cooling

12 MG Road

[View Job]
```

Job details:

```text
Ravi Kumar
9876543210

AC Repair

AC not cooling

12 MG Road

[Call]
[Open Maps]

Status: Pending

[Accept]
```

After accepting:

```text
[On the way]
```

After on-the-way:

```text
[Start Job]
```

After starting:

```text
Work Notes

Items

Payment

[Complete Job]
```

---

# 38. Maps

Do not implement live GPS tracking.

For MVP, only provide:

```text
Open in Google Maps
```

using the customer address or coordinates.

The mobile app can launch the external map application.

---

# 39. Notifications

Design the backend so notifications can be added.

For MVP, create a notification abstraction:

```go
type NotificationService interface {
    NotifyJobAssigned(...)
    NotifyJobStatusChanged(...)
    NotifyJobCompleted(...)
}
```

Do not tightly couple the job service to Firebase.

If Firebase Cloud Messaging is implemented, use it for:

```text
New job assigned
Job reassigned
```

Keep notification implementation simple.

---

# 40. Receipt Generation

For MVP:

1. Create receipt record.
2. Generate public token.
3. Provide public receipt URL.
4. Owner/technician can share the URL.

Do not initially build complicated PDF generation.

A responsive HTML receipt page is sufficient.

Later add:

```text
Download PDF
WhatsApp API
SMS
Email
```

---

# 41. Important UX principle

This application is for small business owners.

Keep it extremely simple.

Avoid:

* complex CRM screens
* excessive configuration
* complicated dashboards
* unnecessary forms
* too many statuses
* AI features
* enterprise features

The owner should be able to create a job in less than 30 seconds.

The technician should be able to update a job with one or two taps.

---

# 42. MVP Scope

Implement ONLY:

### Authentication

* register
* login
* logout
* current user

### Business

* business creation during registration
* business profile

### Customers

* create
* edit
* list
* search
* history

### Technicians

* create
* edit
* activate/deactivate
* list
* jobs

### Jobs

* create
* assign
* list
* details
* accept
* on the way
* start
* complete
* cancel
* status history

### Payments

* add payment
* payment history
* payment status
* daily collection

### Receipts

* receipt
* public receipt URL

### Dashboard

* today's jobs
* job status counts
* today's collection

---

# 43. Explicitly DO NOT implement yet

Do NOT implement:

* AI/LLM
* chatbot
* voice assistant
* OCR
* automatic scheduling
* route optimization
* live GPS tracking
* customer mobile app
* inventory management
* GST accounting
* payroll
* full accounting
* CRM automation
* subscription billing
* marketplace
* advanced analytics
* microservices
* event-driven architecture
* Kubernetes
* complex infrastructure

Keep the system simple.

---

# 44. Recommended Backend Layering

Use:

```text
HTTP Handler
      ↓
Service
      ↓
Repository
      ↓
PostgreSQL
```

Example:

```text
jobs/
├── handler.go
├── service.go
├── repository.go
├── model.go
├── dto.go
└── validation.go
```

Handler should handle:

```text
HTTP
JSON
authentication context
validation errors
HTTP status codes
```

Service should handle:

```text
business logic
state transitions
authorization rules
transactions
```

Repository should handle:

```text
SQL
database operations
```

Do not put business logic directly in handlers.

---

# 45. Transactions

The following operations must use database transactions:

### Accept job

```text
update job
+
insert status history
```

### Start job

```text
update job
+
insert status history
```

### Complete job

```text
insert items
+
calculate total
+
update job
+
insert payment
+
insert receipt
+
insert status history
```

If any operation fails, rollback the entire transaction.

---

# 46. Important Database Queries

Dashboard query should efficiently calculate today's jobs.

Conceptually:

```sql
SELECT
    COUNT(*) AS total,
    COUNT(*) FILTER (WHERE status = 'PENDING') AS pending,
    COUNT(*) FILTER (WHERE status = 'ACCEPTED') AS accepted,
    COUNT(*) FILTER (WHERE status = 'ON_THE_WAY') AS on_the_way,
    COUNT(*) FILTER (WHERE status = 'STARTED') AS started,
    COUNT(*) FILTER (WHERE status = 'COMPLETED') AS completed,
    COUNT(*) FILTER (WHERE status = 'CANCELLED') AS cancelled
FROM jobs
WHERE business_id = $1
AND created_at >= $2
AND created_at < $3;
```

Collection:

```sql
SELECT
    COALESCE(SUM(amount), 0) AS total
FROM payments
WHERE business_id = $1
AND status = 'PAID'
AND paid_at >= $2
AND paid_at < $3;
```

Payment method breakdown:

```sql
SELECT
    method,
    COALESCE(SUM(amount), 0) AS total
FROM payments
WHERE business_id = $1
AND status = 'PAID'
AND paid_at >= $2
AND paid_at < $3
GROUP BY method;
```

---

# 47. Seed Data

Create development seed data:

```text
Business:
ABC AC Services

Owner:
Rajesh

Technicians:
Kumar
Manoj
Ravi

Customers:
Ravi Kumar
Suresh
Priya
Ganesh

Jobs:
Several jobs in different statuses
```

Include example payments.

This allows the dashboard and mobile app to be tested immediately.

---

# 48. Testing

Write tests for the most important business rules.

At minimum:

### Authentication

* valid login
* invalid login
* inactive user

### Tenant isolation

Business A cannot access Business B customers/jobs/payments.

### Job state machine

Test:

```text
PENDING → ACCEPTED       valid
ACCEPTED → ON_THE_WAY   valid
ON_THE_WAY → STARTED    valid
STARTED → COMPLETED     valid
PENDING → COMPLETED     invalid
COMPLETED → STARTED     invalid
```

### Technician authorization

Technician A cannot modify Technician B's job.

### Payments

Test:

```text
₹1,500 job
₹500 payment → PARTIAL
₹1,000 payment → PAID
```

### Completion

Verify that completion transaction creates:

```text
job items
payment
receipt
status history
```

If one fails, nothing should be committed.

---

# 49. API Documentation

Create an OpenAPI specification:

```text
openapi.yaml
```

Document:

* authentication
* request bodies
* response bodies
* error responses
* authorization
* pagination
* filters

Keep the OpenAPI specification synchronized with the implementation.

---

# 50. Environment Configuration

Create:

```text
.env.example
```

Example:

```env
APP_ENV=development
APP_PORT=8080

DATABASE_URL=postgres://postgres:postgres@localhost:5432/service_app

JWT_SECRET=change-me
JWT_EXPIRATION=24h

CORS_ORIGINS=http://localhost:5173
```

Never commit real secrets.

---

# 51. Docker Development Environment

Create:

```text
docker-compose.yml
```

Initially only PostgreSQL is required.

Example services:

```text
postgres
```

The Go API can run locally during development.

---

# 52. Project Deliverables

Build the project in this order:

## Step 1

Create project structure.

## Step 2

Create PostgreSQL migrations.

## Step 3

Implement database connection.

## Step 4

Implement authentication.

## Step 5

Implement tenant middleware.

## Step 6

Implement customer module.

## Step 7

Implement technician module.

## Step 8

Implement job module and state machine.

## Step 9

Implement payments.

## Step 10

Implement receipts.

## Step 11

Implement dashboard.

## Step 12

Implement React owner dashboard.

## Step 13

Implement React Native technician app.

## Step 14

Implement notifications abstraction.

## Step 15

Add tests.

## Step 16

Add OpenAPI documentation.

---

# 53. Coding Rules

Write clean, maintainable production-oriented code.

Prefer simple solutions.

Do not over-engineer.

Do not introduce dependencies unless they provide clear value.

Use context.Context throughout backend requests.

Use parameterized SQL.

Validate all input.

Use transactions where specified.

Return consistent API responses.

Do not expose internal errors to users.

Log useful server-side errors.

Never log:

```text
password
JWT
authorization header
private tokens
```

---

# 54. Development Approach

Do NOT generate the entire application blindly in one step.

Implement one module at a time.

After each major module:

1. Build the project.
2. Run tests.
3. Fix compilation errors.
4. Verify database migrations.
5. Verify API behavior.
6. Then proceed to the next module.

Start with:

```text
PostgreSQL
+
Go backend
+
authentication
+
customers
+
technicians
+
jobs
```

Only after the backend APIs work correctly should the frontend/mobile UI be built.

---

# 55. First Implementation Target

The first working vertical slice should be:

```text
Owner registers
      ↓
Owner logs in
      ↓
Owner creates technician
      ↓
Owner creates customer
      ↓
Owner creates job
      ↓
Owner assigns technician
      ↓
Technician logs in
      ↓
Technician sees job
      ↓
Technician accepts
      ↓
On the way
      ↓
Starts
      ↓
Completes
      ↓
Adds ₹1,500 payment
      ↓
Receipt generated
      ↓
Owner dashboard shows collection
```

This complete flow is more important than building many disconnected features.

---

# 56. Final Product Principle

The application should answer these questions for a small service business:

```text
What jobs do I have today?

Who is handling each job?

What is the current status?

Which jobs are completed?

How much money did I collect?

Which customers did we service?

What did we do for this customer previously?
```

If the application can answer those questions quickly, the MVP is successful.

Build for simplicity first.
Do not add AI unless a real customer problem later justifies it.

---

# 57. Running the MVP

## Prerequisites

* Go 1.25+
* Node 20+ / npm
* Docker (for PostgreSQL)

## 1. Database

```bash
docker compose up -d
```

PostgreSQL is exposed on host port `5433` (see `docker-compose.yml`). Create
`backend\.env` from `backend\.env.example`; it already points at port `5433`.

## 2. API server

```bash
cd backend
go run ./cmd/server
```

Serves the REST API on `http://localhost:8080` (base path `/api/v1`). On first
start it applies migrations and seeds demo data. Seed users are created with the
password `password123` (see `main.go` → `SeedPasswords`).

Demo accounts:

| Role        | Phone       | Password    |
| ----------- | ----------- | ----------- |
| Owner       | 9876543210  | password123 |
| Technician  | 9876543211  | password123 |
|             | 9876543212  | password123 |
|             | 9876543213  | password123 |

## 3. Owner web dashboard

```bash
cd web
npm install
npm run dev
```

Open `http://localhost:5173`. Sign in as the owner demo account. Build check:
`npm run build` (also runs `tsc -b`).

## 4. Technician mobile app (Expo)

```bash
cd mobile
npm install
```

The mobile app is a React Native (Expo SDK 57) TypeScript app.

* Android emulator (default): the API base is already `http://10.0.2.2:8080`.
* Real device / iOS simulator: point the app at your machine's LAN IP.

```bash
# Windows PowerShell
$env:EXPO_PUBLIC_API_URL = "http://<YOUR-LAN-IP>:8080"
npm start
```

Then press `a` (Android) or `i` (iOS). Typecheck: `npm run typecheck`.

The CORS origin is configured for `http://localhost:5173`; native mobile
requests are not subject to CORS.

## 5. First vertical slice

1. Sign in to the web dashboard as the owner.
2. Create a technician, a customer, and a job assigned to the technician.
3. Sign in to the mobile app as that technician (phone + password).
4. Accept → On the way → Start → complete the job with items and a payment.
5. Back in the dashboard, the job shows COMPLETED and today's collection updates.
6. Share the public receipt URL (owner job page / `GET /api/v1/jobs/:id/receipt`).
