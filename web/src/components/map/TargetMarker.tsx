import { AdvancedMarker } from '@vis.gl/react-google-maps'

const priorityColors: Record<string, string> = {
  a: '#ef4444',
  b: '#f59e0b',
  c: '#94a3b8',
}

interface TargetMarkerProps {
  readonly lng: number
  readonly lat: number
  readonly name: string
  readonly priority?: string
  readonly selected?: boolean
  readonly highlighted?: boolean
  readonly onClick?: () => void
  readonly onHover?: (hovered: boolean) => void
}

export function TargetMarker({
  lng, lat, name, priority = 'c',
  selected = false, highlighted = false,
  onClick, onHover,
}: TargetMarkerProps) {
  const size = highlighted || selected ? 18 : 14
  const color = priorityColors[priority.toLowerCase()] ?? priorityColors.c

  return (
    <AdvancedMarker
      position={{ lat, lng }}
      title={name}
      onClick={onClick}
    >
      <div
        role="img"
        aria-label={name}
        onMouseEnter={() => onHover?.(true)}
        onMouseLeave={() => onHover?.(false)}
        style={{
          width: size,
          height: size,
          borderRadius: '50%',
          backgroundColor: color,
          border: selected ? '3px solid #0d9488' : '2px solid white',
          boxShadow: highlighted
            ? '0 0 0 4px rgba(13,148,136,0.3), 0 1px 3px rgba(0,0,0,0.3)'
            : '0 1px 3px rgba(0,0,0,0.3)',
          cursor: 'pointer',
          transition: 'width 0.15s, height 0.15s, border 0.15s, box-shadow 0.15s',
          zIndex: highlighted || selected ? 10 : 1,
        }}
      />
    </AdvancedMarker>
  )
}
