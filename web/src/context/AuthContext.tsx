import { createContext, useContext, useEffect, useState } from 'react'
import type { ReactNode } from 'react'
import { api, TOKEN_KEY } from '../api/client'
import type { AuthResponse, MeResponse, User } from '../api/types'

interface AuthContextValue {
  user: User | null
  businessName: string | null
  loading: boolean
  login: (phone: string, password: string) => Promise<void>
  register: (data: {
    business_name: string
    owner_name: string
    phone: string
    email: string
    password: string
  }) => Promise<void>
  logout: () => Promise<void>
}

const AuthContext = createContext<AuthContextValue | undefined>(undefined)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null)
  const [businessName, setBusinessName] = useState<string | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    const token = localStorage.getItem(TOKEN_KEY)
    if (!token) {
      setLoading(false)
      return
    }
    api
      .get<MeResponse>('/api/v1/auth/me')
      .then((res) => {
        setUser(res.data.user)
        setBusinessName(res.data.business.name)
      })
      .catch(() => {
        localStorage.removeItem(TOKEN_KEY)
      })
      .finally(() => setLoading(false))
  }, [])

  async function login(phone: string, password: string) {
    const res = await api.post<AuthResponse>('/api/v1/auth/login', { phone, password })
    localStorage.setItem(TOKEN_KEY, res.data.token)
    setUser(res.data.user)
    setBusinessName(res.data.business.name)
  }

  async function register(data: {
    business_name: string
    owner_name: string
    phone: string
    email: string
    password: string
  }) {
    const res = await api.post<AuthResponse>('/api/v1/auth/register', data)
    localStorage.setItem(TOKEN_KEY, res.data.token)
    setUser(res.data.user)
    setBusinessName(res.data.business.name)
  }

  async function logout() {
    try {
      await api.post('/api/v1/auth/logout')
    } catch {
      // ignore logout errors; clear local session regardless
    }
    localStorage.removeItem(TOKEN_KEY)
    setUser(null)
    setBusinessName(null)
  }

  return (
    <AuthContext.Provider value={{ user, businessName, loading, login, register, logout }}>
      {children}
    </AuthContext.Provider>
  )
}

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuth must be used within AuthProvider')
  return ctx
}