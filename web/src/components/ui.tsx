import type { ReactNode } from 'react'

const statusColors: Record<string, string> = {
  PENDING: 'badge-amber',
  ACCEPTED: 'badge-blue',
  ON_THE_WAY: 'badge-indigo',
  STARTED: 'badge-purple',
  COMPLETED: 'badge-green',
  CANCELLED: 'badge-red',
  UNPAID: 'badge-red',
  PARTIAL: 'badge-amber',
  PAID: 'badge-green',
  AVAILABLE: 'badge-green',
  BUSY: 'badge-amber',
  INACTIVE: 'badge-gray',
}

export function StatusBadge({ value }: { value: string }) {
  const cls = statusColors[value] ?? 'badge-gray'
  return <span className={`badge ${cls}`}>{value.replace(/_/g, ' ')}</span>
}

const methodColors: Record<string, string> = {
  CASH: 'badge-green',
  UPI: 'badge-blue',
  CARD: 'badge-indigo',
  BANK_TRANSFER: 'badge-purple',
  OTHER: 'badge-gray',
}

export function MethodBadge({ value }: { value: string }) {
  return <span className={`badge ${methodColors[value] ?? 'badge-gray'}`}>{value.replace(/_/g, ' ')}</span>
}

export function PageHeader({
  title,
  subtitle,
  action,
}: {
  title: string
  subtitle?: string
  action?: ReactNode
}) {
  return (
    <div className="page-header">
      <div>
        <h1>{title}</h1>
        {subtitle && <p>{subtitle}</p>}
      </div>
      {action}
    </div>
  )
}

export function Card({ title, children, className = '' }: { title?: string; children: ReactNode; className?: string }) {
  return (
    <div className={`card ${className}`}>
      {title && <h3 className="card-title">{title}</h3>}
      {children}
    </div>
  )
}

export function StatCard({ label, value, hint }: { label: string; value: ReactNode; hint?: string }) {
  return (
    <div className="stat-card">
      <div className="stat-label">{label}</div>
      <div className="stat-value">{value}</div>
      {hint && <div className="stat-hint">{hint}</div>}
    </div>
  )
}

export function EmptyState({ message }: { message: string }) {
  return <div className="empty-state">{message}</div>
}

export function Spinner({ label = 'Loading…' }: { label?: string }) {
  return (
    <div className="spinner-wrap">
      <div className="spinner" />
      <span>{label}</span>
    </div>
  )
}

export function Field({
  label,
  children,
  required,
}: {
  label: string
  children: ReactNode
  required?: boolean
}) {
  return (
    <label className="field">
      <span className="field-label">
        {label}
        {required && <span className="required"> *</span>}
      </span>
      {children}
    </label>
  )
}

export function Alert({ kind, children }: { kind: 'error' | 'success' | 'info'; children: ReactNode }) {
  return <div className={`alert alert-${kind}`}>{children}</div>
}