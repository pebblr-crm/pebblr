import { type ReactNode } from 'react'
import { APIProvider, Map } from '@vis.gl/react-google-maps'

interface MapContainerProps {
  readonly center?: [number, number]
  readonly zoom?: number
  readonly className?: string
  readonly children?: ReactNode
}

const API_KEY = import.meta.env.VITE_GOOGLE_MAPS_API_KEY ?? ''

export function MapContainer({
  center = [26.1, 44.43], // Bucharest default [lng, lat]
  zoom = 11,
  className = '',
  children,
}: MapContainerProps) {
  return (
    <div className={`relative ${className}`}>
      <APIProvider apiKey={API_KEY}>
        <Map
          className="h-full w-full"
          defaultCenter={{ lat: center[1], lng: center[0] }}
          defaultZoom={zoom}
          gestureHandling="greedy"
          disableDefaultUI={false}
          mapId="pebblr-map"
        >
          {children}
        </Map>
      </APIProvider>
    </div>
  )
}
