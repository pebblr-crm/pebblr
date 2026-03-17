import { createRoute } from '@tanstack/react-router'
import { createColumnHelper } from '@tanstack/react-table'
import { Route as rootRoute } from '../__root'
import { DataTable } from '../../components/DataTable'
import { LoadingSpinner } from '../../components/LoadingSpinner'
import { useLeads } from '../../services/leads'
import type { Lead } from '../../types/lead'

const columnHelper = createColumnHelper<Lead>()

const columns = [
  columnHelper.accessor('title', { header: 'Title' }),
  columnHelper.accessor('status', { header: 'Status' }),
  columnHelper.accessor('customerType', { header: 'Customer Type' }),
]

export const Route = createRoute({
  getParentRoute: () => rootRoute,
  path: '/leads',
  component: LeadsPage,
})

function LeadsPage() {
  const { data, isLoading } = useLeads()

  if (isLoading) {
    return <LoadingSpinner size="lg" label="Loading leads..." />
  }

  return (
    <div>
      <div className="page-header">
        <h1 className="page-title">Leads</h1>
      </div>
      <div className="page-body">
        <DataTable columns={columns} data={data?.items ?? []} />
      </div>
    </div>
  )
}
