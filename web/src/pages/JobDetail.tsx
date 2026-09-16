import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import { useParams } from 'react-router-dom'
import { api, errorMessage, publicReceiptUrl } from '../api/client'
import type { Job, Payment } from '../api/types'
import { Alert, Card, MethodBadge, PageHeader, Spinner, StatusBadge } from '../components/ui'
import { formatDate, formatDateTime, formatMoney, formatTime } from '../lib/format'

function useJob(id: string) {
  return useQuery({
    queryKey: ['job', id],
    queryFn: async () => (await api.get<Job>(`/api/v1/jobs/${id}`)).data,
  })
}

const cancellable = new Set(['PENDING', 'ACCEPTED', 'ON_THE_WAY'])

export function JobDetailPage() {
  const { id } = useParams<{ id: string }>()
  const qc = useQueryClient()
  const job = useJob(id!)
  const [error, setError] = useState('')

  const [payAmount, setPayAmount] = useState('')
  const [payMethod, setPayMethod] = useState('CASH')
  const [payRef, setPayRef] = useState('')

  const cancelJob = useMutation({
    mutationFn: async () => api.post(`/api/v1/jobs/${id}/cancel`),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['job', id] })
      qc.invalidateQueries({ queryKey: ['jobs'] })
      qc.invalidateQueries({ queryKey: ['dashboard'] })
    },
    onError: (err) => setError(errorMessage(err)),
  })

  const addPayment = useMutation({
    mutationFn: async () =>
      api.post<Payment>(`/api/v1/jobs/${id}/payments`, {
        amount: Number(payAmount),
        method: payMethod,
        transaction_reference: payRef || undefined,
      }),
    onSuccess: () => {
      setPayAmount('')
      setPayRef('')
      qc.invalidateQueries({ queryKey: ['job', id] })
      qc.invalidateQueries({ queryKey: ['dashboard'] })
    },
    onError: (err) => setError(errorMessage(err)),
  })

  if (job.isLoading) return <Spinner />
  if (job.isError || !job.data) {
    return <Alert kind="error">Unable to load job.</Alert>
  }

  const j = job.data
  const remaining = Math.max(0, (j.final_amount || j.estimated_amount) - (j.total_paid ?? 0))

  return (
    <div>
      <PageHeader
        title={`${j.job_number} · ${j.service_type}`}
        subtitle={`Created ${formatDateTime(j.created_at)}`}
        action={
          <div className="row-buttons">
            <StatusBadge value={j.status} />
            {cancellable.has(j.status) && (
              <button
                className="btn btn-danger"
                disabled={cancelJob.isPending}
                onClick={() => {
                  if (window.confirm('Cancel this job?')) cancelJob.mutate()
                }}
              >
                Cancel job
              </button>
            )}
          </div>
        }
      />

      {error && <Alert kind="error">{error}</Alert>}

      <div className="grid-detail">
        <Card title="Customer">
          {j.customer && (
            <div className="detail-block">
              <div className="detail-name">{j.customer.name}</div>
              <div className="muted">{j.customer.phone}</div>
              {j.customer.email && <div className="muted">{j.customer.email}</div>}
              {j.customer.address && <div className="muted">{j.customer.address}</div>}
            </div>
          )}
        </Card>

        <Card title="Technician">
          {j.technician ? (
            <div className="detail-block">
              <div className="detail-name">{j.technician.name}</div>
              <div className="muted">{j.technician.phone}</div>
              <StatusBadge value={j.technician.status} />
            </div>
          ) : (
            <div className="muted">Unassigned</div>
          )}
        </Card>

        <Card title="Job details">
          <div className="detail-block">
            <div className="detail-row">
              <span>Problem</span>
              <strong>{j.problem_description ?? '—'}</strong>
            </div>
            <div className="detail-row">
              <span>Address</span>
              <strong>{j.address ?? '—'}</strong>
            </div>
            <div className="detail-row">
              <span>Scheduled</span>
              <strong>{formatDateTime(j.scheduled_at)}</strong>
            </div>
            <div className="detail-row">
              <span>Estimated</span>
              <strong>{formatMoney(j.estimated_amount)}</strong>
            </div>
            <div className="detail-row">
              <span>Final</span>
              <strong>{formatMoney(j.final_amount)}</strong>
            </div>
            <div className="detail-row">
              <span>Payment status</span>
              <StatusBadge value={j.payment_status ?? 'UNPAID'} />
            </div>
          </div>
        </Card>
      </div>

      <div className="grid-2-stack">
        <Card title="Items & charges">
          {j.items && j.items.length > 0 ? (
            <table className="table">
              <thead>
                <tr>
                  <th>Description</th>
                  <th>Qty</th>
                  <th>Unit price</th>
                  <th className="align-right">Total</th>
                </tr>
              </thead>
              <tbody>
                {j.items.map((item) => (
                  <tr key={item.id}>
                    <td>{item.description}</td>
                    <td>{item.quantity}</td>
                    <td>{formatMoney(item.unit_price)}</td>
                    <td className="align-right">{formatMoney(item.total_price)}</td>
                  </tr>
                ))}
              </tbody>
              <tfoot>
                <tr>
                  <td colSpan={3}>Total</td>
                  <td className="align-right">{formatMoney(j.final_amount)}</td>
                </tr>
              </tfoot>
            </table>
          ) : (
            <div className="muted">No items yet. Items are added when the job is completed.</div>
          )}

          {j.notes && (
            <div className="notes-block">
              <strong>Notes:</strong> {j.notes}
            </div>
          )}
        </Card>

        <Card title="Payments">
          {j.payments && j.payments.length > 0 ? (
            <table className="table">
              <thead>
                <tr>
                  <th>Date</th>
                  <th>Method</th>
                  <th>Ref</th>
                  <th className="align-right">Amount</th>
                </tr>
              </thead>
              <tbody>
                {j.payments.map((p) => (
                  <tr key={p.id}>
                    <td>{formatDate(p.paid_at)} {formatTime(p.paid_at)}</td>
                    <td>
                      <MethodBadge value={p.method} />
                    </td>
                    <td>{p.transaction_reference ?? '—'}</td>
                    <td className="align-right">{formatMoney(p.amount)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          ) : (
            <div className="muted">No payments recorded.</div>
          )}

          {remaining > 0 && j.status !== 'CANCELLED' && (
            <form
              className="payment-form"
              onSubmit={(e) => {
                e.preventDefault()
                setError('')
                addPayment.mutate()
              }}
            >
              <strong>Record payment</strong>
              <div className="grid-3">
                <input
                  className="input"
                  type="number"
                  min="0.01"
                  step="0.01"
                  max={remaining}
                  required
                  placeholder={`Up to ${formatMoney(remaining)}`}
                  value={payAmount}
                  onChange={(e) => setPayAmount(e.target.value)}
                />
                <select className="select" value={payMethod} onChange={(e) => setPayMethod(e.target.value)}>
                  <option value="CASH">Cash</option>
                  <option value="UPI">UPI</option>
                  <option value="CARD">Card</option>
                  <option value="BANK_TRANSFER">Bank transfer</option>
                  <option value="OTHER">Other</option>
                </select>
                <input
                  className="input"
                  placeholder="Ref (optional)"
                  value={payRef}
                  onChange={(e) => setPayRef(e.target.value)}
                />
              </div>
              <button className="btn btn-primary" disabled={addPayment.isPending}>
                {addPayment.isPending ? 'Saving…' : 'Add payment'}
              </button>
            </form>
          )}
        </Card>
      </div>

      <div className="grid-2-stack">
        <Card title="Receipt">
          {j.receipt ? (
            <div>
              <div className="detail-block">
                <div className="detail-row">
                  <span>Receipt no.</span>
                  <strong>{j.receipt.receipt_number}</strong>
                </div>
                <div className="detail-row">
                  <span>Issued</span>
                  <strong>{formatDateTime(j.receipt.issued_at)}</strong>
                </div>
              </div>
              <ReceiptLink jobId={j.id} />
            </div>
          ) : (
            <div className="muted">
              No receipt yet. At least one payment is recorded when the job is completed.
            </div>
          )}
        </Card>

        <Card title="Status history">
          {j.status_history && j.status_history.length > 0 ? (
            <ol className="timeline">
              {j.status_history.map((h) => (
                <li key={h.id}>
                  <div className="timeline-status">
                    {h.old_status ? `${h.old_status} → ` : ''}
                    <strong>{h.new_status}</strong>
                  </div>
                  <div className="muted">{formatDateTime(h.created_at)}</div>
                </li>
              ))}
            </ol>
          ) : (
            <div className="muted">No status changes recorded.</div>
          )}
        </Card>
      </div>
    </div>
  )
}

function ReceiptLink({ jobId }: { jobId: string }) {
  const res = useQuery({
    queryKey: ['receipt', jobId],
    queryFn: async () => (await api.get<{ public_url: string }>(`/api/v1/jobs/${jobId}/receipt`)).data,
  })
  if (!res.data?.public_url) return null
  const url = publicReceiptUrl(res.data.public_url)
  return (
    <a className="btn btn-ghost" href={url} target="_blank" rel="noreferrer">
      Open public receipt ↗
    </a>
  )
}