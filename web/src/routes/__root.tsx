import { createRootRoute, Outlet, useRouterState } from '@tanstack/react-router'
import { AppShell } from '@/layouts/AppShell'
import { ErrorBoundary } from '@/components/ui/ErrorBoundary'

export const Route = createRootRoute({
  component: RootLayout,
})

function RootLayout() {
  const { location } = useRouterState()
  const isSignIn = location.pathname === '/sign-in'

  if (isSignIn) {
    return <Outlet />
  }

  return (
    <AppShell currentPath={location.pathname}>
      <ErrorBoundary>
        <Outlet />
      </ErrorBoundary>
    </AppShell>
  )
}
