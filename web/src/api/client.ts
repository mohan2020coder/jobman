import axios, { AxiosError } from 'axios'
import type { ApiError } from './types'

export const API_BASE = import.meta.env.VITE_API_URL ?? 'http://localhost:8080'

export const TOKEN_KEY = 'jobman.token'

export const api = axios.create({
  baseURL: API_BASE,
  timeout: 20000,
})

api.interceptors.request.use((config) => {
  const token = localStorage.getItem(TOKEN_KEY)
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

api.interceptors.response.use(
  (res) => res,
  (error: AxiosError<ApiError>) => {
    if (error.response?.status === 401 && localStorage.getItem(TOKEN_KEY)) {
      localStorage.removeItem(TOKEN_KEY)
      if (!window.location.pathname.startsWith('/login')) {
        window.location.href = '/login'
      }
    }
    return Promise.reject(error)
  },
)

export function errorMessage(err: unknown): string {
  if (axios.isAxiosError<ApiError>(err)) {
    return err.response?.data?.error?.message ?? err.message ?? 'Something went wrong.'
  }
  if (err instanceof Error) return err.message
  return 'Something went wrong.'
}

export function publicReceiptUrl(publicUrl: string): string {
  return API_BASE + publicUrl.replace(/^\//, '/')
}