/** Maps activity status keys to Badge variant names. */
export const statusVariant: Record<string, 'primary' | 'success' | 'danger' | 'default'> = {
  planificat: 'primary',
  realizat: 'success',
  anulat: 'danger',
}

/** Maps activity status keys to dot color classes. */
export const statusDot: Record<string, string> = {
  realizat: 'bg-emerald-500',
  completed: 'bg-emerald-500',
  planificat: 'bg-blue-500',
  planned: 'bg-blue-500',
  anulat: 'bg-red-500',
  cancelled: 'bg-red-500',
}

/** Maps priority (a/b/c) to badge styling classes. */
export const priorityStyle: Record<string, string> = {
  a: 'bg-red-50 text-red-700 border border-red-100',
  b: 'bg-amber-50 text-amber-700 border border-amber-100',
  c: 'bg-slate-100 text-slate-600 border border-slate-200',
}

/** Maps priority (a/b/c) to dot color classes. */
export const priorityDot: Record<string, string> = {
  a: 'bg-red-500',
  b: 'bg-amber-500',
  c: 'bg-slate-400',
}

/** Maps priority (a/b/c) to human-readable labels. */
export const priorityLabel: Record<string, string> = {
  a: 'Priority A',
  b: 'Priority B',
  c: 'Priority C',
}

/** Maps status transition keys to button color classes. */
export const transitionColors: Record<string, string> = {
  realizat: 'bg-emerald-600 text-white hover:bg-emerald-700',
  anulat: 'bg-red-600 text-white hover:bg-red-700',
}

/** Returns className and label for a visit type badge. */
export function visitTypeBadge(visitType: string): { className: string; label: string } {
  if (visitType === 'f2f') {
    return {
      className: 'bg-amber-50 text-amber-700 border border-amber-200',
      label: 'In person',
    }
  }
  return {
    className: 'bg-blue-50 text-blue-700 border border-blue-200',
    label: 'Remote',
  }
}
