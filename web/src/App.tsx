import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { createRouter, RouterProvider } from '@tanstack/react-router'
import { Route as rootRoute } from './routes/__root'
import { Route as indexRoute } from './routes/index'
import { Route as calendarRoute } from './routes/calendar/index'
import { Route as teamRoute } from './routes/team/index'
import { Route as targetsIndexRoute } from './routes/targets/index'
import { Route as targetDetailRoute } from './routes/targets/$targetId'

const routeTree = rootRoute.addChildren([
  indexRoute,
  calendarRoute,
  teamRoute,
  targetsIndexRoute,
  targetDetailRoute,
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

export function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <RouterProvider router={router} />
    </QueryClientProvider>
  )
}
