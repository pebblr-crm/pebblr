Scaffold a new Pebblr Web Component.

## Usage

```
/new-component <name>
```

Where `<name>` is the component name **without** the `pbl-` prefix (e.g., `lead-card` → `<pbl-lead-card>`).

## Instructions

Given the component name `$ARGUMENTS`, scaffold a new Web Component following Pebblr conventions:

1. **Determine paths:**
   - Component file: `web/src/components/pbl-$ARGUMENTS.ts`
   - Test file: `web/src/components/pbl-$ARGUMENTS.test.ts`

2. **Create the component file** at `web/src/components/pbl-$ARGUMENTS.ts`:

```typescript
/**
 * <pbl-$ARGUMENTS>
 *
 * [Brief description of what this component does]
 */
export class Pbl$PascalName extends HTMLElement {
  static readonly tag = 'pbl-$ARGUMENTS' as const;

  connectedCallback(): void {
    this.render();
  }

  private render(): void {
    this.innerHTML = `
      <div class="pbl-$ARGUMENTS">
        <!-- TODO: implement -->
      </div>
    `;
  }
}

customElements.define(Pbl$PascalName.tag, Pbl$PascalName);
```

Where `$PascalName` is the component name converted to PascalCase (e.g., `lead-card` → `LeadCard`).

3. **Create the test file** at `web/src/components/pbl-$ARGUMENTS.test.ts`:

```typescript
import './pbl-$ARGUMENTS';
import { Pbl$PascalName } from './pbl-$ARGUMENTS';

describe('pbl-$ARGUMENTS', () => {
  let el: Pbl$PascalName;

  beforeEach(() => {
    el = document.createElement(Pbl$PascalName.tag) as Pbl$PascalName;
    document.body.appendChild(el);
  });

  afterEach(() => {
    el.remove();
  });

  it('registers as a custom element', () => {
    expect(customElements.get(Pbl$PascalName.tag)).toBeDefined();
  });

  it('renders without errors', () => {
    expect(el).toBeInstanceOf(HTMLElement);
  });
});
```

4. **Conventions to follow:**
   - Tag name MUST start with `pbl-` — this is a hard requirement
   - Class name MUST be `Pbl` + PascalCase of the name (e.g., `PblLeadCard`)
   - Export a `static readonly tag` string constant for use in `customElements.define()`
   - Use `connectedCallback` for initialization (not constructors with DOM access)
   - Keep components self-contained — no framework imports, no global state
   - TypeScript strict mode is enforced — no implicit `any`

5. After creating both files, remind the user to:
   - Fill in the component implementation in `render()`
   - Add the component to `web/src/index.ts` if there's a barrel export
   - Run `make typecheck` to verify TypeScript is satisfied
