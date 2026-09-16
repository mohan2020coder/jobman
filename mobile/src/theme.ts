export const colors = {
  primary: '#2563eb',
  primaryDark: '#1d4ed8',
  primarySoft: '#eff6ff',

  accent: '#6366f1',
  accentDark: '#4f46e5',
  accentSoft: '#eef2ff',

  bg: '#f4f6fb',
  surface: '#ffffff',
  surfaceAlt: '#f8fafc',
  border: '#e2e8f0',
  borderStrong: '#cbd5e1',

  text: '#0f172a',
  textSoft: '#334155',
  muted: '#64748b',
  faint: '#94a3b8',

  danger: '#dc2626',
  dangerDark: '#b91c1c',
  dangerSoft: '#fef2f2',

  success: '#059669',
  successSoft: '#ecfdf5',

  warning: '#d97706',
  warningSoft: '#fffbeb',

  info: '#2563eb',

  white: '#ffffff',
}

export type StatusKey =
  | 'PENDING'
  | 'ACCEPTED'
  | 'ON_THE_WAY'
  | 'STARTED'
  | 'COMPLETED'
  | 'CANCELLED'
  | 'UNPAID'
  | 'PARTIAL'
  | 'PAID'
  | 'AVAILABLE'
  | 'BUSY'
  | 'INACTIVE'

export const statusColor: Record<StatusKey, string> = {
  PENDING: '#d97706',
  ACCEPTED: '#2563eb',
  ON_THE_WAY: '#7c3aed',
  STARTED: '#0891b2',
  COMPLETED: '#059669',
  CANCELLED: '#64748b',
  UNPAID: '#dc2626',
  PARTIAL: '#d97706',
  PAID: '#059669',
  AVAILABLE: '#059669',
  BUSY: '#d97706',
  INACTIVE: '#94a3b8',
}

export const statusSoft: Record<StatusKey, string> = {
  PENDING: '#fef3c7',
  ACCEPTED: '#dbeafe',
  ON_THE_WAY: '#ede9fe',
  STARTED: '#cffafe',
  COMPLETED: '#d1fae5',
  CANCELLED: '#e2e8f0',
  UNPAID: '#fee2e2',
  PARTIAL: '#fef3c7',
  PAID: '#d1fae5',
  AVAILABLE: '#d1fae5',
  BUSY: '#fef3c7',
  INACTIVE: '#e2e8f0',
}

export const methodColor: Record<string, string> = {
  CASH: '#059669',
  UPI: '#2563eb',
  CARD: '#7c3aed',
  BANK_TRANSFER: '#0891b2',
  OTHER: '#64748b',
}

export function tint(hex: string, alpha: number): string {
  const r = parseInt(hex.slice(1, 3), 16)
  const g = parseInt(hex.slice(3, 5), 16)
  const b = parseInt(hex.slice(5, 7), 16)
  return `rgba(${r},${g},${b},${alpha})`
}

export const spacing = {
  xs: 4,
  sm: 8,
  md: 12,
  lg: 16,
  xl: 24,
  xxl: 32,
}

export const radius = {
  sm: 8,
  md: 12,
  lg: 16,
  xl: 22,
  pill: 999,
}

export const shadows = {
  card: {
    shadowColor: '#0f172a',
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.05,
    shadowRadius: 8,
    elevation: 2,
  },
  raised: {
    shadowColor: '#2563eb',
    shadowOffset: { width: 0, height: 6 },
    shadowOpacity: 0.25,
    shadowRadius: 12,
    elevation: 6,
  },
}

export const typography = {
  title: { fontSize: 26, fontWeight: '800' as const, letterSpacing: -0.4 },
  h2: { fontSize: 20, fontWeight: '800' as const, letterSpacing: -0.2 },
  h3: { fontSize: 17, fontWeight: '700' as const },
  body: { fontSize: 15, fontWeight: '400' as const },
  small: { fontSize: 13, fontWeight: '500' as const },
  tiny: { fontSize: 11, fontWeight: '600' as const },
}