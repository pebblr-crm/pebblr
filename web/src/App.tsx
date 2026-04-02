import { useState, useEffect, useCallback, type ReactNode } from 'react'
import { QueryClient, QueryClientProvider, MutationCache } from '@tanstack/react-query'
import { createRouter, RouterProvider } from '@tanstack/react-router'
import { AuthProvider } from '@/auth/provider'
import { useAuth } from '@/auth/context'
import { GlobalToast } from '@/components/ui/GlobalToast'
import { ErrorBoundary } from '@/components/ui/ErrorBoundary'
import { emitToast } from '@/lib/toast-store'
import { ApiError } from '@/types/api'
import '@/i18n'

import { Route as rootRoute } from '@/routes/__root'
import { Route as indexRoute } from '@/routes/index'
import { Route as plannerRoute } from '@/routes/planner'
import { Route as targetsRoute } from '@/routes/targets'
import { Route as activitiesRoute } from '@/routes/activities'
import { Route as dashboardRoute } from '@/routes/dashboard'
import { Route as coverageRoute } from '@/routes/coverage'
import { Route as consoleRoute } from '@/routes/console'
import { Route as auditRoute } from '@/routes/audit'
import { Route as targetDetailRoute } from '@/routes/targets.$id'
import { Route as repDrillDownRoute } from '@/routes/reps.$id'
import { Route as signInRoute } from '@/routes/sign-in'

const routeTree = rootRoute.addChildren([
  indexRoute,
  signInRoute,
  plannerRoute,
  targetsRoute,
  targetDetailRoute,
  activitiesRoute,
  dashboardRoute,
  repDrillDownRoute,
  coverageRoute,
  consoleRoute,
  auditRoute,
])

const router = createRouter({ routeTree })

declare module '@tanstack/react-router' {
  interface Register {
    router: typeof router
  }
}

const mutationCache = new MutationCache({
  onError: (error) => {
    if (error instanceof ApiError) {
      emitToast(error.message, 'error')
    } else if (error instanceof Error) {
      emitToast(error.message || 'An unexpected error occurred', 'error')
    } else {
      emitToast('An unexpected error occurred', 'error')
    }
  },
})

const queryClient = new QueryClient({
  mutationCache,
  defaultOptions: {
    queries: {
      staleTime: 30_000,
      retry: 1,
    },
  },
})

interface DemoAccount {
  id: string
  name: string
  email: string
  role: string
}

function DemoGate({ children }: { children: ReactNode }) {
  const { user, isDemoMode, demoLogin } = useAuth()
  const [accounts, setAccounts] = useState<DemoAccount[]>([])

  useEffect(() => {
    if (!isDemoMode) return
    fetch('/demo/accounts')
      .then((r) => {
        if (!r.ok) return []
        return r.json() as Promise<DemoAccount[]>
      })
      .then((data) => setAccounts(Array.isArray(data) ? data : []))
      .catch(() => {})
  }, [isDemoMode])

  const handleLogin = useCallback(
    async (userId: string) => {
      await demoLogin(userId)
      queryClient.removeQueries()
      await queryClient.invalidateQueries()
    },
    [demoLogin],
  )

  if (isDemoMode && !user) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-slate-50">
        <div className="w-80 rounded-xl border border-slate-200 bg-white p-6 shadow-sm">
          <h2 className="text-lg font-semibold text-slate-900">Pebblr v2 Demo</h2>
          <p className="mt-1 text-sm text-slate-500">Select a demo account to continue.</p>
          <div className="mt-4 space-y-2">
            {accounts.map((acct) => (
              <button
                key={acct.id}
                onClick={() => handleLogin(acct.id)}
                className="w-full rounded-lg border border-slate-200 px-4 py-2.5 text-left hover:bg-slate-50"
              >
                <div className="text-sm font-medium text-slate-900">{acct.name}</div>
                <div className="text-xs text-slate-500">{acct.role} &middot; {acct.email}</div>
              </button>
            ))}
            {accounts.length === 0 && (
              <p className="text-sm text-slate-400">Loading accounts...</p>
            )}
          </div>
        </div>
      </div>
    )
  }

  return <>{children}</>
}

export function App() {
  return (
    <ErrorBoundary>
      <AuthProvider>
        <QueryClientProvider client={queryClient}>
          <GlobalToast />
          <DemoGate>
            <RouterProvider router={router} />
          </DemoGate>
        </QueryClientProvider>
      </AuthProvider>
    </ErrorBoundary>
  )
}
