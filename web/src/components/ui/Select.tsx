import { forwardRef, type SelectHTMLAttributes } from 'react'

export const selectStyles =
  'w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-teal-500 focus:outline-none focus:ring-1 focus:ring-teal-500 bg-white'

export const Select = forwardRef<HTMLSelectElement, SelectHTMLAttributes<HTMLSelectElement>>(
  ({ className = '', children, ...props }, ref) => {
    return (
      <select
        ref={ref}
        className={`${selectStyles} ${className}`}
        {...props}
      >
        {children}
      </select>
    )
  },
)

Select.displayName = 'Select'
