import axios, { AxiosError } from 'axios'
import type { ApiError, AuthResponse, JobsResponse, Job, MeResponse, Payment } from './types'

// For a device, point this at your machine's LAN IP, e.g.
// EXPO_PUBLIC_API_URL=http://192.168.1.10:8080
export const API_BASE = process.env.EXPO_PUBLIC_API_URL ?? 'http://10.0.2.2:8080'

export const WS_URL = API_BASE.replace(/^http/, 'ws')

let authToken: string | null = null
let onUnauthorized: (() => void) | null = null

export function setAuthToken(token: string | null) {
  authToken = token
}

export function getAuthToken(): string | null {
  return authToken
}

export function registerUnauthorizedHandler(fn: () => void) {
  onUnauthorized = fn
}

export const api = axios.create({
  baseURL: API_BASE,
  timeout: 20000,
})

api.interceptors.request.use((config) => {
  if (authToken) {
    config.headers.Authorization = `Bearer ${authToken}`
  }
  return config
})

api.interceptors.response.use(
  (res) => res,
  (error: AxiosError<ApiError>) => {
    if (error.response?.status === 401 && authToken) {
      setAuthToken(null)
      onUnauthorized?.()
    }
    return Promise.reject(error)
  },
)

export function apiErrorMessage(err: unknown): string {
  if (axios.isAxiosError<ApiError>(err)) {
    return err.response?.data?.error?.message ?? err.message ?? 'Something went wrong.'
  }
  if (err instanceof Error) return err.message
  return 'Something went wrong.'
}

// ---- API helpers used by screens ----

export async function loginRequest(phone: string, password: string): Promise<AuthResponse> {
  const res = await api.post<AuthResponse>('/api/v1/auth/login', { phone, password })
  return res.data
}

export async function meRequest() {
  const res = await api.get<MeResponse>('/api/v1/auth/me')
  return res.data
}

export async function myJobsRequest(filter: '' | 'today' | 'upcoming' | 'completed' = '') {
  const res = await api.get<JobsResponse>('/api/v1/technician/jobs', { params: { filter } })
  return res.data.data
}

export async function jobRequest(id: string): Promise<Job> {
  const res = await api.get<Job>(`/api/v1/jobs/${id}`)
  return res.data
}

export async function jobAccept(id: string): Promise<Job> {
  const res = await api.post<Job>(`/api/v1/jobs/${id}/accept`)
  return res.data
}

export async function jobOnTheWay(id: string): Promise<Job> {
  const res = await api.post<Job>(`/api/v1/jobs/${id}/on-the-way`)
  return res.data
}

export async function jobStart(id: string): Promise<Job> {
  const res = await api.post<Job>(`/api/v1/jobs/${id}/start`)
  return res.data
}

export interface CompleteInput {
  notes?: string
  items: { description: string; quantity: number; unit_price: number }[]
  payment?: { amount: number; method: string; transaction_reference?: string } | null
}

export async function jobComplete(id: string, input: CompleteInput): Promise<Job> {
  const res = await api.post<Job>(`/api/v1/jobs/${id}/complete`, input)
  return res.data
}

export async function listPayments(id: string): Promise<Payment[]> {
  const res = await api.get<{ data: Payment[] }>(`/api/v1/jobs/${id}/payments`)
  return res.data.data
}

export async function addPayment(id: string, input: { amount: number; method: string; transaction_reference?: string }) {
  const res = await api.post<Payment>(`/api/v1/jobs/${id}/payments`, input)
  return res.data
}

export async function receiptUrlRequest(id: string): Promise<string> {
  const res = await api.get<{ public_url: string }>(`/api/v1/jobs/${id}/receipt`)
  return res.data.public_url
}