import { type ReactNode } from 'react'

interface CardProps {
  readonly children: ReactNode
  readonly className?: string
}

export function Card({ children, className = '' }: CardProps) {
  return (
    <div className={`rounded-xl border border-slate-200 bg-white p-4 shadow-sm ${className}`}>
      {children}
    </div>
  )
}
