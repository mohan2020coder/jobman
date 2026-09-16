import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { api, errorMessage } from '../api/client'
import type { Job, Technician } from '../api/types'
import { Alert, Card, PageHeader, Spinner, StatusBadge } from '../components/ui'
import { formatDateTime, formatMoney } from '../lib/format'

export function TechnicianDetailPage() {
  const { id } = useParams<{ id: string }>()
  const qc = useQueryClient()

  const technician = useQuery({
    queryKey: ['technician', id],
    queryFn: async () => (await api.get<Technician>(`/api/v1/technicians/${id}`)).data,
  })

  const jobs = useQuery({
    queryKey: ['technician', id, 'jobs'],
    queryFn: async () =>
      (await api.get<{ data: Job[] }>(`/api/v1/technicians/${id}/jobs`)).data.data,
  })

  const [error, setError] = useState('')

  const updateTech = useMutation({
    mutationFn: async (body: { status?: string; is_active?: boolean }) =>
      api.put<Technician>(`/api/v1/technicians/${id}`, body),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['technician', id] }),
    onError: (err) => setError(errorMessage(err)),
  })

  if (technician.isLoading) return <Spinner />

  const t = technician.data
  if (!t) return <Alert kind="error">Technician not found.</Alert>

  const active = Number(jobs.data?.filter((j) => j.status !== 'CANCELLED').length ?? 0)
  const assignedJobs = jobs.data ?? []

  return (
    <div>
      <PageHeader
        title={t.name}
        subtitle={`Member since ${formatDateTime(t.created_at)}`}
        action={
          <div className="row-buttons">
            <StatusBadge value={t.status} />
            {t.is_active ? (
              <button
                className="btn btn-danger"
                onClick={() => {
                  if (window.confirm(`Deactivate ${t.name}?`)) {
                    updateTech.mutate({ is_active: false, status: 'INACTIVE' })
                  }
                }}
              >
                Deactivate
              </button>
            ) : (
              <button className="btn btn-primary" onClick={() => updateTech.mutate({ is_active: true, status: 'AVAILABLE' })}>
                Activate
              </button>
            )}
          </div>
        }
      />

      {error && <Alert kind="error">{error}</Alert>}

      <div className="grid-detail">
        <Card title="Details">
          <div className="detail-block">
            <div className="detail-row">
              <span>Phone</span>
              <strong>{t.phone}</strong>
            </div>
            <div className="detail-row">
              <span>Email</span>
              <strong>{t.email ?? '—'}</strong>
            </div>
            <div className="detail-row">
              <span>Mobile login</span>
              <strong>{t.phone}</strong>
            </div>
            <div className="detail-row">
              <span>Active</span>
              <strong>{t.is_active ? 'Yes' : 'No'}</strong>
            </div>
            <div className="detail-row">
              <span>Last login</span>
              <strong>{t.last_login_at ? formatDateTime(t.last_login_at) : 'Never'}</strong>
            </div>
          </div>
        </Card>

        <Card title="Current workload">
          <div className="detail-block">
            <div className="detail-row">
              <span>Assigned jobs</span>
              <strong>{active}</strong>
            </div>
          </div>
        </Card>
      </div>

      <Card title="Jobs">
        {jobs.isLoading ? (
          <Spinner />
        ) : assignedJobs.length > 0 ? (
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
              {assignedJobs.map((job) => (
                <tr key={job.id}>
                  <td>
                    <Link to={`/jobs/${job.id}`} className="link">
                      {job.job_number}
                    </Link>
                  </td>
                  <td>{job.customer?.name ?? '—'}</td>
                  <td>{job.service_type}</td>
                  <td>{formatDateTime(job.scheduled_at)}</td>
                  <td>
                    <StatusBadge value={job.status} />
                  </td>
                  <td>{formatMoney(job.final_amount || job.estimated_amount)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        ) : (
          <div className="muted">No jobs assigned yet.</div>
        )}
      </Card>
    </div>
  )
}