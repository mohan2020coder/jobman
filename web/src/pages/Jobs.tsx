import { useQuery } from '@tanstack/react-query'
import { useMemo, useState } from 'react'
import { Link } from 'react-router-dom'
import { api } from '../api/client'
import type { Customer, Job, ListResponse, Technician } from '../api/types'
import { Card, EmptyState, PageHeader, Spinner, StatusBadge } from '../components/ui'
import { formatMoney, formatTime } from '../lib/format'

const statusOptions = [
  'PENDING',
  'ACCEPTED',
  'ON_THE_WAY',
  'STARTED',
  'COMPLETED',
  'CANCELLED',
]

export function JobsPage() {
  const [status, setStatus] = useState('')
  const [technicianId, setTechnicianId] = useState('')
  const [customerId, setCustomerId] = useState('')
  const [date, setDate] = useState('')
  const [search, setSearch] = useState('')
  const [page, setPage] = useState(1)

  const queryKey = ['jobs', 'list', status, technicianId, customerId, date, search, page]
  const jobs = useQuery({
    queryKey,
    queryFn: async () => {
      const res = await api.get<ListResponse<Job>>('/api/v1/jobs', {
        params: { status, technician_id: technicianId, customer_id: customerId, date, search, page, limit: 15 },
      })
      return res.data
    },
  })

  const customers = useQuery({
    queryKey: ['customers', 'options'],
    queryFn: async () => (await api.get<ListResponse<Customer>>('/api/v1/customers', { params: { limit: 100 } })).data,
  })

  const technicians = useQuery({
    queryKey: ['technicians', 'options'],
    queryFn: async () => (await api.get<{ data: Technician[] }>('/api/v1/technicians')).data,
  })

  const totalPages = useMemo(() => {
    const t = jobs.data?.pagination.total ?? 0
    const l = jobs.data?.pagination.limit ?? 15
    return Math.max(1, Math.ceil(t / l))
  }, [jobs.data])

  function resetPage() {
    setPage(1)
  }

  return (
    <div>
      <PageHeader
        title="Jobs"
        subtitle="Create, assign and track service jobs"
        action={
          <Link to="/jobs/new" className="btn btn-primary">
            New job
          </Link>
        }
      />

      <Card>
        <div className="filter-row">
          <input
            className="input"
            type="search"
            placeholder="Search job / customer / service…"
            value={search}
            onChange={(e) => {
              setSearch(e.target.value)
              resetPage()
            }}
          />
          <select className="select" value={status} onChange={(e) => { setStatus(e.target.value); resetPage() }}>
            <option value="">All statuses</option>
            {statusOptions.map((s) => (
              <option key={s} value={s}>
                {s.replace(/_/g, ' ')}
              </option>
            ))}
          </select>
          <select className="select" value={technicianId} onChange={(e) => { setTechnicianId(e.target.value); resetPage() }}>
            <option value="">All technicians</option>
            {technicians.data?.data.map((t) => (
              <option key={t.id} value={t.id}>
                {t.name}
              </option>
            ))}
          </select>
          <select className="select" value={customerId} onChange={(e) => { setCustomerId(e.target.value); resetPage() }}>
            <option value="">All customers</option>
            {customers.data?.data.map((c) => (
              <option key={c.id} value={c.id}>
                {c.name}
              </option>
            ))}
          </select>
          <input
            className="input"
            type="date"
            value={date}
            onChange={(e) => {
              setDate(e.target.value)
              resetPage()
            }}
          />
        </div>
      </Card>

      {jobs.isLoading ? (
        <Spinner />
      ) : jobs.data && jobs.data.data.length > 0 ? (
        <Card>
          <table className="table">
            <thead>
              <tr>
                <th>Job</th>
                <th>Customer</th>
                <th>Technician</th>
                <th>Service</th>
                <th>Scheduled</th>
                <th>Status</th>
                <th>Amount</th>
                <th>Paid</th>
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
                  <td>{job.technician?.name ?? 'Unassigned'}</td>
                  <td>{job.service_type}</td>
                  <td>{formatTime(job.scheduled_at)}</td>
                  <td>
                    <StatusBadge value={job.status} />
                  </td>
                  <td>{formatMoney(job.final_amount || job.estimated_amount)}</td>
                  <td>{formatMoney(job.total_paid ?? 0)}</td>
                </tr>
              ))}
            </tbody>
          </table>

          <div className="pagination">
            <button className="btn btn-ghost" disabled={page <= 1} onClick={() => setPage((p) => p - 1)}>
              Previous
            </button>
            <span className="pagination-info">
              Page {page} of {totalPages} · {jobs.data.pagination.total} total
            </span>
            <button className="btn btn-ghost" disabled={page >= totalPages} onClick={() => setPage((p) => p + 1)}>
              Next
            </button>
          </div>
        </Card>
      ) : (
        <Card>
          <EmptyState message="No jobs match your filters." />
        </Card>
      )}
    </div>
  )
}