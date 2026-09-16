import { useQuery } from '@tanstack/react-query'
import { Link } from 'react-router-dom'
import { api } from '../api/client'
import type { Job, ListResponse, Payment, TodayDashboard } from '../api/types'
import { Card, EmptyState, MethodBadge, PageHeader, Spinner, StatusBadge } from '../components/ui'
import { formatDateTime, formatMoney } from '../lib/format'

export function PaymentsPage() {
  const dash = useQuery({
    queryKey: ['dashboard', 'payments'],
    queryFn: async () => (await api.get<TodayDashboard>('/api/v1/dashboard/today')).data,
  })

  const jobs = useQuery({
    queryKey: ['jobs', 'collections'],
    queryFn: async () =>
      (await api.get<ListResponse<Job>>('/api/v1/jobs', { params: { limit: 50 } })).data,
  })

  const withPayments = (jobs.data?.data ?? []).filter(
    (j) => (j.payments?.length ?? 0) > 0,
  )

  return (
    <div>
      <PageHeader title="Collections" subtitle="Payments collected today and per job" />

      {dash.isLoading ? (
        <Spinner />
      ) : (
        dash.data && (
          <div className="stat-grid">
            <div className="stat-card stat-highlight">
              <div className="stat-label">Today's collection</div>
              <div className="stat-value">{formatMoney(dash.data.collection.total)}</div>
            </div>
            <div className="stat-card">
              <div className="stat-label">Cash</div>
              <div className="stat-value">{formatMoney(dash.data.collection.cash)}</div>
            </div>
            <div className="stat-card">
              <div className="stat-label">UPI</div>
              <div className="stat-value">{formatMoney(dash.data.collection.upi)}</div>
            </div>
            <div className="stat-card">
              <div className="stat-label">Card</div>
              <div className="stat-value">{formatMoney(dash.data.collection.card)}</div>
            </div>
            <div className="stat-card">
              <div className="stat-label">Bank transfer</div>
              <div className="stat-value">{formatMoney(dash.data.collection.bank_transfer)}</div>
            </div>
          </div>
        )
      )}

      <Card title="Payments by job">
        {jobs.isLoading ? (
          <Spinner />
        ) : withPayments.length > 0 ? (
          <table className="table">
            <thead>
              <tr>
                <th>Job</th>
                <th>Customer</th>
                <th>Completed</th>
                <th>Final amount</th>
                <th>Paid</th>
                <th>Status</th>
                <th>Methods</th>
              </tr>
            </thead>
            <tbody>
              {withPayments.map((job) => (
                <tr key={job.id}>
                  <td>
                    <Link to={`/jobs/${job.id}`} className="link">
                      {job.job_number}
                    </Link>
                  </td>
                  <td>{job.customer?.name ?? '—'}</td>
                  <td>{formatDateTime(job.completed_at)}</td>
                  <td>{formatMoney(job.final_amount)}</td>
                  <td>{formatMoney(job.total_paid ?? 0)}</td>
                  <td>
                    <StatusBadge value={job.payment_status ?? 'UNPAID'} />
                  </td>
                  <td>
                    <PaymentMethods payments={job.payments ?? []} />
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        ) : (
          <EmptyState message="No payments recorded yet." />
        )}
      </Card>
    </div>
  )
}

function PaymentMethods({ payments }: { payments: Payment[] }) {
  if (payments.length === 0) return <span className="muted">—</span>
  return (
    <span className="badge-row">
      {payments.map((p) => (
        <MethodBadge key={p.id} value={p.method} />
      ))}
    </span>
  )
}