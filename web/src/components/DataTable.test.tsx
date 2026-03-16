import { render, screen } from '@testing-library/react'
import { describe, it, expect } from 'vitest'
import { createColumnHelper } from '@tanstack/react-table'
import { DataTable } from './DataTable'

interface Row {
  name: string
  status: string
}

const columnHelper = createColumnHelper<Row>()

const columns = [
  columnHelper.accessor('name', { header: 'Name' }),
  columnHelper.accessor('status', { header: 'Status' }),
]

describe('DataTable', () => {
  it('renders column headers', () => {
    render(<DataTable columns={columns} data={[]} />)
    expect(screen.getByText('Name')).toBeInTheDocument()
    expect(screen.getByText('Status')).toBeInTheDocument()
  })

  it('renders rows from data', () => {
    const data: Row[] = [
      { name: 'Acme Corp', status: 'new' },
      { name: 'Globex', status: 'qualified' },
    ]
    render(<DataTable columns={columns} data={data} />)
    expect(screen.getByText('Acme Corp')).toBeInTheDocument()
    expect(screen.getByText('Globex')).toBeInTheDocument()
    expect(screen.getByText('new')).toBeInTheDocument()
    expect(screen.getByText('qualified')).toBeInTheDocument()
  })

  it('shows empty state when data is empty', () => {
    render(<DataTable columns={columns} data={[]} />)
    expect(screen.getByText('No results.')).toBeInTheDocument()
  })
})
