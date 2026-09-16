import { Navigate, Route, Routes } from 'react-router-dom'
import { Layout } from './components/Layout'
import { useAuth } from './context/AuthContext'
import { LoginPage } from './pages/Login'
import { DashboardPage } from './pages/Dashboard'
import { JobsPage } from './pages/Jobs'
import { JobNewPage } from './pages/JobNew'
import { JobDetailPage } from './pages/JobDetail'
import { CustomersPage } from './pages/Customers'
import { CustomerDetailPage } from './pages/CustomerDetail'
import { TechniciansPage } from './pages/Technicians'
import { TechnicianDetailPage } from './pages/TechnicianDetail'
import { PaymentsPage } from './pages/Payments'
import { Spinner } from './components/ui'

function Protected() {
  const { user, loading } = useAuth()
  if (loading) return <Spinner label="Checking session…" />
  if (!user) return <Navigate to="/login" replace />
  if (user.role === 'TECHNICIAN') {
    return <Navigate to="/login" replace />
  }
  return <Layout />
}

export function AppRoutes() {
  return (
    <Routes>
      <Route path="/login" element={<PublicLogin />} />
      <Route element={<Protected />}>
        <Route path="/dashboard" element={<DashboardPage />} />
        <Route path="/jobs" element={<JobsPage />} />
        <Route path="/jobs/new" element={<JobNewPage />} />
        <Route path="/jobs/:id" element={<JobDetailPage />} />
        <Route path="/customers" element={<CustomersPage />} />
        <Route path="/customers/:id" element={<CustomerDetailPage />} />
        <Route path="/technicians" element={<TechniciansPage />} />
        <Route path="/technicians/:id" element={<TechnicianDetailPage />} />
        <Route path="/payments" element={<PaymentsPage />} />
        <Route path="/" element={<Navigate to="/dashboard" replace />} />
      </Route>
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  )
}

function PublicLogin() {
  const { user } = useAuth()
  if (user) return <Navigate to="/dashboard" replace />
  return <LoginPage />
}