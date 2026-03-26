import {
  useReactTable,
  getCoreRowModel,
  flexRender,
  type ColumnDef,
} from '@tanstack/react-table'

interface DataTableProps<TData> {
  // ColumnDef value type is intentionally any to support mixed accessor types
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  columns: ColumnDef<TData, any>[]
  data: TData[]
}

export function DataTable<TData>({ columns, data }: DataTableProps<TData>) {
  const table = useReactTable({
    data,
    columns,
    getCoreRowModel: getCoreRowModel(),
  })

  return (
    <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: 'var(--text-sm, 13px)' }}>
      <thead>
        {table.getHeaderGroups().map((headerGroup) => (
          <tr key={headerGroup.id}>
            {headerGroup.headers.map((header) => (
              <th
                key={header.id}
                style={{
                  textAlign: 'left',
                  padding: '8px 12px',
                  fontWeight: 'var(--font-semibold, 600)',
                  color: 'var(--color-text-secondary, #666)',
                  borderBottom: '1px solid var(--color-border, #e5e5e5)',
                  whiteSpace: 'nowrap',
                }}
              >
                {header.isPlaceholder
                  ? null
                  : flexRender(header.column.columnDef.header, header.getContext())}
              </th>
            ))}
          </tr>
        ))}
      </thead>
      <tbody>
        {table.getRowModel().rows.length > 0 ? (
          table.getRowModel().rows.map((row) => (
            <tr
              key={row.id}
              style={{ borderBottom: '1px solid var(--color-border, #e5e5e5)' }}
            >
              {row.getVisibleCells().map((cell) => (
                <td
                  key={cell.id}
                  style={{ padding: '8px 12px', color: 'var(--color-text-primary, #181818)' }}
                >
                  {flexRender(cell.column.columnDef.cell, cell.getContext())}
                </td>
              ))}
            </tr>
          ))
        ) : (
          <tr>
            <td
              colSpan={columns.length}
              style={{
                padding: '32px 12px',
                textAlign: 'center',
                color: 'var(--color-text-tertiary, #999)',
              }}
            >
              No results.
            </td>
          </tr>
        )}
      </tbody>
    </table>
  )
}
