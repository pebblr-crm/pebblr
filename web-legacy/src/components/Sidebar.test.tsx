import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, it, expect, beforeEach } from 'vitest'
import {
  createRouter,
  RouterProvider,
  createRootRoute,
  createRoute,
  Outlet,
} from '@tanstack/react-router'
import i18n from '@/i18n'
import { Layout } from './Layout'

function makeTestRouter() {
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
    component: () => <div>test page</div>,
  })
  return createRouter({ routeTree: root.addChildren([index]) })
}

describe('Language switcher', () => {
  beforeEach(async () => {
    localStorage.removeItem('pebblr-language')
    await i18n.changeLanguage('en')
  })

  it('renders navigation in English by default', async () => {
    render(<RouterProvider router={makeTestRouter()} />)
    await waitFor(() => {
      expect(screen.getByText('Dashboard')).toBeInTheDocument()
      expect(screen.getByText('Targets')).toBeInTheDocument()
      expect(screen.getByText('Planner')).toBeInTheDocument()
      expect(screen.getByText('Team')).toBeInTheDocument()
    })
  })

  it('switches language when language switcher is clicked', async () => {
    const user = userEvent.setup()
    render(<RouterProvider router={makeTestRouter()} />)

    // Open settings popover
    await waitFor(() => {
      expect(screen.getByText('Settings')).toBeInTheDocument()
    })
    await user.click(screen.getByText('Settings'))

    // Click language switcher (shows "English" initially)
    const langButton = screen.getByTestId('language-switcher')
    expect(langButton).toHaveTextContent('English')
    await user.click(langButton)

    // Now should show Romanian labels
    await waitFor(() => {
      expect(screen.getByText('Panou')).toBeInTheDocument() // Dashboard in Romanian
      expect(screen.getByText('Conturi')).toBeInTheDocument() // Targets in Romanian
      expect(screen.getByText('Planificator')).toBeInTheDocument() // Planner in Romanian
      expect(screen.getByText('Echipă')).toBeInTheDocument() // Team in Romanian
    })
  })

  it('persists language selection in localStorage', async () => {
    const user = userEvent.setup()
    render(<RouterProvider router={makeTestRouter()} />)

    await waitFor(() => {
      expect(screen.getByText('Settings')).toBeInTheDocument()
    })
    await user.click(screen.getByText('Settings'))

    // Switch to Romanian
    await user.click(screen.getByTestId('language-switcher'))

    expect(localStorage.getItem('pebblr-language')).toBe('ro')
  })

  it('restores language from localStorage on init', async () => {
    // Set Romanian before rendering
    await i18n.changeLanguage('ro')

    render(<RouterProvider router={makeTestRouter()} />)

    await waitFor(() => {
      expect(screen.getByText('Panou')).toBeInTheDocument()
    })
  })

  it('cycles back to English after Romanian', async () => {
    const user = userEvent.setup()
    render(<RouterProvider router={makeTestRouter()} />)

    await waitFor(() => {
      expect(screen.getByText('Settings')).toBeInTheDocument()
    })
    await user.click(screen.getByText('Settings'))

    // Click once → Romanian
    await user.click(screen.getByTestId('language-switcher'))
    await waitFor(() => {
      expect(screen.getByText('Panou')).toBeInTheDocument()
    })

    // The language switcher should still be visible (popover stays open),
    // now showing "Română" label. Click again → English.
    await user.click(screen.getByTestId('language-switcher'))
    await waitFor(() => {
      expect(screen.getByText('Dashboard')).toBeInTheDocument()
    })
  })
})
