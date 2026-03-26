import { useState, useCallback, useEffect, useMemo } from 'react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { createRouter, RouterProvider } from '@tanstack/react-router'
import { DemoAccountPicker } from './components/DemoAccountPicker'
import { ToastProvider } from './components/Toast'
import { PlannerContext, type PlannerState } from './contexts/planner'
import { ThemeContext, type Theme } from './contexts/theme'
import { isDemoMode, getCurrentUser, demoLogin, onAuthChange } from './services/auth'
import { Route as rootRoute } from './routes/__root'
import { Route as indexRoute } from './routes/index'
import { Route as plannerRoute } from './routes/planner/index'
import { Route as plannerDailyRoute } from './routes/planner/daily'
import { Route as plannerMapRoute } from './routes/planner/map'
import { Route as teamRoute } from './routes/team/index'
import { Route as targetsIndexRoute } from './routes/targets/index'
import { Route as targetDetailRoute } from './routes/targets/$targetId'
import { Route as newActivityRoute } from './routes/activities/new'
import { Route as activityDetailRoute } from './routes/activities/$activityId'
import { Route as editActivityRoute } from './routes/activities/$activityId.edit'

const routeTree = rootRoute.addChildren([
  indexRoute,
  plannerRoute,
  plannerDailyRoute,
  plannerMapRoute,
  teamRoute,
  targetsIndexRoute,
  targetDetailRoute,
  newActivityRoute,
  activityDetailRoute,
  editActivityRoute,
])

const router = createRouter({ routeTree })

declare module '@tanstack/react-router' {
  interface Register {
    router: typeof router
  }
}

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 30_000,
      retry: 1,
    },
  },
})

function getInitialTheme(): Theme {
  if (typeof globalThis.window === 'undefined') return 'light'
  const stored = localStorage.getItem('pebblr-theme')
  if (stored === 'dark' || stored === 'light') return stored
  return 'light'
}

export function App() {
  const [plannerState, setPlannerState] = useState<PlannerState>({ week: null, from: null })
  const setWeek = useCallback((week: string) => setPlannerState((s) => ({ ...s, week })), [])
  const setFrom = useCallback((from: string) => setPlannerState((s) => ({ ...s, from })), [])

  const [currentTheme, setThemeState] = useState<Theme>(getInitialTheme)
  const setTheme = useCallback((t: Theme) => {
    setThemeState(t)
    localStorage.setItem('pebblr-theme', t)
    document.documentElement.classList.toggle('dark', t === 'dark')
  }, [])
  const toggle = useCallback(() => setTheme(currentTheme === 'dark' ? 'light' : 'dark'), [currentTheme, setTheme])

  // Track whether a demo user is selected.
  const [authed, setAuthed] = useState(() => !isDemoMode() || getCurrentUser() !== null)

  useEffect(() => {
    onAuthChange(() => {
      setAuthed(getCurrentUser() !== null)
      // Clear TanStack Query cache on account switch so data reloads for the new user.
      queryClient.clear()
    })
  }, [])

  // Apply theme class on mount
  useEffect(() => {
    document.documentElement.classList.toggle('dark', currentTheme === 'dark')
  }, [currentTheme])

  const themeContextValue = useMemo(() => ({ theme: currentTheme, setTheme, toggle }), [currentTheme, setTheme, toggle])
  const plannerContextValue = useMemo(() => ({ state: plannerState, setWeek, setFrom }), [plannerState, setWeek, setFrom])

  // Demo mode: show account picker if no user selected.
  if (isDemoMode() && !authed) {
    return <DemoAccountPicker onSelect={demoLogin} />
  }

  return (
    <QueryClientProvider client={queryClient}>
      <ThemeContext.Provider value={themeContextValue}>
        <PlannerContext.Provider value={plannerContextValue}>
          <ToastProvider>
            <RouterProvider router={router} />
          </ToastProvider>
        </PlannerContext.Provider>
      </ThemeContext.Provider>
    </QueryClientProvider>
  )
}
