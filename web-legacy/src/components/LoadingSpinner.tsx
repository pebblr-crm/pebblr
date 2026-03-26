interface LoadingSpinnerProps {
  size?: 'sm' | 'md' | 'lg'
  label?: string
}

const sizes = {
  sm: { width: 16, height: 16, borderWidth: 2 },
  md: { width: 24, height: 24, borderWidth: 2 },
  lg: { width: 40, height: 40, borderWidth: 3 },
} as const

export function LoadingSpinner({ size = 'md', label = 'Loading...' }: LoadingSpinnerProps) {
  const dims = sizes[size]
  return (
    <div
      role="status"
      aria-label={label}
      style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', padding: 16 }}
    >
      <div
        style={{
          width: dims.width,
          height: dims.height,
          borderTop: `${dims.borderWidth}px solid var(--color-accent, #1e40af)`,
          borderRight: `${dims.borderWidth}px solid var(--color-border, #e5e5e5)`,
          borderBottom: `${dims.borderWidth}px solid var(--color-border, #e5e5e5)`,
          borderLeft: `${dims.borderWidth}px solid var(--color-border, #e5e5e5)`,
          borderRadius: '50%',
          animation: 'spin 0.75s linear infinite',
        }}
      />
      <span
        style={{
          position: 'absolute',
          width: 1,
          height: 1,
          padding: 0,
          margin: -1,
          overflow: 'hidden',
          clip: 'rect(0,0,0,0)',
          whiteSpace: 'nowrap',
          border: 0,
        }}
      >
        {label}
      </span>
    </div>
  )
}
