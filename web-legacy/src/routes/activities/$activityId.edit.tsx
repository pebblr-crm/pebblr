import { createRoute, redirect } from '@tanstack/react-router'
import { Route as rootRoute } from '../__root'

// The /edit route has been removed in favour of the inline-editable detail view.
// Redirect to the detail page to avoid broken bookmarks or external links.
export const Route = createRoute({
  getParentRoute: () => rootRoute,
  path: '/activities/$activityId/edit',
  beforeLoad: ({ params }) => {
    throw redirect({ to: '/activities/$activityId', params, replace: true })
  },
  component: () => null,
})
