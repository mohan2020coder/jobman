import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { api, errorMessage } from '../api/client'
import type { Customer, Job } from '../api/types'
import { Alert, Card, PageHeader, Spinner, StatusBadge } from '../components/ui'
import { formatDateTime, formatMoney } from '../lib/format'

export function CustomerDetailPage() {
  const { id } = useParams<{ id: string }>()
  const qc = useQueryClient()

  const customer = useQuery({
    queryKey: ['customer', id],
    queryFn: async () => (await api.get<Customer>(`/api/v1/customers/${id}`)).data,
  })

  const jobs = useQuery({
    queryKey: ['customer', id, 'jobs'],
    queryFn: async () =>
      (await api.get<{ data: Job[] }>(`/api/v1/customers/${id}/jobs`)).data.data,
  })

  const [editing, setEditing] = useState(false)
  const [name, setName] = useState('')
  const [phone, setPhone] = useState('')
  const [email, setEmail] = useState('')
  const [address, setAddress] = useState('')
  const [notes, setNotes] = useState('')
  const [error, setError] = useState('')

  const updateCustomer = useMutation({
    mutationFn: async () =>
      api.put<Customer>(`/api/v1/customers/${id}`, {
        name,
        phone,
        email: email || null,
        address: address || null,
        notes: notes || null,
      }),
    onSuccess: () => {
      setEditing(false)
      qc.invalidateQueries({ queryKey: ['customer', id] })
    },
    onError: (err) => setError(errorMessage(err)),
  })

  const deleteCustomer = useMutation({
    mutationFn: async () => api.delete(`/api/v1/customers/${id}`),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['customers'] }),
    onError: (err) => setError(errorMessage(err)),
  })

  if (customer.isLoading) return <Spinner />

  const c = customer.data
  if (!c) return <Alert kind="error">Customer not found.</Alert>

  const beginEdit = () => {
    setName(c.name)
    setPhone(c.phone)
    setEmail(c.email ?? '')
    setAddress(c.address ?? '')
    setNotes(c.notes ?? '')
    setEditing(true)
  }

  return (
    <div>
      <PageHeader
        title={c.name}
        subtitle={`Customer since ${formatDateTime(c.created_at)}`}
        action={
          <div className="row-buttons">
            {!editing && (
              <button className="btn btn-primary" onClick={beginEdit}>
                Edit
              </button>
            )}
            <button
              className="btn btn-danger"
              onClick={() => {
                if (window.confirm(`Delete customer ${c.name}?`)) deleteCustomer.mutate()
              }}
            >
              Delete
            </button>
          </div>
        }
      />

      {error && <Alert kind="error">{error}</Alert>}

      {editing ? (
        <Card title="Edit customer">
          <form
            className="stack"
            onSubmit={(e) => {
              e.preventDefault()
              setError('')
              updateCustomer.mutate()
            }}
          >
            <div className="grid-2">
              <label className="field">
                <span className="field-label">Name</span>
                <input className="input" required value={name} onChange={(e) => setName(e.target.value)} />
              </label>
              <label className="field">
                <span className="field-label">Phone</span>
                <input className="input" required value={phone} onChange={(e) => setPhone(e.target.value)} />
              </label>
              <label className="field">
                <span className="field-label">Email</span>
                <input className="input" value={email} onChange={(e) => setEmail(e.target.value)} />
              </label>
              <label className="field">
                <span className="field-label">Address</span>
                <input className="input" value={address} onChange={(e) => setAddress(e.target.value)} />
              </label>
            </div>
            <label className="field">
              <span className="field-label">Notes</span>
              <textarea className="input" rows={2} value={notes} onChange={(e) => setNotes(e.target.value)} />
            </label>
            <div className="row-buttons">
              <button className="btn btn-primary" disabled={updateCustomer.isPending}>
                {updateCustomer.isPending ? 'Saving…' : 'Save changes'}
              </button>
              <button type="button" className="btn btn-ghost" onClick={() => setEditing(false)}>
                Cancel
              </button>
            </div>
          </form>
        </Card>
      ) : (
        <div className="grid-detail">
          <Card title="Contact details">
            <div className="detail-block">
              <div className="detail-row">
                <span>Phone</span>
                <strong>{c.phone}</strong>
              </div>
              <div className="detail-row">
                <span>Email</span>
                <strong>{c.email ?? '—'}</strong>
              </div>
              <div className="detail-row">
                <span>Address</span>
                <strong>{c.address ?? '—'}</strong>
              </div>
              <div className="detail-row">
                <span>Notes</span>
                <strong>{c.notes ?? '—'}</strong>
              </div>
            </div>
          </Card>
        </div>
      )}

      <Card title={`Service history`}>
        {jobs.isLoading ? (
          <Spinner />
        ) : jobs.data && jobs.data.length > 0 ? (
          <table className="table">
            <thead>
              <tr>
                <th>Job</th>
                <th>Service</th>
                <th>Date</th>
                <th>Status</th>
                <th>Amount</th>
              </tr>
            </thead>
            <tbody>
              {jobs.data.map((job) => (
                <tr key={job.id}>
                  <td>
                    <Link to={`/jobs/${job.id}`} className="link">
                      {job.job_number}
                    </Link>
                  </td>
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
          <div className="muted">No service history for this customer yet.</div>
        )}
      </Card>
    </div>
  )
}