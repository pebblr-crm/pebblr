import { useState, useEffect } from 'react'
import { LogIn, Shield, Users, UserCheck } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import type { Role } from '@/types/user'

interface DemoAccount {
  id: string
  name: string
  email: string
  role: Role
  avatar?: string
}

interface DemoAccountPickerProps {
  onSelect: (userId: string) => Promise<void>
}

const roleConfig: Record<Role, { tKey: string; icon: typeof Shield; color: string; bg: string }> = {
  admin: { tKey: 'demo.admin', icon: Shield, color: 'text-red-600', bg: 'bg-red-50' },
  manager: { tKey: 'demo.manager', icon: Users, color: 'text-amber-600', bg: 'bg-amber-50' },
  rep: { tKey: 'demo.rep', icon: UserCheck, color: 'text-blue-600', bg: 'bg-blue-50' },
}

function getInitials(name: string): string {
  return name
    .split(' ')
    .map((w) => w[0])
    .join('')
    .toUpperCase()
    .slice(0, 2)
}

export function DemoAccountPicker({ onSelect }: Readonly<DemoAccountPickerProps>) {
  const { t } = useTranslation()
  const [accounts, setAccounts] = useState<DemoAccount[]>([])
  const [loading, setLoading] = useState(true)
  const [selecting, setSelecting] = useState<string | null>(null)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    fetch('/demo/accounts')
      .then((r) => {
        if (!r.ok) throw new Error(`Failed to load accounts: ${r.status}`)
        return r.json() as Promise<DemoAccount[]>
      })
      .then(setAccounts)
      .catch((e: Error) => setError(e.message))
      .finally(() => setLoading(false))
  }, [])

  async function handleSelect(userId: string) {
    setSelecting(userId)
    setError(null)
    try {
      await onSelect(userId)
    } catch (e) {
      setError(e instanceof Error ? e.message : t('demo.failedToSignIn'))
      setSelecting(null)
    }
  }

  return (
    <div className="min-h-screen bg-surface flex items-center justify-center p-4">
      <div className="w-full max-w-md">
        <div className="text-center mb-8">
          <div className="text-3xl font-bold tracking-tight text-primary font-headline mb-2">
            Pebblr
          </div>
          <p className="text-on-surface-variant text-sm">
            {t('demo.chooseAccount')}
          </p>
        </div>

        {error && (
          <div className="mb-4 p-3 rounded-xl bg-error-container text-error text-sm text-center">
            {error}
          </div>
        )}

        {loading ? (
          <div className="flex justify-center py-12">
            <div className="w-8 h-8 border-3 border-primary/20 border-t-primary rounded-full animate-spin" />
          </div>
        ) : (
          <div className="space-y-3">
            {accounts.map((account) => {
              const config = roleConfig[account.role] ?? roleConfig.rep
              const Icon = config.icon
              const isSelecting = selecting === account.id

              return (
                <button
                  key={account.id}
                  onClick={() => handleSelect(account.id)}
                  disabled={selecting !== null}
                  className={`
                    w-full flex items-center gap-4 p-4 rounded-2xl border border-slate-100
                    bg-white hover:border-primary/30 hover:shadow-md
                    transition-all duration-200 group text-left
                    disabled:opacity-60 disabled:cursor-not-allowed
                    ${isSelecting ? 'border-primary/40 shadow-md' : ''}
                  `}
                >
                  {/* Avatar */}
                  <div className="w-12 h-12 rounded-full bg-primary-fixed flex items-center justify-center text-primary font-bold text-sm shrink-0">
                    {getInitials(account.name)}
                  </div>

                  {/* Info */}
                  <div className="flex-1 min-w-0">
                    <div className="font-headline font-semibold text-on-surface truncate">
                      {account.name}
                    </div>
                    <div className="text-sm text-on-surface-variant truncate">
                      {account.email}
                    </div>
                  </div>

                  {/* Role badge */}
                  <div className={`flex items-center gap-1.5 px-2.5 py-1 rounded-lg text-xs font-medium ${config.bg} ${config.color} shrink-0`}>
                    <Icon className="w-3.5 h-3.5" />
                    {t(config.tKey)}
                  </div>

                  {/* Arrow / spinner */}
                  <div className="shrink-0 w-5 h-5">
                    {isSelecting ? (
                      <div className="w-5 h-5 border-2 border-primary/20 border-t-primary rounded-full animate-spin" />
                    ) : (
                      <LogIn className="w-5 h-5 text-slate-300 group-hover:text-primary transition-colors" />
                    )}
                  </div>
                </button>
              )
            })}
          </div>
        )}

        <p className="text-center text-xs text-on-surface-variant mt-8">
          {t('demo.demoEnvironment')}
        </p>
      </div>
    </div>
  )
}
