import { createContext, useContext, useEffect, useState } from 'react'
import type { ReactNode } from 'react'
import * as SecureStore from 'expo-secure-store'
import {
  loginRequest,
  meRequest,
  registerUnauthorizedHandler,
  setAuthToken,
} from '../api/client'
import type { User } from '../api/types'

const TOKEN_KEY = 'jobman.tech_token'

interface AuthContextValue {
  user: User | null
  businessName: string | null
  loading: boolean
  signIn: (phone: string, password: string) => Promise<void>
  signOut: () => Promise<void>
}

const AuthContext = createContext<AuthContextValue | undefined>(undefined)

async function persistToken(token: string | null) {
  if (token) {
    await SecureStore.setItemAsync(TOKEN_KEY, token)
  } else {
    await SecureStore.deleteItemAsync(TOKEN_KEY)
  }
  setAuthToken(token)
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null)
  const [businessName, setBusinessName] = useState<string | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    registerUnauthorizedHandler(() => {
      setUser(null)
      setBusinessName(null)
    })

    ;(async () => {
      try {
        const token = await SecureStore.getItemAsync(TOKEN_KEY)
        if (token) {
          setAuthToken(token)
          const me = await meRequest()
          setUser(me.user)
          setBusinessName(me.business.name)
        }
      } catch {
        await persistToken(null)
      } finally {
        setLoading(false)
      }
    })()
  }, [])

  async function signIn(phone: string, password: string) {
    const res = await loginRequest(phone, password)
    await persistToken(res.token)
    setUser(res.user)
    setBusinessName(res.business.name)
  }

  async function signOut() {
    try {
      await apiLogout()
    } catch {
      // ignore
    }
    await persistToken(null)
    setUser(null)
    setBusinessName(null)
  }

  return (
    <AuthContext.Provider value={{ user, businessName, loading, signIn, signOut }}>
      {children}
    </AuthContext.Provider>
  )
}

function apiLogout() {
  // local logout is enough for MVP; token is discarded client-side
  return Promise.resolve()
}

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuth must be used within AuthProvider')
  return ctx
}