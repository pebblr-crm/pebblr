import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { createRouter, RouterProvider } from '@tanstack/react-router'
import { Route as rootRoute } from './routes/__root'
import { Route as indexRoute } from './routes/index'
import { Route as leadsIndexRoute } from './routes/leads/index'
import { Route as leadDetailRoute } from './routes/leads/$leadId'
import { Route as customersIndexRoute } from './routes/customers/index'
import { Route as customerDetailRoute } from './routes/customers/$customerId'
import { Route as calendarRoute } from './routes/calendar/index'
import { Route as teamRoute } from './routes/team/index'

const routeTree = rootRoute.addChildren([
  indexRoute,
  leadsIndexRoute,
  leadDetailRoute,
  customersIndexRoute,
  customerDetailRoute,
  calendarRoute,
  teamRoute,
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
