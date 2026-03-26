import { useRef, useEffect, useState } from 'react'
import maplibregl from 'maplibre-gl'
import 'maplibre-gl/dist/maplibre-gl.css'

interface MapContainerProps {
  center?: [number, number]
  zoom?: number
  className?: string
  children?: (map: maplibregl.Map) => React.ReactNode
}

export function MapContainer({
  center = [26.1, 44.43], // Bucharest default
  zoom = 11,
  className = '',
  children,
}: MapContainerProps) {
  const containerRef = useRef<HTMLDivElement>(null)
  const mapObjRef = useRef<maplibregl.Map | null>(null)
  const [map, setMap] = useState<maplibregl.Map | null>(null)

  useEffect(() => {
    if (!containerRef.current || mapObjRef.current) return

    const instance = new maplibregl.Map({
      container: containerRef.current,
      style: {
        version: 8,
        sources: {
          osm: {
            type: 'raster',
            tiles: ['https://tile.openstreetmap.org/{z}/{x}/{y}.png'],
            tileSize: 256,
            attribution: '&copy; OpenStreetMap contributors',
          },
        },
        layers: [{ id: 'osm', type: 'raster', source: 'osm' }],
      },
      center,
      zoom,
    })

    instance.addControl(new maplibregl.NavigationControl(), 'top-right')

    instance.on('load', () => {
      mapObjRef.current = instance
      setMap(instance)
    })

    return () => {
      instance.remove()
      mapObjRef.current = null
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  return (
    <div className={`relative ${className}`}>
      <div ref={containerRef} className="h-full w-full" />
      {!map && (
        <div className="absolute inset-0 flex items-center justify-center bg-slate-100">
          <span className="text-sm text-slate-500">Loading map...</span>
        </div>
      )}
      {map && children?.(map)}
    </div>
  )
}
