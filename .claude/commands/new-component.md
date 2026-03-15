Scaffold a new Pebblr React component.

## Usage

```
/new-component <name>
```

Where `<name>` is the component name in PascalCase (e.g., `LeadCard`, `RepBadge`).

## Instructions

Given the component name `$ARGUMENTS`, scaffold a new React component following Pebblr conventions:

1. **Determine paths:**
   - Component file: `web/src/components/$ARGUMENTS.tsx`
   - Test file: `web/src/components/$ARGUMENTS.test.tsx`

2. **Create the component file** at `web/src/components/$ARGUMENTS.tsx`:

```tsx
import React from 'react';

interface $ArgumentsProps {
  // TODO: define props
}

export function $Arguments({ }: $ArgumentsProps): React.ReactElement {
  return (
    <div className="$kebab-name">
      {/* TODO: implement */}
    </div>
  );
}
```

Where `$kebab-name` is the component name converted to kebab-case (e.g., `LeadCard` → `lead-card`).

3. **Create the test file** at `web/src/components/$ARGUMENTS.test.tsx`:

```tsx
import React from 'react';
import { render, screen } from '@testing-library/react';
import { describe, it, expect } from 'vitest';
import { $Arguments } from './$Arguments';

describe('$Arguments', () => {
  it('renders without errors', () => {
    render(<$Arguments />);
    // TODO: add assertions
  });
});
```

4. **Conventions to follow:**
   - Functional components only — no class components
   - Props interface named `${ComponentName}Props`
   - Strict TypeScript — no implicit `any`
   - Use TanStack Query (`useQuery`, `useMutation`) for any server state
   - Use TanStack Table for tabular data components
   - No global state — prefer props, React context, or TanStack Query cache
   - Tests use Vitest + React Testing Library

5. After creating both files, remind the user to:
   - Fill in the props interface and component implementation
   - Add the component to `web/src/components/index.ts` if there's a barrel export
   - Run `make typecheck` to verify TypeScript is satisfied
   - Run `bun test` to verify tests pass
