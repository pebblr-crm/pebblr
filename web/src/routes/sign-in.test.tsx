import { render, screen } from '@testing-library/react'
import { vi, describe, it, expect } from 'vitest'

let capturedComponent: React.ComponentType | null = null

vi.mock('@tanstack/react-router', () => ({
  createRoute: (opts: { component?: React.ComponentType }) => {
    if (opts?.component) capturedComponent = opts.component
    return {}
  },
}))

vi.mock('./__root', () => ({
  Route: {},
}))

vi.mock('@/components/ui/Button', () => ({
  Button: ({ children, ...props }: Record<string, unknown>) => (
    <button {...props}>{children as React.ReactNode}</button>
  ),
}))

await import('./sign-in')

describe('SignInPage', () => {
  it('renders the sign-in form', () => {
    const Component = capturedComponent!
    render(<Component />)

    expect(screen.getByText('Welcome back')).toBeInTheDocument()
    expect(screen.getByText('Sign in with Microsoft')).toBeInTheDocument()
    expect(screen.getByLabelText('Email')).toBeInTheDocument()
    expect(screen.getByLabelText('Password')).toBeInTheDocument()
  })

  it('renders Pebblr branding', () => {
    const Component = capturedComponent!
    render(<Component />)

    expect(screen.getByText('Pebblr')).toBeInTheDocument()
    expect(screen.getByText('Field Sales CRM')).toBeInTheDocument()
  })

  it('renders feature list', () => {
    const Component = capturedComponent!
    render(<Component />)

    expect(screen.getByText('Map-based visit planning')).toBeInTheDocument()
    expect(screen.getByText('Real-time coverage tracking')).toBeInTheDocument()
  })
})
