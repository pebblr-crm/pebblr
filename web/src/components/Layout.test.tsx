import { render, screen, waitFor } from '@testing-library/react'
import { describe, it, expect } from 'vitest'
import {
  createRouter,
  RouterProvider,
  createRootRoute,
  createRoute,
  Outlet,
} from '@tanstack/react-router'
import { Layout } from './Layout'

function makeTestRouter(content: React.ReactNode = null) {
  const root = createRootRoute({
    component: () => (
      <Layout>
        <Outlet />
      </Layout>
    ),
  })
  const index = createRoute({
    getParentRoute: () => root,
    path: '/',
    component: () => <>{content}</>,
  })
  return createRouter({ routeTree: root.addChildren([index]) })
}

describe('Layout', () => {
  it('renders sidebar logo', async () => {
    render(<RouterProvider router={makeTestRouter()} />)
    await waitFor(() => {
      expect(screen.getByText('Pebblr')).toBeInTheDocument()
    })
  })

  it('renders navigation links', async () => {
    render(<RouterProvider router={makeTestRouter()} />)
    await waitFor(() => {
      expect(screen.getByText('Dashboard')).toBeInTheDocument()
      expect(screen.getByText('Leads')).toBeInTheDocument()
      expect(screen.getByText('Calendar')).toBeInTheDocument()
      expect(screen.getByText('Team')).toBeInTheDocument()
    })
  })

  it('renders children in main content area', async () => {
    render(<RouterProvider router={makeTestRouter(<div data-testid="child">content</div>)} />)
    await waitFor(() => {
      expect(screen.getByTestId('child')).toBeInTheDocument()
    })
  })
})
