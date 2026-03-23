import type { UserActivityStats } from '@/types/dashboard'
import type { TenantConfig } from '@/types/config'

interface UserStatsTableProps {
  users: UserActivityStats[]
  config?: TenantConfig
}

function resolveStatusLabel(config: TenantConfig | undefined, statusKey: string): string {
  if (!config) return statusKey
  const st = config.activities.statuses.find((s) => s.key === statusKey)
  return st?.label ?? statusKey
}

export function UserStatsTable({ users, config }: UserStatsTableProps) {
  if (users.length === 0) {
    return (
      <div className="bg-surface-container-low p-6 rounded-xl text-center text-on-surface-variant">
        No activity data for this period.
      </div>
    )
  }

  // Collect all unique statuses across all users.
  const allStatuses = Array.from(
    new Set(users.flatMap((u) => u.byStatus.map((s) => s.status))),
  )

  return (
    <div className="bg-surface-container-low p-6 rounded-xl">
      <h3 className="text-xl font-bold text-primary font-headline mb-4">Rep Performance</h3>
      <div className="overflow-x-auto">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-slate-200">
              <th className="text-left py-2 pr-4 text-[10px] font-bold text-on-surface-variant uppercase">Rep</th>
              <th className="text-right py-2 px-3 text-[10px] font-bold text-on-surface-variant uppercase">Total</th>
              {allStatuses.map((status) => (
                <th key={status} className="text-right py-2 px-3 text-[10px] font-bold text-on-surface-variant uppercase">
                  {resolveStatusLabel(config, status)}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {users.map((user) => (
              <tr key={user.userId} className="border-b border-slate-100 hover:bg-surface-container-lowest transition-colors">
                <td className="py-3 pr-4 font-semibold text-on-surface">{user.userName || user.userId}</td>
                <td className="py-3 px-3 text-right font-extrabold text-primary">{user.total}</td>
                {allStatuses.map((status) => {
                  const count = user.byStatus.find((s) => s.status === status)?.count ?? 0
                  return (
                    <td key={status} className="py-3 px-3 text-right text-on-surface">
                      {count}
                    </td>
                  )
                })}
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}
