import { createContext, useContext } from 'react'

export interface PlannerState {
  /** Monday of the currently viewed week (YYYY-MM-DD). */
  week: string | null
  /** Which planner view the user came from. */
  from: string | null
}

export interface PlannerContextValue {
  state: PlannerState
  setWeek: (week: string) => void
  setFrom: (from: string) => void
}

export const PlannerContext = createContext<PlannerContextValue>({
  state: { week: null, from: null },
  setWeek: () => {},
  setFrom: () => {},
})

export function usePlannerState() {
  return useContext(PlannerContext)
}
