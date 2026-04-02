import { forwardRef, type InputHTMLAttributes } from 'react'

export const inputStyles =
  'w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-teal-500 focus:outline-none focus:ring-1 focus:ring-teal-500'

export const Input = forwardRef<HTMLInputElement, InputHTMLAttributes<HTMLInputElement>>(
  ({ className = '', ...props }, ref) => {
    return (
      <input
        ref={ref}
        className={`${inputStyles} ${className}`}
        {...props}
      />
    )
  },
)

Input.displayName = 'Input'
