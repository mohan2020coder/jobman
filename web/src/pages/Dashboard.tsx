import { useQuery } from '@tanstack/react-query'
import { Link } from 'react-router-dom'
import { api } from '../api/client'
import type { Job, TodayDashboard } from '../api/types'
import { Card, EmptyState, PageHeader, Spinner, StatCard, StatusBadge } from '../components/ui'
import { formatMoney, formatTime, greeting } from '../lib/format'
import { useAuth } from '../context/AuthContext'

function useTodayDashboard() {
  return useQuery({
    queryKey: ['dashboard', 'today'],
    queryFn: async () => (await api.get<TodayDashboard>('/api/v1/dashboard/today')).data,
  })
}

function useTodayJobs() {
  return useQuery({
    queryKey: ['jobs', 'today'],
    queryFn: async () => {
      const res = await api.get<{ data: Job[]; pagination: { total: number } }>('/api/v1/jobs', {
        params: { limit: 10, sort_by: 'scheduled_at' },
      })
      return res.data
    },
  })
}

export function DashboardPage() {
  const { user } = useAuth()
  const dash = useTodayDashboard()
  const jobs = useTodayJobs()

  return (
    <div>
      <PageHeader
        title={`${greeting()}, ${user?.name ?? ''}`}
        subtitle="Here's what's happening today"
      />

      {dash.isLoading ? (
        <Spinner />
      ) : (
        <div className="stat-grid">
          <StatCard label="Today's Jobs" value={dash.data?.jobs.total ?? 0} />
          <StatCard label="Pending" value={dash.data?.jobs.pending ?? 0} />
          <StatCard label="Accepted" value={dash.data?.jobs.accepted ?? 0} />
          <StatCard label="On the Way" value={dash.data?.jobs.on_the_way ?? 0} />
          <StatCard label="Started" value={dash.data?.jobs.started ?? 0} />
          <StatCard label="Completed" value={dash.data?.jobs.completed ?? 0} />
          <StatCard label="Collection" value={formatMoney(dash.data?.collection.total)} />
        </div>
      )}

      {!dash.isLoading && dash.data && (
        <Card title="Collection by method">
          <div className="collection-grid">
            <div className="collection-item">
              <span>Cash</span>
              <strong>{formatMoney(dash.data.collection.cash)}</strong>
            </div>
            <div className="collection-item">
              <span>UPI</span>
              <strong>{formatMoney(dash.data.collection.upi)}</strong>
            </div>
            <div className="collection-item">
              <span>Card</span>
              <strong>{formatMoney(dash.data.collection.card)}</strong>
            </div>
            <div className="collection-item">
              <span>Bank transfer</span>
              <strong>{formatMoney(dash.data.collection.bank_transfer)}</strong>
            </div>
            <div className="collection-item">
              <span>Other</span>
              <strong>{formatMoney(dash.data.collection.other)}</strong>
            </div>
          </div>
        </Card>
      )}

      <Card title="Recent jobs">
        {jobs.isLoading ? (
          <Spinner />
        ) : jobs.data && jobs.data.data.length > 0 ? (
          <table className="table">
            <thead>
              <tr>
                <th>Job</th>
                <th>Customer</th>
                <th>Service</th>
                <th>Scheduled</th>
                <th>Status</th>
                <th>Amount</th>
              </tr>
            </thead>
            <tbody>
              {jobs.data.data.map((job) => (
                <tr key={job.id}>
                  <td>
                    <Link to={`/jobs/${job.id}`} className="link">
                      {job.job_number}
                    </Link>
                  </td>
                  <td>{job.customer?.name ?? '—'}</td>
                  <td>{job.service_type}</td>
                  <td>
                    {formatTime(job.scheduled_at)} · {job.customer?.address ?? '—'}
                  </td>
                  <td>
                    <StatusBadge value={job.status} />
                  </td>
                  <td>{formatMoney(job.final_amount || job.estimated_amount)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        ) : (
          <EmptyState message="No jobs yet. Create your first job." />
        )}
      </Card>
    </div>
  )
}