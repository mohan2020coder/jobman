import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { api, errorMessage } from '../api/client'
import type { Customer, ListResponse, Technician } from '../api/types'
import { Alert, Field, PageHeader } from '../components/ui'
import { toLocalInputValue, toRfc3339 } from '../lib/format'

export function JobNewPage() {
  const navigate = useNavigate()
  const qc = useQueryClient()

  const [customerId, setCustomerId] = useState('')
  const [technicianId, setTechnicianId] = useState('')
  const [serviceType, setServiceType] = useState('')
  const [problemDescription, setProblemDescription] = useState('')
  const [address, setAddress] = useState('')
  const [estimatedAmount, setEstimatedAmount] = useState('')
  const [scheduledAt, setScheduledAt] = useState(
    toLocalInputValue(new Date().toISOString()),
  )
  const [error, setError] = useState('')

  const [quickCustomer, setQuickCustomer] = useState(false)
  const [qcName, setQcName] = useState('')
  const [qcPhone, setQcPhone] = useState('')

  const customers = useQuery({
    queryKey: ['customers', 'new-job'],
    queryFn: async () =>
      (await api.get<ListResponse<Customer>>('/api/v1/customers', { params: { limit: 100 } })).data,
  })

  const technicians = useQuery({
    queryKey: ['technicians', 'new-job'],
    queryFn: async () => (await api.get<{ data: Technician[] }>('/api/v1/technicians')).data,
  })

  const createJob = useMutation({
    mutationFn: async () => {
      let custId = customerId
      if (quickCustomer) {
        const created = await api.post<Customer>('/api/v1/customers', {
          name: qcName,
          phone: qcPhone,
        })
        custId = created.data.id
      }
      return api.post('/api/v1/jobs', {
        customer_id: custId,
        technician_id: technicianId || undefined,
        service_type: serviceType,
        problem_description: problemDescription || undefined,
        address: address || undefined,
        estimated_amount: estimatedAmount ? Number(estimatedAmount) : 0,
        scheduled_at: toRfc3339(scheduledAt),
      })
    },
    onSuccess: async (res) => {
      await qc.invalidateQueries({ queryKey: ['jobs'] })
      await qc.invalidateQueries({ queryKey: ['dashboard'] })
      navigate(`/jobs/${res.data.id}`, { replace: true })
    },
    onError: (err) => setError(errorMessage(err)),
  })

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setError('')
    createJob.mutate()
  }

  const activeTechnicians = technicians.data?.data.filter((t) => t.is_active) ?? []

  return (
    <div className="page-narrow">
      <PageHeader title="New job" subtitle="Create a job in under 30 seconds" />

      <form onSubmit={handleSubmit} className="stack">
        {quickCustomer ? (
          <>
            <Alert kind="info">Creating a new customer along with the job.</Alert>
            <Field label="Customer name" required>
              <input className="input" required value={qcName} onChange={(e) => setQcName(e.target.value)} placeholder="Ravi Kumar" />
            </Field>
            <Field label="Customer phone" required>
              <input className="input" required value={qcPhone} onChange={(e) => setQcPhone(e.target.value)} placeholder="9876543210" />
            </Field>
            <button type="button" className="btn btn-ghost" onClick={() => setQuickCustomer(false)}>
              Choose existing customer instead
            </button>
          </>
        ) : (
          <>
            <Field label="Customer" required>
              <select className="select" required value={customerId} onChange={(e) => setCustomerId(e.target.value)}>
                <option value="">Select customer…</option>
                {customers.data?.data.map((c) => (
                  <option key={c.id} value={c.id}>
                    {c.name} · {c.phone}
                  </option>
                ))}
              </select>
            </Field>
            <button type="button" className="btn btn-ghost" onClick={() => setQuickCustomer(true)}>
              + New customer
            </button>
          </>
        )}

        <Field label="Service type" required>
          <input
            className="input"
            required
            value={serviceType}
            onChange={(e) => setServiceType(e.target.value)}
            placeholder="AC Repair"
            list="service-types"
          />
          <datalist id="service-types">
            <option value="AC Repair" />
            <option value="AC Service" />
            <option value="Refrigerator Repair" />
            <option value="Washing Machine Repair" />
            <option value="Plumbing" />
            <option value="Electrical" />
            <option value="RO Service" />
            <option value="CCTV Installation" />
            <option value="Pest Control" />
          </datalist>
        </Field>

        <Field label="Problem description">
          <textarea
            className="input"
            rows={3}
            value={problemDescription}
            onChange={(e) => setProblemDescription(e.target.value)}
            placeholder="AC not cooling"
          />
        </Field>

        <Field label="Address">
          <input
            className="input"
            value={address}
            onChange={(e) => setAddress(e.target.value)}
            placeholder="12 MG Road"
          />
        </Field>

        <div className="grid-2">
          <Field label="Estimated amount (₹)">
            <input
              className="input"
              type="number"
              min="0"
              step="0.01"
              value={estimatedAmount}
              onChange={(e) => setEstimatedAmount(e.target.value)}
              placeholder="1500"
            />
          </Field>
          <Field label="Scheduled time">
            <input
              className="input"
              type="datetime-local"
              value={scheduledAt}
              onChange={(e) => setScheduledAt(e.target.value)}
            />
          </Field>
        </div>

        <Field label="Assign technician (optional)">
          <select className="select" value={technicianId} onChange={(e) => setTechnicianId(e.target.value)}>
            <option value="">Assign later</option>
            {activeTechnicians.map((t) => (
              <option key={t.id} value={t.id}>
                {t.name} · {t.phone}
              </option>
            ))}
          </select>
        </Field>

        {error && <Alert kind="error">{error}</Alert>}

        <div className="row-buttons">
          <button className="btn btn-primary" disabled={createJob.isPending}>
            {createJob.isPending ? 'Creating…' : 'Create job'}
          </button>
          <button type="button" className="btn btn-ghost" onClick={() => navigate('/jobs')}>
            Cancel
          </button>
        </div>
      </form>
    </div>
  )
}