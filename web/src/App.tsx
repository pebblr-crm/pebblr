import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { createRouter, RouterProvider } from '@tanstack/react-router'
import { ToastProvider } from './components/Toast'
import { Route as rootRoute } from './routes/__root'
import { Route as indexRoute } from './routes/index'
import { Route as plannerRoute } from './routes/planner/index'
import { Route as plannerDailyRoute } from './routes/planner/daily'
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

export function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <ToastProvider>
        <RouterProvider router={router} />
      </ToastProvider>
    </QueryClientProvider>
  )
}
