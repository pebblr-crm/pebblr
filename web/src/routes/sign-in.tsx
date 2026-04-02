import { createRoute } from '@tanstack/react-router'
import { Route as rootRoute } from './__root'
import { Button } from '@/components/ui/Button'

export const Route = createRoute({
  getParentRoute: () => rootRoute,
  path: '/sign-in',
  component: SignInPage,
})

function SignInPage() {
  return (
    <div className="flex min-h-screen">
      {/* Left: branding */}
      <div className="hidden w-1/2 flex-col justify-center bg-teal-600 p-12 text-white lg:flex">
        <div className="mb-8 flex items-center gap-3">
          <div className="h-12 w-12 rounded-xl bg-white/20 flex items-center justify-center text-xl font-bold">
            P
          </div>
          <span className="text-3xl font-bold">Pebblr</span>
        </div>
        <h2 className="text-2xl font-semibold">Field Sales CRM</h2>
        <p className="mt-4 text-lg text-teal-100 max-w-md">
          Plan visits, track coverage, and manage your field team — all in one place.
        </p>
        <div className="mt-8 space-y-4">
          {[
            'Map-based visit planning',
            'Real-time coverage tracking',
            'Mobile activity logging',
            'Team performance dashboards',
          ].map((feature) => (
            <div key={feature} className="flex items-center gap-3">
              <div className="h-2 w-2 rounded-full bg-teal-300" />
              <span className="text-teal-100">{feature}</span>
            </div>
          ))}
        </div>
      </div>

      {/* Right: login form */}
      <div className="flex flex-1 items-center justify-center p-8">
        <div className="w-full max-w-sm space-y-6">
          <div className="text-center">
            <h1 className="text-2xl font-bold text-slate-900">Welcome back</h1>
            <p className="mt-2 text-sm text-slate-500">Sign in with your organization account</p>
          </div>

          <Button className="w-full" size="lg" disabled>
            <svg className="h-5 w-5" viewBox="0 0 21 21" fill="none">
              <path d="M0 0h10v10H0z" fill="#F25022" />
              <path d="M11 0h10v10H11z" fill="#7FBA00" />
              <path d="M0 11h10v10H0z" fill="#00A4EF" />
              <path d="M11 11h10v10H11z" fill="#FFB900" />
            </svg>
            Sign in with Microsoft
          </Button>

          <div className="relative">
            <div className="absolute inset-0 flex items-center">
              <div className="w-full border-t border-slate-200" />
            </div>
            <div className="relative flex justify-center">
              <span className="bg-white px-3 text-xs text-slate-400">or</span>
            </div>
          </div>

          <form className="space-y-4" onSubmit={(e) => e.preventDefault()}>
            <div>
              <label htmlFor="sign-in-email" className="mb-1 block text-sm font-medium text-slate-700">Email</label>
              <input
                id="sign-in-email"
                type="email"
                placeholder="name@company.com"
                className="w-full rounded-lg border border-slate-300 px-3 py-2.5 text-sm focus:border-teal-500 focus:outline-none focus:ring-1 focus:ring-teal-500"
                disabled
              />
            </div>
            <div>
              <label htmlFor="sign-in-password" className="mb-1 block text-sm font-medium text-slate-700">Password</label>
              <input
                id="sign-in-password"
                type="password"
                placeholder="Enter your password"
                className="w-full rounded-lg border border-slate-300 px-3 py-2.5 text-sm focus:border-teal-500 focus:outline-none focus:ring-1 focus:ring-teal-500"
                disabled
              />
            </div>
            <Button className="w-full" size="lg" disabled>
              Sign in
            </Button>
          </form>

          <p className="text-center text-xs text-slate-400">
            Authentication is managed by your organization&apos;s Azure AD.
          </p>
        </div>
      </div>
    </div>
  )
}
