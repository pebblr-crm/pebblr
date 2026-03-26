export interface Territory {
  id: string
  name: string
  teamId: string
  region?: string
  boundary?: Record<string, unknown>
  createdAt: string
  updatedAt: string
}

export interface CreateTerritoryInput {
  name: string
  teamId: string
  region?: string
  boundary?: Record<string, unknown>
}
