export interface User {
  id: string
  name: string
  role: 'OWNER' | 'ADMIN' | 'TECHNICIAN'
  business_id: string
}

export interface Business {
  id: string
  name: string
}

export interface AuthResponse {
  token: string
  user: User
  business: Business
}

export interface MeResponse {
  user: User
  business: Business
}

export interface Customer {
  id: string
  business_id: string
  name: string
  phone: string
  email: string | null
  address: string | null
  latitude: number | null
  longitude: number | null
  notes: string | null
  created_at: string
  updated_at: string
}

export interface TechnicianLite {
  id: string
  name: string
  phone: string
  status: string
}

export interface JobItem {
  id: string
  business_id: string
  job_id: string
  description: string
  quantity: number
  unit_price: number
  total_price: number
  created_at: string
}

export interface Payment {
  id: string
  business_id: string
  job_id: string
  amount: number
  method: 'CASH' | 'UPI' | 'CARD' | 'BANK_TRANSFER' | 'OTHER'
  status: 'PAID' | 'REFUNDED'
  transaction_reference: string | null
  paid_at: string
  created_by: string
  created_at: string
}

export type JobStatus =
  | 'PENDING'
  | 'ACCEPTED'
  | 'ON_THE_WAY'
  | 'STARTED'
  | 'COMPLETED'
  | 'CANCELLED'

export type PaymentStatus = 'UNPAID' | 'PARTIAL' | 'PAID'

export interface StatusHistory {
  id: string
  old_status: string | null
  new_status: string
  changed_by: string
  created_at: string
}

export interface Job {
  id: string
  business_id: string
  customer_id: string
  technician_id: string | null
  job_number: string
  service_type: string
  problem_description: string | null
  address: string | null
  latitude: number | null
  longitude: number | null
  status: JobStatus
  estimated_amount: number
  final_amount: number
  scheduled_at: string | null
  accepted_at: string | null
  on_the_way_at: string | null
  started_at: string | null
  completed_at: string | null
  notes: string | null
  created_by: string
  created_at: string
  updated_at: string
  customer?: Customer
  technician?: TechnicianLite
  items?: JobItem[]
  payments?: Payment[]
  status_history?: StatusHistory[]
  total_paid?: number
  payment_status?: PaymentStatus
  receipt?: { id: string; receipt_number: string; issued_at: string } | null
}

export interface JobsResponse {
  data: Job[]
}

export interface JobDetailResponse extends Job {}

export interface ApiError {
  error: {
    code: string
    message: string
  }
}