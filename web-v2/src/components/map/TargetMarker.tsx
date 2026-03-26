import { useEffect, useRef } from 'react'
import maplibregl from 'maplibre-gl'

const priorityColors: Record<string, string> = {
  A: '#ef4444',
  B: '#f59e0b',
  C: '#94a3b8',
}

interface TargetMarkerProps {
  map: maplibregl.Map | null
  lng: number
  lat: number
  name: string
  priority?: string
  onClick?: () => void
}

export function TargetMarker({ map, lng, lat, name, priority = 'C', onClick }: TargetMarkerProps) {
  const markerRef = useRef<maplibregl.Marker | null>(null)

  useEffect(() => {
    if (!map) return

    const el = document.createElement('div')
    el.className = 'target-marker'
    el.style.width = '14px'
    el.style.height = '14px'
    el.style.borderRadius = '50%'
    el.style.backgroundColor = priorityColors[priority] ?? priorityColors.C
    el.style.border = '2px solid white'
    el.style.boxShadow = '0 1px 3px rgba(0,0,0,0.3)'
    el.style.cursor = 'pointer'
    el.title = name

    if (onClick) {
      el.addEventListener('click', onClick)
    }

    const marker = new maplibregl.Marker({ element: el }).setLngLat([lng, lat]).addTo(map)
    markerRef.current = marker

    return () => {
      marker.remove()
    }
  }, [map, lng, lat, name, priority, onClick])

  return null
}
