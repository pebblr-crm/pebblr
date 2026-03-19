import { createRoute } from '@tanstack/react-router'
import { Route as rootRoute } from '../__root'
import { LoadingSpinner } from '../../components/LoadingSpinner'
import { useCustomer } from '../../services/customers'

export const Route = createRoute({
  getParentRoute: () => rootRoute,
  path: '/customers/$customerId',
  component: CustomerDetailPage,
})

function CustomerDetailPage() {
  const { customerId } = Route.useParams()
  const { data: customer, isLoading } = useCustomer(customerId)

  if (isLoading) {
    return <LoadingSpinner size="lg" label="Loading customer..." />
  }

  if (!customer) {
    return (
      <div className="page-body">
        <p style={{ color: 'var(--color-text-secondary)' }}>Customer not found.</p>
      </div>
    )
  }

  const addr = customer.address
  const addressLine = [addr?.street, addr?.city, addr?.state, addr?.zip, addr?.country]
    .filter(Boolean)
    .join(', ')

  return (
    <div>
      <div className="page-header">
        <h1 className="page-title">{customer.name}</h1>
      </div>
      <div className="page-body">
        <dl style={{ display: 'grid', gridTemplateColumns: 'max-content 1fr', gap: '8px 24px' }}>
          <dt style={{ color: 'var(--color-text-secondary)', fontWeight: 'var(--font-medium)' }}>Type</dt>
          <dd style={{ textTransform: 'capitalize' }}>{customer.type}</dd>
          <dt style={{ color: 'var(--color-text-secondary)', fontWeight: 'var(--font-medium)' }}>Email</dt>
          <dd>{customer.email || '—'}</dd>
          <dt style={{ color: 'var(--color-text-secondary)', fontWeight: 'var(--font-medium)' }}>Phone</dt>
          <dd>{customer.phone || '—'}</dd>
          <dt style={{ color: 'var(--color-text-secondary)', fontWeight: 'var(--font-medium)' }}>Address</dt>
          <dd>{addressLine || '—'}</dd>
          {customer.notes && (
            <>
              <dt style={{ color: 'var(--color-text-secondary)', fontWeight: 'var(--font-medium)' }}>Notes</dt>
              <dd>{customer.notes}</dd>
            </>
          )}
        </dl>
      </div>
    </div>
  )
}
