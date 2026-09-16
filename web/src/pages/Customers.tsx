import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useMemo, useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { api, errorMessage } from '../api/client'
import type { Customer, ListResponse } from '../api/types'
import { Alert, Card, EmptyState, Field, PageHeader, Spinner } from '../components/ui'

export function CustomersPage() {
  const qc = useQueryClient()
  const navigate = useNavigate()
  const [search, setSearch] = useState('')
  const [page, setPage] = useState(1)

  const customers = useQuery({
    queryKey: ['customers', 'list', search, page],
    queryFn: async () =>
      (await api.get<ListResponse<Customer>>('/api/v1/customers', { params: { search, page, limit: 15 } })).data,
  })

  const [creating, setCreating] = useState(false)
  const [name, setName] = useState('')
  const [phone, setPhone] = useState('')
  const [email, setEmail] = useState('')
  const [address, setAddress] = useState('')
  const [notes, setNotes] = useState('')
  const [error, setError] = useState('')

  const createCustomer = useMutation({
    mutationFn: async () =>
      api.post<Customer>('/api/v1/customers', {
        name,
        phone,
        email: email || undefined,
        address: address || undefined,
        notes: notes || undefined,
      }),
    onSuccess: (res) => {
      setCreating(false)
      setName('')
      setPhone('')
      setEmail('')
      setAddress('')
      setNotes('')
      qc.invalidateQueries({ queryKey: ['customers'] })
      navigate(`/customers/${res.data.id}`)
    },
    onError: (err) => setError(errorMessage(err)),
  })

  const totalPages = useMemo(() => {
    const t = customers.data?.pagination.total ?? 0
    const l = customers.data?.pagination.limit ?? 15
    return Math.max(1, Math.ceil(t / l))
  }, [customers.data])

  return (
    <div>
      <PageHeader
        title="Customers"
        subtitle="Manage your customer list"
        action={
          <button className="btn btn-primary" onClick={() => setCreating((v) => !v)}>
            {creating ? 'Close' : 'Add customer'}
          </button>
        }
      />

      {creating && (
        <Card title="Add customer">
          <form
            className="stack"
            onSubmit={(e) => {
              e.preventDefault()
              setError('')
              createCustomer.mutate()
            }}
          >
            <div className="grid-2">
              <Field label="Name" required>
                <input className="input" required value={name} onChange={(e) => setName(e.target.value)} />
              </Field>
              <Field label="Phone" required>
                <input className="input" required value={phone} onChange={(e) => setPhone(e.target.value)} />
              </Field>
              <Field label="Email">
                <input className="input" value={email} onChange={(e) => setEmail(e.target.value)} />
              </Field>
              <Field label="Address">
                <input className="input" value={address} onChange={(e) => setAddress(e.target.value)} />
              </Field>
            </div>
            <Field label="Notes">
              <textarea className="input" rows={2} value={notes} onChange={(e) => setNotes(e.target.value)} />
            </Field>
            {error && <Alert kind="error">{error}</Alert>}
            <button className="btn btn-primary" disabled={createCustomer.isPending}>
              {createCustomer.isPending ? 'Saving…' : 'Save customer'}
            </button>
          </form>
        </Card>
      )}

      <Card>
        <div className="filter-row">
          <input
            className="input"
            type="search"
            placeholder="Search by name or phone…"
            value={search}
            onChange={(e) => {
              setSearch(e.target.value)
              setPage(1)
            }}
          />
        </div>
      </Card>

      {customers.isLoading ? (
        <Spinner />
      ) : customers.data && customers.data.data.length > 0 ? (
        <Card>
          <table className="table">
            <thead>
              <tr>
                <th>Name</th>
                <th>Phone</th>
                <th>Address</th>
                <th>Added</th>
              </tr>
            </thead>
            <tbody>
              {customers.data.data.map((c) => (
                <tr key={c.id}>
                  <td>
                    <Link to={`/customers/${c.id}`} className="link">
                      {c.name}
                    </Link>
                  </td>
                  <td>{c.phone}</td>
                  <td>{c.address ?? '—'}</td>
                  <td>{new Date(c.created_at).toLocaleDateString('en-IN')}</td>
                </tr>
              ))}
            </tbody>
          </table>

          <div className="pagination">
            <button className="btn btn-ghost" disabled={page <= 1} onClick={() => setPage((p) => p - 1)}>
              Previous
            </button>
            <span className="pagination-info">
              Page {page} of {totalPages} · {customers.data.pagination.total} total
            </span>
            <button className="btn btn-ghost" disabled={page >= totalPages} onClick={() => setPage((p) => p + 1)}>
              Next
            </button>
          </div>
        </Card>
      ) : (
        <Card>
          <EmptyState message="No customers found." />
        </Card>
      )}
    </div>
  )
}