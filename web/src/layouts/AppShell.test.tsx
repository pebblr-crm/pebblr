import { render, screen, fireEvent } from '@testing-library/react'
import { vi, describe, it, expect } from 'vitest'

vi.mock('./Sidebar', () => ({
  Sidebar: ({ currentPath, onNavigate }: { currentPath: string; onNavigate: () => void }) => (
    <div data-testid="sidebar" data-path={currentPath}>
      <button onClick={onNavigate}>nav-click</button>
    </div>
  ),
}))

vi.mock('lucide-react', () => ({
  Menu: () => <span data-testid="menu-icon" />,
}))

import { AppShell } from './AppShell'

describe('AppShell', () => {
  it('renders children and sidebar', () => {
    render(
      <AppShell currentPath="/dashboard">
        <div data-testid="child">Hello</div>
      </AppShell>,
    )
    expect(screen.getByTestId('child')).toBeInTheDocument()
    expect(screen.getByTestId('sidebar')).toBeInTheDocument()
  })

  it('passes currentPath to Sidebar', () => {
    render(
      <AppShell currentPath="/targets">
        <div>content</div>
      </AppShell>,
    )
    expect(screen.getByTestId('sidebar')).toHaveAttribute('data-path', '/targets')
  })

  it('opens mobile sidebar on menu button click', () => {
    render(
      <AppShell currentPath="/dashboard">
        <div>content</div>
      </AppShell>,
    )
    const menuButton = screen.getByLabelText('Open menu')
    fireEvent.click(menuButton)
    // After clicking, overlay should appear (Close sidebar button)
    expect(screen.getByLabelText('Close sidebar')).toBeInTheDocument()
  })

  it('closes mobile sidebar when overlay is clicked', () => {
    render(
      <AppShell currentPath="/dashboard">
        <div>content</div>
      </AppShell>,
    )
    fireEvent.click(screen.getByLabelText('Open menu'))
    expect(screen.getByLabelText('Close sidebar')).toBeInTheDocument()

    fireEvent.click(screen.getByLabelText('Close sidebar'))
    expect(screen.queryByLabelText('Close sidebar')).not.toBeInTheDocument()
  })
})
