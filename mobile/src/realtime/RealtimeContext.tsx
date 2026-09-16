import { Ionicons } from '@expo/vector-icons'
import { useQueryClient } from '@tanstack/react-query'
import { useEffect, useRef, useState } from 'react'
import type { ReactNode } from 'react'
import { AppState, Pressable, StyleSheet, Text, View } from 'react-native'
import { useSafeAreaInsets } from 'react-native-safe-area-context'
import { getAuthToken, WS_URL } from '../api/client'
import { useAuth } from '../auth/AuthContext'
import { navigationRef } from '../navigation/ref'
import { colors, radius, shadows, spacing } from '../theme'

export interface SocketJob {
  id: string
  job_number: string
  service_type: string
  customer_name?: string
  status: string
  estimated_amount?: number
  scheduled_at?: string | null
  address?: string | null
}

export interface SocketMessage {
  type: 'job_assigned' | 'job_status_changed' | 'job_completed' | 'reconnected'
  ts?: number
  job?: SocketJob
}

type Listener = (msg: SocketMessage) => void
type Timer = ReturnType<typeof setTimeout>

let ws: WebSocket | null = null
let listeners = new Set<Listener>()
let reconnectTimer: Timer | null = null
let currentToken: string | null = null
let everConnected = false
let closing = false
let attempts = 0

function emit(msg: SocketMessage) {
  listeners.forEach((l) => {
    try {
      l(msg)
    } catch {
      // listener errors are isolated
    }
  })
}

function scheduleReconnect() {
  if (!currentToken || closing) return
  attempts += 1
  const delay = Math.min(30_000, Math.max(1_000, 1_000 * 2 ** (attempts - 1)))
  reconnectTimer = setTimeout(startSocket, delay)
}

function startSocket() {
  if (!currentToken || closing) return
  if (ws && (ws.readyState === WebSocket.OPEN || ws.readyState === WebSocket.CONNECTING)) return
  ws = new WebSocket(`${WS_URL}/api/v1/ws?token=${encodeURIComponent(currentToken)}`)
  ws.onopen = () => {
    attempts = 0
    if (everConnected) emit({ type: 'reconnected' })
    everConnected = true
  }
  ws.onmessage = (e) => {
    try {
      const m = JSON.parse(e.data as string) as SocketMessage
      if (m && m.type && m.job) emit(m)
    } catch {
      // ignore non-JSON frames
    }
  }
  ws.onerror = () => {
    ws?.close()
  }
  ws.onclose = () => {
    ws = null
    if (currentToken && !closing) scheduleReconnect()
  }
}

function start(token: string) {
  closing = false
  currentToken = token
  if (reconnectTimer) {
    clearTimeout(reconnectTimer)
    reconnectTimer = null
  }
  startSocket()
}

function stop() {
  closing = true
  currentToken = null
  if (reconnectTimer) {
    clearTimeout(reconnectTimer)
    reconnectTimer = null
  }
  if (ws) {
    ws.onclose = null
    ws.close()
    ws = null
  }
  everConnected = false
}

function ensureConnected() {
  if (!currentToken || closing) return
  if (!ws || ws.readyState === WebSocket.CLOSED) startSocket()
}

/**
 * Connects the authenticated technician to the live job stream, keeps the
 * job caches fresh, and surfaces new-job assignments as an in-app banner.
 */
export function RealtimeProvider({ children }: { children: ReactNode }) {
  const { user } = useAuth()
  const qc = useQueryClient()
  const insets = useSafeAreaInsets()
  const [notice, setNotice] = useState<SocketMessage | null>(null)
  const hideTimer = useRef<Timer | null>(null)

  useEffect(() => {
    if (!user) {
      stop()
      return
    }
    const token = getAuthToken()
    if (!token) return
    start(token)

    const unsubscribe = subscribe((msg) => {
      if (msg.type === 'reconnected') {
        qc.invalidateQueries({ queryKey: ['techjobs'] })
        return
      }
      if (!msg.job || !msg.job.id) return
      const jobId = msg.job.id
      if (msg.type === 'job_assigned') {
        qc.invalidateQueries({ queryKey: ['techjobs'] })
        showNotice(msg)
      } else if (msg.type === 'job_status_changed' || msg.type === 'job_completed') {
        qc.invalidateQueries({ queryKey: ['techjobs'] })
        qc.invalidateQueries({ queryKey: ['job', jobId] })
        qc.invalidateQueries({ queryKey: ['job', jobId, 'payments'] })
        qc.invalidateQueries({ queryKey: ['job', jobId, 'receipt'] })
      }
    })

    const sub = AppState.addEventListener('change', (state) => {
      if (state === 'active') ensureConnected()
    })

    return () => {
      unsubscribe()
      sub.remove()
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [user?.id])

  useEffect(() => () => stop(), [])

  function showNotice(msg: SocketMessage) {
    setNotice(msg)
    if (hideTimer.current) clearTimeout(hideTimer.current)
    hideTimer.current = setTimeout(() => setNotice(null), 12_000)
  }

  function hideNotice() {
    if (hideTimer.current) clearTimeout(hideTimer.current)
    setNotice(null)
  }

  function openJob() {
    const jobId = notice?.job?.id
    hideNotice()
    if (jobId && navigationRef.isReady()) {
      navigationRef.navigate('JobDetail', { id: jobId })
    }
  }

  const job = notice?.job

  return (
    <>
      {children}
      {notice && job ? (
        <Pressable
          style={[styles.toast, { top: insets.top + 8 }]}
          onPress={openJob}
          accessibilityRole="alert"
        >
          <View style={styles.iconWrap}>
            <Ionicons name="notifications" size={20} color={colors.primary} />
          </View>
          <View style={styles.body}>
            <Text style={styles.title}>New job assigned</Text>
            <Text numberOfLines={1} style={styles.sub}>
              {job.customer_name ? `${job.customer_name} · ` : ''}
              {job.service_type}
            </Text>
            {typeof job.estimated_amount === 'number' ? (
              <Text numberOfLines={1} style={styles.meta}>
                {job.job_number} · ₹{job.estimated_amount}
              </Text>
            ) : null}
          </View>
          <Pressable hitSlop={8} onPress={hideNotice} style={styles.closeBtn}>
            <Ionicons name="close" size={18} color={colors.muted} />
          </Pressable>
        </Pressable>
      ) : null}
    </>
  )
}

function subscribe(listener: Listener) {
  listeners.add(listener)
  return () => listeners.delete(listener)
}

const styles = StyleSheet.create({
  toast: {
    position: 'absolute',
    left: spacing.lg,
    right: spacing.lg,
    flexDirection: 'row',
    alignItems: 'center',
    gap: spacing.md,
    backgroundColor: colors.surface,
    borderRadius: radius.lg,
    padding: spacing.md,
    ...shadows.raised,
    borderWidth: 1,
    borderColor: colors.border,
    zIndex: 100,
  },
  iconWrap: {
    width: 40,
    height: 40,
    borderRadius: 20,
    backgroundColor: colors.primarySoft,
    alignItems: 'center',
    justifyContent: 'center',
  },
  body: { flex: 1 },
  title: { fontSize: 14, fontWeight: '800', color: colors.text },
  sub: { fontSize: 13, color: colors.textSoft, marginTop: 2 },
  meta: { fontSize: 12, color: colors.muted, marginTop: 2, fontWeight: '600' },
  closeBtn: {
    width: 28,
    height: 28,
    borderRadius: 14,
    alignItems: 'center',
    justifyContent: 'center',
  },
})