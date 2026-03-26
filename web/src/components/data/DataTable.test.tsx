import { render, screen, fireEvent } from '@testing-library/react'
import { createColumnHelper } from '@tanstack/react-table'
import { DataTable } from './DataTable'

interface TestRow {
  name: string
  value: number
}

const columnHelper = createColumnHelper<TestRow>()
const columns = [
  columnHelper.accessor('name', { header: 'Name' }),
  columnHelper.accessor('value', { header: 'Value' }),
]

function makeRows(count: number): TestRow[] {
  return Array.from({ length: count }, (_, i) => ({
    name: `Row ${i + 1}`,
    value: i + 1,
  }))
}

describe('DataTable', () => {
  it('renders headers and rows', () => {
    render(<DataTable data={makeRows(3)} columns={columns} />)
    expect(screen.getByText('Name')).toBeInTheDocument()
    expect(screen.getByText('Value')).toBeInTheDocument()
    expect(screen.getByText('Row 1')).toBeInTheDocument()
    expect(screen.getByText('Row 2')).toBeInTheDocument()
    expect(screen.getByText('Row 3')).toBeInTheDocument()
  })

  it('paginates when rows exceed pageSize', () => {
    render(<DataTable data={makeRows(30)} columns={columns} pageSize={10} />)
    // First page should have 10 rows
    expect(screen.getByText('Row 1')).toBeInTheDocument()
    expect(screen.getByText('Row 10')).toBeInTheDocument()
    expect(screen.queryByText('Row 11')).not.toBeInTheDocument()
    // Should show pagination
    expect(screen.getByText('Page 1 of 3')).toBeInTheDocument()
  })

  it('navigates between pages', () => {
    render(<DataTable data={makeRows(30)} columns={columns} pageSize={10} />)
    fireEvent.click(screen.getByText('Next'))
    expect(screen.getByText('Page 2 of 3')).toBeInTheDocument()
    expect(screen.getByText('Row 11')).toBeInTheDocument()
    expect(screen.queryByText('Row 1')).not.toBeInTheDocument()

    fireEvent.click(screen.getByText('Previous'))
    expect(screen.getByText('Page 1 of 3')).toBeInTheDocument()
    expect(screen.getByText('Row 1')).toBeInTheDocument()
  })

  it('disables Previous on first page and Next on last page', () => {
    render(<DataTable data={makeRows(10)} columns={columns} pageSize={5} />)
    expect(screen.getByText('Previous')).toBeDisabled()
    expect(screen.getByText('Next')).not.toBeDisabled()

    fireEvent.click(screen.getByText('Next'))
    expect(screen.getByText('Previous')).not.toBeDisabled()
    expect(screen.getByText('Next')).toBeDisabled()
  })

  it('does not show pagination when data fits one page', () => {
    render(<DataTable data={makeRows(3)} columns={columns} />)
    expect(screen.queryByText('Page')).not.toBeInTheDocument()
  })

  it('calls onRowClick when a row is clicked', () => {
    const onClick = vi.fn()
    render(<DataTable data={makeRows(3)} columns={columns} onRowClick={onClick} />)
    fireEvent.click(screen.getByText('Row 2'))
    expect(onClick).toHaveBeenCalledWith({ name: 'Row 2', value: 2 })
  })

  it('sorts by column when header is clicked', () => {
    const data = [
      { name: 'Charlie', value: 3 },
      { name: 'Alice', value: 1 },
      { name: 'Bob', value: 2 },
    ]
    render(<DataTable data={data} columns={columns} />)

    // Click Name header to sort ascending
    fireEvent.click(screen.getByText('Name'))
    const cells = screen.getAllByRole('cell')
    const names = cells.filter((_, i) => i % 2 === 0).map((c) => c.textContent)
    expect(names).toEqual(['Alice', 'Bob', 'Charlie'])
  })
})
