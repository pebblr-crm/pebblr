import { createRoute, redirect } from '@tanstack/react-router'
import { Route as rootRoute } from './__root'

export const Route = createRoute({
  getParentRoute: () => rootRoute,
  path: '/',
  beforeLoad: () => {
    // Role-based redirect happens in the component since we need React context.
    // For now, redirect to planner as default.
    throw redirect({ to: '/planner' })
  },
})
