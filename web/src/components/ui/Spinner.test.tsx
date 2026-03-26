import { render, screen } from '@testing-library/react'
import { Spinner } from './Spinner'

describe('Spinner', () => {
  it('renders default label', () => {
    render(<Spinner />)
    expect(screen.getByText('Loading...')).toBeInTheDocument()
  })

  it('renders custom label', () => {
    render(<Spinner label="Fetching data..." />)
    expect(screen.getByText('Fetching data...')).toBeInTheDocument()
  })
})
