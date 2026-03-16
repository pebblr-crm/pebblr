import { createRoute } from '@tanstack/react-router'
import { Route as rootRoute } from '../__root'
import { LoadingSpinner } from '../../components/LoadingSpinner'
import { useLead } from '../../services/leads'

export const Route = createRoute({
  getParentRoute: () => rootRoute,
  path: '/leads/$leadId',
  component: LeadDetailPage,
})

function LeadDetailPage() {
  const { leadId } = Route.useParams()
  const { data: lead, isLoading } = useLead(leadId)

  if (isLoading) {
    return <LoadingSpinner size="lg" label="Loading lead..." />
  }

  if (!lead) {
    return (
      <div className="page-body">
        <p style={{ color: 'var(--color-text-secondary)' }}>Lead not found.</p>
      </div>
    )
  }

  return (
    <div>
      <div className="page-header">
        <h1 className="page-title">{lead.companyName}</h1>
      </div>
      <div className="page-body">
        <dl style={{ display: 'grid', gridTemplateColumns: 'max-content 1fr', gap: '8px 24px' }}>
          <dt style={{ color: 'var(--color-text-secondary)', fontWeight: 'var(--font-medium)' }}>Contact</dt>
          <dd>{lead.contactName}</dd>
          <dt style={{ color: 'var(--color-text-secondary)', fontWeight: 'var(--font-medium)' }}>Email</dt>
          <dd>{lead.contactEmail}</dd>
          <dt style={{ color: 'var(--color-text-secondary)', fontWeight: 'var(--font-medium)' }}>Status</dt>
          <dd>{lead.status}</dd>
          <dt style={{ color: 'var(--color-text-secondary)', fontWeight: 'var(--font-medium)' }}>Notes</dt>
          <dd>{lead.notes}</dd>
        </dl>
      </div>
    </div>
  )
}
