/** Target collection domain types — mirror the Go backend domain model. */

export interface Collection {
  id: string
  name: string
  creatorId: string
  teamId: string
  targetIds: string[]
  createdAt: string
  updatedAt: string
}

export interface CreateCollectionInput {
  name: string
  targetIds: string[]
}

export interface UpdateCollectionInput {
  id: string
  name: string
  targetIds: string[]
}
