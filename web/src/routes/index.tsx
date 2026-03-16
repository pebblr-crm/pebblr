import { createRoute } from '@tanstack/react-router'
import { Route as rootRoute } from './__root'

export const Route = createRoute({
  getParentRoute: () => rootRoute,
  path: '/',
  component: DashboardPage,
})

function DashboardPage() {
  return (
    <div>
      <div className="page-header">
        <h1 className="page-title">Dashboard</h1>
      </div>
      <div className="page-body">
        <p style={{ color: 'var(--color-text-secondary)' }}>Welcome to Pebblr CRM.</p>
      </div>
    </div>
  )
}
