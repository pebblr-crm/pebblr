import { useEffect, useRef } from 'react'
import maplibregl from 'maplibre-gl'

const priorityColors: Record<string, string> = {
  a: '#ef4444',
  b: '#f59e0b',
  c: '#94a3b8',
}

interface TargetMarkerProps {
  map: maplibregl.Map | null
  lng: number
  lat: number
  name: string
  priority?: string
  selected?: boolean
  highlighted?: boolean
  onClick?: () => void
  onHover?: (hovered: boolean) => void
}

export function TargetMarker({
  map, lng, lat, name, priority = 'c',
  selected = false, highlighted = false,
  onClick, onHover,
}: TargetMarkerProps) {
  const markerRef = useRef<maplibregl.Marker | null>(null)
  const elRef = useRef<HTMLDivElement | null>(null)

  useEffect(() => {
    if (!map) return

    const el = document.createElement('div')
    el.className = 'target-marker'
    elRef.current = el
    el.title = name
    applyStyles(el, priority, selected, highlighted)

    if (onClick) el.addEventListener('click', onClick)
    if (onHover) {
      el.addEventListener('mouseenter', () => onHover(true))
      el.addEventListener('mouseleave', () => onHover(false))
    }

    const marker = new maplibregl.Marker({ element: el }).setLngLat([lng, lat]).addTo(map)
    markerRef.current = marker

    return () => { marker.remove() }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [map, lng, lat, name, onClick, onHover])

  // Update styles reactively without recreating the marker
  useEffect(() => {
    if (elRef.current) applyStyles(elRef.current, priority, selected, highlighted)
  }, [priority, selected, highlighted])

  return null
}

function applyStyles(el: HTMLDivElement, priority: string, selected: boolean, highlighted: boolean) {
  const size = highlighted || selected ? '18px' : '14px'
  el.style.width = size
  el.style.height = size
  el.style.borderRadius = '50%'
  el.style.backgroundColor = priorityColors[priority.toLowerCase()] ?? priorityColors.c
  el.style.border = selected ? '3px solid #0d9488' : '2px solid white'
  el.style.boxShadow = highlighted
    ? '0 0 0 4px rgba(13,148,136,0.3), 0 1px 3px rgba(0,0,0,0.3)'
    : '0 1px 3px rgba(0,0,0,0.3)'
  el.style.cursor = 'pointer'
  el.style.transition = 'width 0.15s, height 0.15s, border 0.15s, box-shadow 0.15s'
  el.style.zIndex = highlighted || selected ? '10' : '1'
}
