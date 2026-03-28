# CrateDesk — Frontend Design System & Component Library

**Stand:** 21. März 2026
**Version:** 1.0 (Implementierungsreif)
**Zielgruppe:** Senior Frontend Lead, Designer, PWA Specialist

---

## 1. DESIGN TOKENS (SCSS & TypeScript)

### 1.1 Farb-System

#### Primärpalette (Blue Brand)

```scss
// src/styles/tokens/colors/_primary.scss

// Light Mode
--color-primary-50:   #eff6ff;   hsl(210, 100%, 97%)
--color-primary-100:  #dbeafe;   hsl(210, 100%, 94%)
--color-primary-200:  #bfdbfe;   hsl(210, 96%, 89%)
--color-primary-300:  #93c5fd;   hsl(210, 94%, 80%)
--color-primary-400:  #60a5fa;   hsl(210, 94%, 71%)
--color-primary-500:  #3b82f6;   hsl(210, 97%, 61%)    // Primary Brand
--color-primary-600:  #2563eb;   hsl(210, 98%, 51%)
--color-primary-700:  #1d4ed8;   hsl(210, 97%, 45%)    // Interactive
--color-primary-800:  #1e40af;   hsl(210, 96%, 40%)
--color-primary-900:  #1e3a8a;   hsl(210, 96%, 34%)

// Dark Mode Overrides
[data-theme="dark"] {
  --color-primary-50:   #0f172a;   // Dunkelster Schatten
  --color-primary-500:  #60a5fa;   // Lighter für Dark Mode
  --color-primary-600:  #93c5fd;   // Für Hover im Dark
}
```

#### Semantische Farben

```scss
// src/styles/tokens/colors/_semantic.scss

// Success (Grün) — für Scanner OK, Check-In bestätigt
--color-success-50:   #f0fdf4;   hsl(120, 100%, 97%)
--color-success-100:  #dcfce7;   hsl(120, 100%, 93%)
--color-success-200:  #bbf7d0;   hsl(120, 91%, 88%)
--color-success-300:  #86efac;   hsl(120, 83%, 81%)
--color-success-400:  #4ade80;   hsl(120, 76%, 70%)
--color-success-500:  #22c55e;   hsl(120, 71%, 53%)    // Standard Success
--color-success-600:  #16a34a;   hsl(120, 70%, 43%)    // Scanner-OK (stark)
--color-success-700:  #15803d;   hsl(120, 71%, 36%)
--color-success-800:  #166534;   hsl(120, 62%, 33%)
--color-success-900:  #145231;   hsl(120, 61%, 28%)

// Warning (Orange) — für Scanner Yellow, Vorsicht
--color-warning-50:   #fffbeb;   hsl(44, 100%, 98%)
--color-warning-100:  #fef3c7;   hsl(48, 100%, 93%)
--color-warning-200:  #fde68a;   hsl(48, 95%, 88%)
--color-warning-300:  #fcd34d;   hsl(48, 96%, 79%)
--color-warning-400:  #fbbf24;   hsl(45, 98%, 70%)
--color-warning-500:  #f59e0b;   hsl(38, 92%, 61%)    // Standard Warning
--color-warning-600:  #d97706;   hsl(38, 92%, 50%)    // Scanner-Warn (stark)
--color-warning-700:  #b45309;   hsl(38, 86%, 43%)
--color-warning-800:  #92400e;   hsl(38, 80%, 36%)
--color-warning-900:  #78350f;   hsl(38, 75%, 29%)

// Error (Rot) — für Scanner ERROR, Defekt
--color-error-50:     #fef2f2;   hsl(0, 93%, 97%)
--color-error-100:    #fee2e2;   hsl(0, 93%, 94%)
--color-error-200:    #fecaca;   hsl(0, 93%, 90%)
--color-error-300:    #fca5a5;   hsl(0, 87%, 83%)
--color-error-400:    #f87171;   hsl(0, 84%, 75%)
--color-error-500:    #ef4444;   hsl(0, 84%, 60%)    // Standard Error
--color-error-600:    #dc2626;   hsl(0, 89%, 50%)    // Scanner-Error (stark)
--color-error-700:    #b91c1c;   hsl(0, 87%, 44%)
--color-error-800:    #991b1b;   hsl(0, 86%, 37%)
--color-error-900:    #7f1d1d;   hsl(0, 85%, 31%)

// Info (Cyan) — für Hinweise, Meldungen
--color-info-50:      #f0f9ff;   hsl(204, 100%, 97%)
--color-info-100:     #e0f2fe;   hsl(204, 100%, 93%)
--color-info-200:     #bae6fd;   hsl(204, 96%, 88%)
--color-info-300:     #7dd3fc;   hsl(204, 94%, 79%)
--color-info-400:     #38bdf8;   hsl(204, 96%, 69%)
--color-info-500:     #0ea5e9;   hsl(204, 100%, 50%)
--color-info-600:     #0284c7;   hsl(204, 97%, 43%)
--color-info-700:     #0369a1;   hsl(204, 96%, 38%)
```

#### Neutral Graustufen

```scss
// src/styles/tokens/colors/_neutral.scss

// Light Mode
--color-gray-50:      #f9fafb;   hsl(210, 13%, 98%)    // Backgrounds
--color-gray-100:     #f3f4f6;   hsl(210, 10%, 96%)
--color-gray-200:     #e5e7eb;   hsl(210, 7%, 90%)     // Borders, Dividers
--color-gray-300:     #d1d5db;   hsl(210, 7%, 82%)
--color-gray-400:     #9ca3af;   hsl(210, 8%, 63%)
--color-gray-500:     #6b7280;   hsl(210, 6%, 45%)     // Secondary Text
--color-gray-600:     #4b5563;   hsl(210, 8%, 35%)
--color-gray-700:     #374151;   hsl(210, 8%, 28%)     // Primary Text
--color-gray-800:     #1f2937;   hsl(210, 12%, 19%)
--color-gray-900:     #111827;   hsl(210, 14%, 11%)    // Dark Text

// Dark Mode (invertiert)
[data-theme="dark"] {
  --color-gray-50:    #0f172a;
  --color-gray-100:   #1e293b;
  --color-gray-200:   #334155;
  --color-gray-300:   #475569;
  --color-gray-400:   #64748b;
  --color-gray-500:   #94a3b8;
  --color-gray-600:   #cbd5e1;
  --color-gray-700:   #e2e8f0;
  --color-gray-800:   #f1f5f9;
  --color-gray-900:   #f8fafc;
}

// High Contrast Mode (für Lisa im Lager, schlechte Beleuchtung)
[data-theme="highcontrast"] {
  --color-gray-50:    #ffffff;
  --color-gray-900:   #000000;
  --color-primary-500: #0000ee;   // Reines Blau
  --color-success-600: #008000;   // Reines Grün
  --color-error-600:  #ff0000;    // Reines Rot
  --color-warning-600: #ff8800;   // Orange
}
```

#### Secondary & Accent Farben

```scss
// src/styles/tokens/colors/_secondary.scss

// Secondary (Lila) — für alternative Aktionen
--color-secondary-500: #a855f7;   hsl(280, 98%, 54%)
--color-secondary-600: #9333ea;   hsl(280, 98%, 46%)
--color-secondary-700: #7e22ce;   hsl(280, 97%, 40%)

// Accent (Indigo) — für Focus-States
--color-accent-500: #6366f1;      hsl(262, 100%, 54%)
--color-accent-600: #4f46e5;      hsl(263, 100%, 48%)

// Zweitfarbe für Equipment-Kategorien
--color-category-tone:    #ec4899;  // Pink — Ton
--color-category-light:   #f59e0b;  // Amber — Licht
--color-category-video:   #8b5cf6;  // Violet — Video
--color-category-rigging: #6366f1;  // Indigo — Rigging
--color-category-cable:   #10b981;  // Emerald — Kabel
--color-category-misc:    #6b7280;  // Gray — Sonstiges
```

### 1.2 Typografie-Skala

```scss
// src/styles/tokens/typography/_fonts.scss

// Font Stack (System-First)
--font-family: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', sans-serif;
--font-family-mono: 'JetBrains Mono', 'Courier New', monospace;

// Font Weights
--font-weight-400: 400;  // Regular
--font-weight-500: 500;  // Medium
--font-weight-600: 600;  // Semibold
--font-weight-700: 700;  // Bold

// Größen & Line-Heights (12px – 48px Skala)
--font-size-xs:    12px; --line-height-xs:   1.5;  // 18px
--font-size-sm:    14px; --line-height-sm:   1.43; // 20px
--font-size-base:  16px; --line-height-base: 1.5;  // 24px
--font-size-lg:    18px; --line-height-lg:   1.56; // 28px
--font-size-xl:    20px; --line-height-xl:   1.4;  // 28px
--font-size-2xl:   24px; --line-height-2xl:  1.33; // 32px
--font-size-3xl:   30px; --line-height-3xl:  1.2;  // 36px
--font-size-4xl:   36px; --line-height-4xl:  1.11; // 40px
--font-size-5xl:   48px; --line-height-5xl:  1;    // 48px

// Typo-Styles für Komponenten
--typo-body-lg:     { font-size: var(--font-size-lg); font-weight: 400; line-height: var(--line-height-lg); }
--typo-body-base:   { font-size: var(--font-size-base); font-weight: 400; line-height: var(--line-height-base); }
--typo-body-sm:     { font-size: var(--font-size-sm); font-weight: 400; line-height: var(--line-height-sm); }
--typo-label:       { font-size: var(--font-size-sm); font-weight: 600; line-height: var(--line-height-sm); }
--typo-heading-h1:  { font-size: var(--font-size-3xl); font-weight: 700; line-height: var(--line-height-3xl); }
--typo-heading-h2:  { font-size: var(--font-size-2xl); font-weight: 700; line-height: var(--line-height-2xl); }
--typo-heading-h3:  { font-size: var(--font-size-xl); font-weight: 700; line-height: var(--line-height-xl); }
--typo-heading-h4:  { font-size: var(--font-size-lg); font-weight: 600; line-height: var(--line-height-lg); }

// Letter Spacing
--letter-spacing-tight: -0.02em;
--letter-spacing-normal: 0;
--letter-spacing-wide:   0.025em;
```

### 1.3 Spacing-Skala (4px Grid)

```scss
// src/styles/tokens/spacing/_spacing.scss

--space-0:    0;      // 0px
--space-1:    4px;    // 1 unit
--space-2:    8px;    // 2 units
--space-3:    12px;   // 3 units
--space-4:    16px;   // 4 units (base)
--space-5:    20px;   // 5 units
--space-6:    24px;   // 6 units
--space-8:    32px;   // 8 units
--space-10:   40px;   // 10 units
--space-12:   48px;   // 12 units
--space-16:   64px;   // 16 units
--space-20:   80px;   // 20 units
--space-24:   96px;   // 24 units

// Padding-Presets (Button, Card, Input)
--padding-xs:  var(--space-2) var(--space-3);
--padding-sm:  var(--space-2) var(--space-4);
--padding-md:  var(--space-3) var(--space-4);
--padding-lg:  var(--space-4) var(--space-6);
--padding-xl:  var(--space-6) var(--space-8);

// Gap-Presets (Flex-Layouts)
--gap-2: var(--space-2);
--gap-3: var(--space-3);
--gap-4: var(--space-4);
--gap-6: var(--space-6);
--gap-8: var(--space-8);
```

### 1.4 Border-Radius

```scss
// src/styles/tokens/border/_radius.scss

--radius-none:  0;
--radius-sm:    4px;   // Input, Button
--radius-md:    8px;   // Card, Modal
--radius-lg:    12px;  // Large Container
--radius-xl:    16px;  // Hero-Element
--radius-full:  9999px; // Pill-Button, Avatar
```

### 1.5 Shadows

```scss
// src/styles/tokens/shadow/_shadow.scss

// Light Mode
--shadow-none:  none;
--shadow-xs:    0 1px 2px 0 rgb(0 0 0 / 0.05);
--shadow-sm:    0 1px 3px 0 rgb(0 0 0 / 0.1), 0 1px 2px -1px rgb(0 0 0 / 0.1);
--shadow-md:    0 4px 6px -1px rgb(0 0 0 / 0.1), 0 2px 4px -2px rgb(0 0 0 / 0.1);
--shadow-lg:    0 10px 15px -3px rgb(0 0 0 / 0.1), 0 4px 6px -4px rgb(0 0 0 / 0.1);
--shadow-xl:    0 20px 25px -5px rgb(0 0 0 / 0.1), 0 8px 10px -6px rgb(0 0 0 / 0.1);
--shadow-2xl:   0 25px 50px -12px rgb(0 0 0 / 0.25);

// Dark Mode (stärker, da Kontrast-Bedarf)
[data-theme="dark"] {
  --shadow-md: 0 4px 6px -1px rgb(0 0 0 / 0.3), 0 2px 4px -2px rgb(0 0 0 / 0.3);
  --shadow-lg: 0 10px 15px -3px rgb(0 0 0 / 0.4), 0 4px 6px -4px rgb(0 0 0 / 0.4);
}

// Modal/Overlay (für Modal-Hintergrund)
--shadow-overlay: 0 0 0 99999px rgb(0 0 0 / 0.5);
```

### 1.6 Breakpoints

```scss
// src/styles/tokens/responsive/_breakpoints.scss

--breakpoint-xs:   0;      // Mobile Portrait
--breakpoint-sm:   640px;  // Mobile Landscape
--breakpoint-md:   768px;  // Tablet
--breakpoint-lg:   1024px; // Desktop
--breakpoint-xl:   1280px; // Desktop Large
--breakpoint-2xl:  1536px; // Desktop XL

// Mixin-Helper
@mixin sm  { @media (min-width: 640px)  { @content; } }
@mixin md  { @media (min-width: 768px)  { @content; } }
@mixin lg  { @media (min-width: 1024px) { @content; } }
@mixin xl  { @media (min-width: 1280px) { @content; } }
@mixin 2xl { @media (min-width: 1536px) { @content; } }

// Container Queries (für Komponenten-Responsive)
--container-xs:  20rem;   // 320px
--container-sm:  24rem;   // 384px
--container-md:  28rem;   // 448px
--container-lg:  32rem;   // 512px
--container-xl:  36rem;   // 576px
--container-2xl: 42rem;   // 672px
```

### 1.7 Z-Index Skala

```scss
// src/styles/tokens/layout/_z-index.scss

--z-dropdown:     100;   // Dropdown-Menü
--z-sticky:       200;   // Sticky Header/Sidebar
--z-modal-bg:     300;   // Modal Overlay
--z-modal:        310;   // Modal Dialog
--z-popover:      400;   // Tooltip, Popover
--z-notification: 500;   // Toast/Alert oben rechts
--z-scanner:      600;   // Scanner-Overlay (ganz oben)
```

### 1.8 Transition & Animation

```scss
// src/styles/tokens/animation/_transitions.scss

// Dauer (Easing nach Material Design)
--duration-75:   75ms;
--duration-100:  100ms;
--duration-150:  150ms;
--duration-200:  200ms;
--duration-300:  300ms;
--duration-500:  500ms;
--duration-700:  700ms;
--duration-1000: 1000ms;

// Easing Functions
--ease-in:       cubic-bezier(0.4, 0, 1, 1);
--ease-out:      cubic-bezier(0, 0, 0.2, 1);
--ease-in-out:   cubic-bezier(0.4, 0, 0.2, 1);
--ease-elastic:  cubic-bezier(0.34, 1.56, 0.64, 1);

// Transition-Defaults
--transition-colors:   color var(--duration-200) var(--ease-out), background-color var(--duration-200) var(--ease-out);
--transition-opacity:  opacity var(--duration-200) var(--ease-out);
--transition-transform: transform var(--duration-200) var(--ease-out);
--transition-all:      all var(--duration-200) var(--ease-out);

// Scanner-Feedback Animations
--animation-scan-success: fadeInScale 0.3s var(--ease-elastic);
--animation-scan-error:   shake 0.4s var(--ease-in-out);
--animation-scan-warn:    pulse 0.6s var(--ease-in-out);

@keyframes fadeInScale {
  from { opacity: 0; transform: scale(0.8); }
  to   { opacity: 1; transform: scale(1); }
}

@keyframes shake {
  0%, 100% { transform: translateX(0); }
  25%  { transform: translateX(-8px); }
  75%  { transform: translateX(8px); }
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50%  { opacity: 0.5; }
}
```

### 1.9 TypeScript Constants (Mirrors SCSS)

```typescript
// src/styles/tokens/tokens.ts

export const tokens = {
  colors: {
    primary: {
      50: '#eff6ff',
      100: '#dbeafe',
      500: '#3b82f6',
      600: '#2563eb',
      700: '#1d4ed8',
      900: '#1e3a8a',
    },
    success: {
      500: '#22c55e',
      600: '#16a34a',
    },
    error: {
      500: '#ef4444',
      600: '#dc2626',
    },
    warning: {
      500: '#f59e0b',
      600: '#d97706',
    },
    gray: {
      50: '#f9fafb',
      100: '#f3f4f6',
      200: '#e5e7eb',
      500: '#6b7280',
      700: '#374151',
      900: '#111827',
    },
  },
  fontSize: {
    xs: '12px',
    sm: '14px',
    base: '16px',
    lg: '18px',
    xl: '20px',
    '2xl': '24px',
  },
  spacing: {
    1: '4px',
    2: '8px',
    3: '12px',
    4: '16px',
    6: '24px',
    8: '32px',
    12: '48px',
    16: '64px',
    24: '96px',
  },
  borderRadius: {
    sm: '4px',
    md: '8px',
    lg: '12px',
    xl: '16px',
    full: '9999px',
  },
  breakpoints: {
    xs: 0,
    sm: 640,
    md: 768,
    lg: 1024,
    xl: 1280,
    '2xl': 1536,
  },
  duration: {
    75: '75ms',
    100: '100ms',
    150: '150ms',
    200: '200ms',
    300: '300ms',
  },
} as const;

// Type-safe color accessor
export type ColorValue = typeof tokens.colors[keyof typeof tokens.colors][keyof typeof tokens.colors[keyof typeof tokens.colors]];
```

---

## 2. COMPONENT LIBRARY (Atomic Design)

### 2.1 Atoms (18+ primitiv)

#### Button

```typescript
// src/components/atoms/Button/Button.tsx

interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: 'primary' | 'secondary' | 'ghost' | 'danger' | 'success';
  size?: 'xs' | 'sm' | 'md' | 'lg' | 'xl';
  loading?: boolean;
  disabled?: boolean;
  isIconOnly?: boolean;
  children: React.ReactNode;
}

export const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(
  ({ variant = 'primary', size = 'md', loading, disabled, isIconOnly, ...props }, ref) => {
    return (
      <button
        ref={ref}
        className={cn(
          'button',
          `button--${variant}`,
          `button--${size}`,
          { 'button--disabled': disabled || loading },
          { 'button--loading': loading }
        )}
        disabled={disabled || loading}
        {...props}
      />
    );
  }
);

Button.displayName = 'Button';
```

```scss
// src/components/atoms/Button/Button.module.scss

.button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: var(--space-2);
  font-weight: var(--font-weight-600);
  border: 1px solid transparent;
  border-radius: var(--radius-sm);
  cursor: pointer;
  transition: var(--transition-all);
  user-select: none;

  &:focus-visible {
    outline: 2px solid var(--color-accent-500);
    outline-offset: 2px;
  }

  &--primary {
    background-color: var(--color-primary-600);
    color: white;

    &:hover:not(:disabled) {
      background-color: var(--color-primary-700);
      box-shadow: var(--shadow-md);
    }

    &:active:not(:disabled) {
      background-color: var(--color-primary-800);
    }
  }

  &--secondary {
    background-color: var(--color-gray-200);
    color: var(--color-gray-900);

    &:hover:not(:disabled) {
      background-color: var(--color-gray-300);
    }
  }

  &--ghost {
    background-color: transparent;
    color: var(--color-primary-600);
    border-color: var(--color-gray-300);

    &:hover:not(:disabled) {
      background-color: var(--color-primary-50);
      border-color: var(--color-primary-300);
    }
  }

  &--danger {
    background-color: var(--color-error-600);
    color: white;

    &:hover:not(:disabled) {
      background-color: var(--color-error-700);
    }
  }

  &--success {
    background-color: var(--color-success-600);
    color: white;

    &:hover:not(:disabled) {
      background-color: var(--color-success-700);
    }
  }

  &--xs {
    padding: var(--space-1) var(--space-2);
    font-size: var(--font-size-xs);
  }

  &--sm {
    padding: var(--space-2) var(--space-3);
    font-size: var(--font-size-sm);
    min-height: 32px;
  }

  &--md {
    padding: var(--space-2) var(--space-4);
    font-size: var(--font-size-base);
    min-height: 40px;
  }

  &--lg {
    padding: var(--space-3) var(--space-6);
    font-size: var(--font-size-lg);
    min-height: 48px;
  }

  &--xl {
    padding: var(--space-4) var(--space-8);
    font-size: var(--font-size-lg);
    min-height: 56px;
  }

  &--disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  &--loading {
    pointer-events: none;
  }

  // Handschuh-freundliche große Targets für Zebra TC21
  @media (max-width: 600px) {
    min-height: 48px;
    min-width: 48px;
    padding: var(--space-3) var(--space-4);
  }
}

.iconOnly {
  width: 40px;
  height: 40px;
  padding: 0;

  @media (max-width: 600px) {
    width: 56px;
    height: 56px;
  }
}
```

#### Input

```typescript
// src/components/atoms/Input/Input.tsx

interface InputProps extends React.InputHTMLAttributes<HTMLInputElement> {
  label?: string;
  error?: string;
  hint?: string;
  required?: boolean;
  variant?: 'default' | 'error';
}

export const Input = React.forwardRef<HTMLInputElement, InputProps>(
  ({ label, error, hint, required, variant = 'default', ...props }, ref) => {
    const id = React.useId();

    return (
      <div className="input-wrapper">
        {label && (
          <label htmlFor={id} className={cn('input-label', { 'input-label--required': required })}>
            {label}
          </label>
        )}
        <input
          ref={ref}
          id={id}
          className={cn('input', `input--${variant}`)}
          aria-invalid={!!error}
          aria-describedby={error ? `${id}-error` : hint ? `${id}-hint` : undefined}
          {...props}
        />
        {error && (
          <span id={`${id}-error`} className="input-error">
            {error}
          </span>
        )}
        {hint && !error && (
          <span id={`${id}-hint`} className="input-hint">
            {hint}
          </span>
        )}
      </div>
    );
  }
);

Input.displayName = 'Input';
```

```scss
// src/components/atoms/Input/Input.module.scss

.input {
  width: 100%;
  padding: var(--space-2) var(--space-3);
  font-size: var(--font-size-base);
  font-family: var(--font-family);
  border: 1px solid var(--color-gray-300);
  border-radius: var(--radius-sm);
  background-color: white;
  color: var(--color-gray-900);
  transition: var(--transition-all);

  &:focus {
    outline: none;
    border-color: var(--color-primary-500);
    box-shadow: 0 0 0 3px var(--color-primary-100);
  }

  &:disabled {
    background-color: var(--color-gray-100);
    color: var(--color-gray-500);
    cursor: not-allowed;
  }

  &--error {
    border-color: var(--color-error-600);

    &:focus {
      box-shadow: 0 0 0 3px var(--color-error-100);
    }
  }

  [data-theme="dark"] & {
    background-color: var(--color-gray-800);
    color: var(--color-gray-100);
    border-color: var(--color-gray-700);
  }

  // Zebra TC21 größer
  @media (max-width: 600px) {
    padding: var(--space-3) var(--space-4);
    font-size: var(--font-size-lg);
    min-height: 48px;
  }
}

.inputLabel {
  display: block;
  margin-bottom: var(--space-1);
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-600);
  color: var(--color-gray-700);

  &--required::after {
    content: ' *';
    color: var(--color-error-600);
  }
}

.inputError {
  display: block;
  margin-top: var(--space-1);
  font-size: var(--font-size-xs);
  color: var(--color-error-600);
  font-weight: var(--font-weight-500);
}

.inputHint {
  display: block;
  margin-top: var(--space-1);
  font-size: var(--font-size-xs);
  color: var(--color-gray-500);
}
```

#### Weitere Atoms (kurz)

```typescript
// Badge, Tag, Avatar, Icon, Checkbox, Radio, Toggle, Select, TextArea
// Skeleton, Spinner, ProgressBar, Divider, Link, StatusBadge

// Beispiel: Badge
interface BadgeProps {
  variant?: 'default' | 'success' | 'warning' | 'error' | 'info';
  size?: 'sm' | 'md' | 'lg';
  dot?: boolean;
  children: React.ReactNode;
}

export const Badge: React.FC<BadgeProps> = ({ variant = 'default', size = 'md', dot, children }) => (
  <span className={cn('badge', `badge--${variant}`, `badge--${size}`, { 'badge--dot': dot })}>
    {dot && <span className="badge-dot" />}
    {children}
  </span>
);
```

### 2.2 Molecules (12+)

#### FormField (Label + Input + Error)

```typescript
// src/components/molecules/FormField/FormField.tsx

interface FormFieldProps {
  label: string;
  required?: boolean;
  error?: string;
  hint?: string;
  children: React.ReactNode;
}

export const FormField: React.FC<FormFieldProps> = ({ label, required, error, hint, children }) => (
  <div className="form-field">
    <label className={cn('form-field-label', { 'form-field-label--required': required })}>
      {label}
    </label>
    {children}
    {error && <span className="form-field-error">{error}</span>}
    {hint && !error && <span className="form-field-hint">{hint}</span>}
  </div>
);
```

#### SearchBar mit Debounce

```typescript
// src/components/molecules/SearchBar/SearchBar.tsx

interface SearchBarProps {
  placeholder?: string;
  onSearch: (value: string) => void;
  debounceMs?: number;
  disabled?: boolean;
}

export const SearchBar: React.FC<SearchBarProps> = ({ placeholder, onSearch, debounceMs = 300, disabled }) => {
  const [value, setValue] = React.useState('');
  const debounceRef = React.useRef<ReturnType<typeof setTimeout>>();

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const newValue = e.target.value;
    setValue(newValue);

    clearTimeout(debounceRef.current);
    debounceRef.current = setTimeout(() => {
      onSearch(newValue);
    }, debounceMs);
  };

  return (
    <div className="search-bar">
      <Icon name="search" className="search-bar-icon" />
      <input
        type="text"
        placeholder={placeholder}
        value={value}
        onChange={handleChange}
        disabled={disabled}
        className="search-bar-input"
      />
      {value && (
        <button onClick={() => { setValue(''); onSearch(''); }} className="search-bar-clear">
          <Icon name="x" />
        </button>
      )}
    </div>
  );
};
```

#### DatePicker & DateRangePicker

```typescript
// src/components/molecules/DatePicker/DatePicker.tsx

interface DatePickerProps {
  value: Date | null;
  onChange: (date: Date | null) => void;
  disabled?: boolean;
  minDate?: Date;
  maxDate?: Date;
}

export const DatePicker: React.FC<DatePickerProps> = ({ value, onChange, disabled, minDate, maxDate }) => {
  // Verwendet native HTML5 date input auf Mobile
  // Radix Popover + Calendar auf Desktop
  return (
    <input
      type="date"
      value={value ? value.toISOString().split('T')[0] : ''}
      onChange={(e) => onChange(e.target.value ? new Date(e.target.value) : null)}
      disabled={disabled}
      min={minDate?.toISOString().split('T')[0]}
      max={maxDate?.toISOString().split('T')[0]}
      className="date-picker"
    />
  );
};
```

#### Alert & Toast (Notification)

```typescript
// src/components/molecules/Alert/Alert.tsx

interface AlertProps {
  variant?: 'info' | 'success' | 'warning' | 'error';
  title?: string;
  children: React.ReactNode;
  onClose?: () => void;
}

export const Alert: React.FC<AlertProps> = ({ variant = 'info', title, children, onClose }) => (
  <div className={cn('alert', `alert--${variant}`)} role="alert">
    <Icon name={alertIconMap[variant]} className="alert-icon" />
    <div className="alert-content">
      {title && <h4 className="alert-title">{title}</h4>}
      {children}
    </div>
    {onClose && (
      <button onClick={onClose} aria-label="Close alert" className="alert-close">
        <Icon name="x" />
      </button>
    )}
  </div>
);

// Toast (überlagert, selbstschließend)
export const Toast: React.FC<AlertProps & { autoClose?: number }> = ({ autoClose = 3000, ...props }) => {
  useEffect(() => {
    if (autoClose) {
      const timer = setTimeout(props.onClose, autoClose);
      return () => clearTimeout(timer);
    }
  }, [autoClose, props.onClose]);

  return <Alert {...props} />;
};
```

#### Pagination

```typescript
// src/components/molecules/Pagination/Pagination.tsx

interface PaginationProps {
  current: number;
  total: number;
  pageSize: number;
  onChange: (page: number) => void;
}

export const Pagination: React.FC<PaginationProps> = ({ current, total, pageSize, onChange }) => {
  const pages = Math.ceil(total / pageSize);

  return (
    <div className="pagination">
      <Button variant="ghost" size="sm" onClick={() => onChange(current - 1)} disabled={current === 1}>
        ← {t('common:previous')}
      </Button>
      {Array.from({ length: pages }, (_, i) => i + 1)
        .filter((p) => p === 1 || p === pages || (p >= current - 1 && p <= current + 1))
        .map((p) => (
          <Button
            key={p}
            variant={p === current ? 'primary' : 'ghost'}
            size="sm"
            onClick={() => onChange(p)}
          >
            {p}
          </Button>
        ))}
      <Button variant="ghost" size="sm" onClick={() => onChange(current + 1)} disabled={current === pages}>
        {t('common:next')} →
      </Button>
    </div>
  );
};
```

#### Tabs

```typescript
// src/components/molecules/Tabs/Tabs.tsx

interface Tab { id: string; label: string; }

interface TabsProps {
  tabs: Tab[];
  activeTab: string;
  onChange: (tabId: string) => void;
  variant?: 'default' | 'pills';
}

export const Tabs: React.FC<TabsProps & { children: React.ReactNode }> = ({ tabs, activeTab, onChange, variant = 'default', children }) => (
  <div className={cn('tabs', `tabs--${variant}`)}>
    <div className="tabs-list" role="tablist">
      {tabs.map((tab) => (
        <button
          key={tab.id}
          className={cn('tabs-trigger', { 'tabs-trigger--active': activeTab === tab.id })}
          onClick={() => onChange(tab.id)}
          role="tab"
          aria-selected={activeTab === tab.id}
        >
          {tab.label}
        </button>
      ))}
    </div>
    <div className="tabs-content" role="tabpanel">
      {children}
    </div>
  </div>
);
```

#### Breadcrumb Navigation

```typescript
// src/components/molecules/Breadcrumb/Breadcrumb.tsx

interface BreadcrumbItem { label: string; href?: string; }

export const Breadcrumb: React.FC<{ items: BreadcrumbItem[] }> = ({ items }) => (
  <nav aria-label="Breadcrumb">
    <ol className="breadcrumb">
      {items.map((item, i) => (
        <li key={i} className="breadcrumb-item">
          {item.href ? <a href={item.href}>{item.label}</a> : <span>{item.label}</span>}
          {i < items.length - 1 && <span className="breadcrumb-sep">/</span>}
        </li>
      ))}
    </ol>
  </nav>
);
```

#### Dropdown Menu

```typescript
// src/components/molecules/DropdownMenu/DropdownMenu.tsx

interface MenuItemProps { label: string; onClick: () => void; icon?: string; }

export const DropdownMenu: React.FC<{ trigger: React.ReactNode; items: MenuItemProps[] }> = ({ trigger, items }) => {
  const [open, setOpen] = React.useState(false);
  const ref = React.useRef<HTMLDivElement>(null);

  useClickOutside(ref, () => setOpen(false));

  return (
    <div className="dropdown" ref={ref}>
      <button onClick={() => setOpen(!open)} className="dropdown-trigger">
        {trigger}
      </button>
      {open && (
        <div className="dropdown-menu" role="menu">
          {items.map((item, i) => (
            <button
              key={i}
              onClick={() => { item.onClick(); setOpen(false); }}
              className="dropdown-item"
              role="menuitem"
            >
              {item.icon && <Icon name={item.icon} />}
              {item.label}
            </button>
          ))}
        </div>
      )}
    </div>
  );
};
```

---

## 3. ORGANISMS (10+)

### 3.1 DataTable (Virtual Scroll für 10.000+ Zeilen)

```typescript
// src/components/organisms/DataTable/DataTable.tsx

interface DataTableColumn<T> {
  key: keyof T;
  header: string;
  render?: (value: T[keyof T], row: T) => React.ReactNode;
  sortable?: boolean;
  width?: string | number;
}

interface DataTableProps<T> {
  columns: DataTableColumn<T>[];
  data: T[];
  onSelect?: (rows: T[]) => void;
  sortBy?: string;
  sortOrder?: 'asc' | 'desc';
  onSort?: (key: string, order: 'asc' | 'desc') => void;
  loading?: boolean;
  rowHeight?: number;
}

export const DataTable = React.forwardRef<any, DataTableProps<any>>(
  ({ columns, data, onSelect, sortBy, sortOrder, onSort, loading, rowHeight = 40 }) => {
    const [selectedRows, setSelectedRows] = React.useState<Set<number>>(new Set());

    // Virtual Scroll mit TanStack/React-Virtual
    const { getVirtualItems, totalSize } = useVirtualizer({
      count: data.length,
      size: rowHeight,
      overscan: 10,
    });

    return (
      <div className="data-table">
        <table className="data-table-table">
          <thead className="data-table-head">
            <tr>
              <th className="data-table-checkbox">
                <input
                  type="checkbox"
                  onChange={(e) => {
                    if (e.target.checked) {
                      setSelectedRows(new Set(data.map((_, i) => i)));
                    } else {
                      setSelectedRows(new Set());
                    }
                    onSelect?.(Array.from(selectedRows).map((i) => data[i]));
                  }}
                />
              </th>
              {columns.map((col) => (
                <th key={String(col.key)} className="data-table-cell">
                  {col.sortable ? (
                    <button onClick={() => onSort?.(String(col.key), sortOrder === 'asc' ? 'desc' : 'asc')}>
                      {col.header}
                      {sortBy === String(col.key) && <Icon name={sortOrder === 'asc' ? 'arrow-up' : 'arrow-down'} />}
                    </button>
                  ) : (
                    col.header
                  )}
                </th>
              ))}
            </tr>
          </thead>
          <tbody style={{ height: totalSize }}>
            {getVirtualItems().map((virtualItem) => (
              <tr
                key={virtualItem.index}
                className="data-table-row"
                style={{ transform: `translateY(${virtualItem.start}px)` }}
              >
                <td className="data-table-checkbox">
                  <input
                    type="checkbox"
                    checked={selectedRows.has(virtualItem.index)}
                    onChange={(e) => {
                      const newSet = new Set(selectedRows);
                      if (e.target.checked) {
                        newSet.add(virtualItem.index);
                      } else {
                        newSet.delete(virtualItem.index);
                      }
                      setSelectedRows(newSet);
                      onSelect?.(Array.from(newSet).map((i) => data[i]));
                    }}
                  />
                </td>
                {columns.map((col) => (
                  <td key={String(col.key)} className="data-table-cell">
                    {col.render ? col.render(data[virtualItem.index][col.key], data[virtualItem.index]) : String(data[virtualItem.index][col.key])}
                  </td>
                ))}
              </tr>
            ))}
          </tbody>
        </table>
        {loading && <div className="data-table-loading"><Spinner /></div>}
      </div>
    );
  }
);
```

### 3.2 Modal/Dialog

```typescript
// src/components/organisms/Modal/Modal.tsx

interface ModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  title?: string;
  description?: string;
  children: React.ReactNode;
  actions?: { label: string; onClick: () => void; variant?: 'primary' | 'danger' }[];
  size?: 'sm' | 'md' | 'lg';
}

export const Modal: React.FC<ModalProps> = ({ open, onOpenChange, title, description, children, actions, size = 'md' }) => {
  if (!open) return null;

  return (
    <div className="modal-overlay" onClick={() => onOpenChange(false)} role="presentation">
      <div className={cn('modal', `modal--${size}`)} onClick={(e) => e.stopPropagation()} role="dialog" aria-labelledby="modal-title">
        <div className="modal-header">
          <h2 id="modal-title" className="modal-title">{title}</h2>
          <button onClick={() => onOpenChange(false)} aria-label="Close modal" className="modal-close">
            <Icon name="x" />
          </button>
        </div>
        {description && <p className="modal-description">{description}</p>}
        <div className="modal-content">{children}</div>
        {actions && (
          <div className="modal-actions">
            {actions.map((action, i) => (
              <Button key={i} variant={action.variant || 'primary'} onClick={action.onClick}>
                {action.label}
              </Button>
            ))}
          </div>
        )}
      </div>
    </div>
  );
};
```

### 3.3 Scanner-Overlay (Lisa, Zebra TC21)

```typescript
// src/components/organisms/ScannerOverlay/ScannerOverlay.tsx

interface ScannerOverlayProps {
  onScan: (result: ScanResult) => void;
  onError?: (error: string) => void;
  expectedItemId?: string;  // Optional: Validation gegen erwartetes Item
}

export const ScannerOverlay: React.FC<ScannerOverlayProps> = ({ onScan, onError, expectedItemId }) => {
  const [cameraActive, setCameraActive] = React.useState(true);
  const [lastScan, setLastScan] = React.useState<ScanResult | null>(null);
  const [feedbackState, setFeedbackState] = React.useState<'idle' | 'success' | 'error' | 'warn'>('idle');
  const videoRef = React.useRef<HTMLVideoElement>(null);

  // @zxing/browser für Barcode-Scanning
  useEffect(() => {
    if (!cameraActive || !videoRef.current) return;

    const codeReader = new BrowserMultiFormatReader();
    codeReader.decodeFromVideoElement(videoRef.current, (result, err) => {
      if (result) {
        const barcode = result.getText();
        const isExpected = !expectedItemId || barcode === expectedItemId;

        setLastScan({ barcode, timestamp: new Date(), valid: isExpected });
        setFeedbackState(isExpected ? 'success' : 'error');

        // Vibration + Audio Feedback
        if ('vibrate' in navigator) {
          navigator.vibrate(isExpected ? [50, 100, 50] : [100, 50, 100, 50, 100]);
        }

        const audio = new Audio(`/audio/${isExpected ? 'beep-success' : 'beep-error'}.mp3`);
        audio.play().catch(() => {});

        // Callback
        onScan({ barcode, timestamp: new Date(), valid: isExpected });

        // Auto-Reset nach 2 Sekunden
        setTimeout(() => setFeedbackState('idle'), 2000);
      }
    });

    return () => codeReader.reset();
  }, [cameraActive, expectedItemId, onScan]);

  return (
    <div className={cn('scanner-overlay', `scanner--${feedbackState}`)}>
      {/* Kamera-Stream */}
      <video ref={videoRef} className="scanner-video" autoPlay playsInline muted />

      {/* Scan-Rahmen (Crosshair) */}
      <svg className="scanner-frame" viewBox="0 0 480 640">
        <defs>
          <mask id="scanner-mask">
            <rect width="480" height="640" fill="white" />
            <rect x="60" y="240" width="360" height="160" fill="black" />
          </mask>
        </defs>
        <rect width="480" height="640" fill="rgba(0, 0, 0, 0.5)" mask="url(#scanner-mask)" />

        {/* Eckpunkte */}
        <line x1="60" y1="240" x2="100" y2="240" stroke="currentColor" strokeWidth="2" />
        <line x1="60" y1="240" x2="60" y2="280" stroke="currentColor" strokeWidth="2" />
        <line x1="420" y1="240" x2="380" y2="240" stroke="currentColor" strokeWidth="2" />
        <line x1="420" y1="240" x2="420" y2="280" stroke="currentColor" strokeWidth="2" />
        {/* Bottom Corners */}
        <line x1="60" y1="400" x2="100" y2="400" stroke="currentColor" strokeWidth="2" />
        <line x1="60" y1="400" x2="60" y2="360" stroke="currentColor" strokeWidth="2" />
        <line x1="420" y1="400" x2="380" y2="400" stroke="currentColor" strokeWidth="2" />
        <line x1="420" y1="400" x2="420" y2="360" stroke="currentColor" strokeWidth="2" />
      </svg>

      {/* Feedback */}
      <div className={cn('scanner-feedback', `scanner-feedback--${feedbackState}`)}>
        {feedbackState === 'success' && <Icon name="check-circle" />}
        {feedbackState === 'error' && <Icon name="x-circle" />}
        {feedbackState === 'warn' && <Icon name="alert-circle" />}
        <p>{feedbackMessageMap[feedbackState]}</p>
      </div>

      {/* Letzte 50 Scans */}
      <div className="scanner-history">
        {lastScan && (
          <div className="scanner-history-item">
            {lastScan.barcode} — {lastScan.timestamp.toLocaleTimeString()}
          </div>
        )}
      </div>

      {/* Quick Actions */}
      <div className="scanner-actions">
        <Button size="lg" variant={cameraActive ? 'danger' : 'success'} onClick={() => setCameraActive(!cameraActive)}>
          {cameraActive ? 'Kamera aus' : 'Kamera an'}
        </Button>
      </div>
    </div>
  );
};
```

```scss
// src/components/organisms/ScannerOverlay/ScannerOverlay.module.scss

.scannerOverlay {
  position: fixed;
  inset: 0;
  z-index: var(--z-scanner);
  background-color: var(--color-gray-900);
  overflow: hidden;

  &--success {
    .scannerFeedback { background-color: var(--color-success-600); }
  }

  &--error {
    .scannerFeedback { background-color: var(--color-error-600); }
  }

  &--warn {
    .scannerFeedback { background-color: var(--color-warning-600); }
  }
}

.scannerVideo {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.scannerFrame {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  color: white;
  stroke: white;

  // Pulsierender Rahmen
  animation: pulse 1s infinite;
}

.scannerFeedback {
  position: absolute;
  bottom: 80px;
  left: 50%;
  transform: translateX(-50%);
  padding: var(--space-4);
  border-radius: var(--radius-lg);
  color: white;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--space-2);
  font-weight: var(--font-weight-600);
  transition: var(--transition-all);
}

.scannerActions {
  position: absolute;
  bottom: 20px;
  left: 50%;
  transform: translateX(-50%);
  display: flex;
  gap: var(--space-2);
  width: 100%;
  padding: var(--space-4);
  max-width: 480px;
  box-sizing: border-box;
}

.scannerHistory {
  position: absolute;
  top: 20px;
  left: 20px;
  right: 20px;
  max-height: 200px;
  overflow-y: auto;
  background: rgba(0, 0, 0, 0.5);
  border-radius: var(--radius-md);
  padding: var(--space-2);

  .scannerHistoryItem {
    color: white;
    font-size: var(--font-size-sm);
    padding: var(--space-1);
    border-bottom: 1px solid rgba(255, 255, 255, 0.1);

    &:last-child { border-bottom: none; }
  }
}
```

### 3.4 Sidebar Navigation

```typescript
// src/components/organisms/Sidebar/Sidebar.tsx

interface NavItem { label: string; href: string; icon: string; badge?: number; children?: NavItem[]; }

export const Sidebar: React.FC<{ items: NavItem[]; collapsed?: boolean }> = ({ items, collapsed }) => {
  const location = useLocation();
  const [expandedItems, setExpandedItems] = React.useState<string[]>([]);

  return (
    <aside className={cn('sidebar', { 'sidebar--collapsed': collapsed })} role="navigation">
      <nav className="sidebar-nav">
        {items.map((item) => (
          <div key={item.href}>
            <a
              href={item.href}
              className={cn('sidebar-item', { 'sidebar-item--active': location.pathname === item.href })}
              onClick={() => item.children && setExpandedItems((p) =>
                p.includes(item.href) ? p.filter((x) => x !== item.href) : [...p, item.href]
              )}
            >
              <Icon name={item.icon} />
              {!collapsed && <span>{item.label}</span>}
              {item.badge && !collapsed && <Badge variant="danger">{item.badge}</Badge>}
            </a>
            {item.children && expandedItems.includes(item.href) && !collapsed && (
              <div className="sidebar-submenu">
                {item.children.map((child) => (
                  <a key={child.href} href={child.href} className="sidebar-subitem">
                    {child.label}
                  </a>
                ))}
              </div>
            )}
          </div>
        ))}
      </nav>
    </aside>
  );
};
```

### 3.5 Header mit Role-Awareness

```typescript
// src/components/organisms/Header/Header.tsx

export const Header: React.FC = () => {
  const { user, role } = useAuth();
  const { unreadCount } = useNotifications();

  return (
    <header className="header">
      <div className="header-left">
        <a href="/" className="header-logo">CrateDesk</a>
        <span className="header-role">{role}</span>
      </div>

      <div className="header-center">
        {/* Role-spezifische Quick Links */}
        {role === 'geschaeftsführung' && (
          <div className="header-quick-access">
            <Button variant="ghost" size="sm" onClick={() => navigate('/dashboard')}>
              {t('common:dashboard')}
            </Button>
            <Button variant="ghost" size="sm" onClick={() => navigate('/projects')}>
              {t('common:projects')}
            </Button>
          </div>
        )}
      </div>

      <div className="header-right">
        <Button variant="ghost" size="sm" isIconOnly>
          <Icon name="bell" />
          {unreadCount > 0 && <Badge>{unreadCount}</Badge>}
        </Button>

        <DropdownMenu
          trigger={<Avatar src={user?.avatar} initials={user?.name.slice(0, 2)} />}
          items={[
            { label: t('common:settings'), onClick: () => navigate('/settings') },
            { label: t('common:logout'), onClick: () => logout() },
          ]}
        />
      </div>
    </header>
  );
};
```

Fortsetzung im nächsten Teil...

---

## 4. PAGE LAYOUTS PRO ROLLE

### 4.1 Marco (GF) — Desktop Dashboard

```typescript
// src/features/dashboard/pages/MarcosDashboard.tsx

export const MarcosDashboard: React.FC = () => {
  const { data: projects, isLoading } = useProjectsList({ status: 'all' });
  const { data: openInvoices } = useOpenInvoices();

  return (
    <DashboardLayout>
      {/* KPI-Cards oben */}
      <div className="grid grid-cols-4 gap-4 mb-6">
        <KPICard
          title={t('dashboard:thisWeekRevenue')}
          value={`€ ${formatCurrency(projects.thisWeek.totalRevenue)}`}
          trend={`+${projects.trend}%`}
          icon="trending-up"
        />
        <KPICard title={t('dashboard:equipment')} value={equipmentCount} icon="box" />
        <KPICard title={t('dashboard:openInvoices')} value={openInvoices.length} variant="warning" />
        <KPICard
          title={t('dashboard:freelancersActive')}
          value={activeFreelancers.length}
          icon="users"
        />
      </div>

      {/* Haupt-Layout: 70% Links, 30% Rechts */}
      <div className="grid grid-cols-3 gap-6">
        {/* Linke Spalte: Projekte + Timeline */}
        <div className="col-span-2">
          <Card>
            <CardHeader title={t('dashboard:thisWeek')} />
            <Tabs
              tabs={[
                { id: 'kanban', label: 'Kanban' },
                { id: 'list', label: 'Liste' },
              ]}
              activeTab={view}
              onChange={setView}
            >
              {view === 'kanban' ? <ProjectKanban projects={projects} /> : <ProjectTimeline projects={projects} />}
            </Tabs>
          </Card>
        </div>

        {/* Rechte Spalte: Alerts + Quick Stats */}
        <div className="space-y-4">
          <Card>
            <CardHeader title={t('dashboard:alerts')} />
            {/* Double-Bookings, Überfällige Rechnungen, Defekte Items */}
            <AlertStack alerts={alerts} />
          </Card>

          <Card>
            <CardHeader title={t('dashboard:lagerStatus')} />
            <EquipmentStatusCard />
          </Card>
        </div>
      </div>
    </DashboardLayout>
  );
};
```

### 4.2 Lisa (Lager) — Zebra TC21 Scanner

```typescript
// src/features/warehouse/pages/ScannerPage.tsx

export const ScannerPage: React.FC = () => {
  const [scanMode, setScanMode] = useZustand((s) => [s.scanMode, s.setScanMode]);
  const { mutate: checkIn } = useCheckInEquipment();
  const { data: packingList } = usePackingList(currentJobId);

  return (
    <ScannerLayout>
      {/* Großer Scan-Button oben */}
      <div className="scanner-header">
        <Button size="xl" variant="primary" className="scanner-button-primary">
          {t('warehouse:scan')}
        </Button>
        <p className="scanner-mode-label">{t(`warehouse:mode_${scanMode}`)}</p>
      </div>

      {/* Modus-Umschalter (3 Buttons horizontal) */}
      <div className="scanner-mode-selector">
        <Button
          size="lg"
          variant={scanMode === 'checkin' ? 'primary' : 'ghost'}
          onClick={() => setScanMode('checkin')}
        >
          {t('warehouse:checkIn')}
        </Button>
        <Button
          size="lg"
          variant={scanMode === 'checkout' ? 'primary' : 'ghost'}
          onClick={() => setScanMode('checkout')}
        >
          {t('warehouse:checkOut')}
        </Button>
        <Button
          size="lg"
          variant={scanMode === 'locate' ? 'primary' : 'ghost'}
          onClick={() => setScanMode('locate')}
        >
          {t('warehouse:locate')}
        </Button>
      </div>

      {/* Scan-Overlay mit Video */}
      <ScannerOverlay
        onScan={(result) => {
          checkIn({ barcode: result.barcode, jobId: currentJobId });
        }}
        expectedItemId={packingList?.items[nextItemIndex]?.barcode}
      />

      {/* Packliste (scrollbar, mit Abhak-Status) */}
      <div className="scanner-packing-list">
        <h3>{t('warehouse:packingList')}</h3>
        <ul className="space-y-2">
          {packingList?.items.map((item, i) => (
            <li
              key={i}
              className={cn('packing-item', {
                'packing-item--done': item.scanned,
                'packing-item--next': i === nextItemIndex,
              })}
            >
              <Checkbox checked={item.scanned} readOnly />
              <span>{item.name} ×{item.quantity}</span>
              {item.scanned && <Icon name="check" />}
            </li>
          ))}
        </ul>
        <ProgressBar value={scannedCount} max={packingList?.items.length} />
      </div>
    </ScannerLayout>
  );
};
```

### 4.3 Thomas (Buchhaltung) — Rechnungen

```typescript
// src/features/accounting/pages/InvoiceListPage.tsx

export const InvoiceListPage: React.FC = () => {
  const [sortBy, setSortBy] = React.useState('dueDate');
  const [sortOrder, setSortOrder] = React.useState<'asc' | 'desc'>('asc');
  const { data: invoices } = useInvoices({ sortBy, sortOrder });

  return (
    <DashboardLayout>
      <Card>
        <CardHeader
          title={t('accounting:invoices')}
          action={<Button onClick={() => navigate('/accounting/invoices/new')}>{t('common:new')}</Button>}
        />

        {/* Excel-ähnliche Tabelle mit Tastatur-Navigation */}
        <DataTable
          columns={[
            { key: 'number', header: t('accounting:invoiceNumber'), sortable: true, width: '100px' },
            { key: 'customer', header: t('accounting:customer'), sortable: true },
            { key: 'amount', header: t('accounting:amount'), sortable: true, width: '120px', render: (v) => formatCurrency(v) },
            { key: 'dueDate', header: t('accounting:dueDate'), sortable: true, width: '100px', render: (v) => formatDate(v) },
            {
              key: 'status',
              header: t('accounting:status'),
              render: (status) => (
                <Badge variant={status === 'paid' ? 'success' : status === 'overdue' ? 'error' : 'warning'}>
                  {t(`accounting:status_${status}`)}
                </Badge>
              ),
            },
            {
              key: 'actions',
              header: '',
              render: (_, row) => (
                <DropdownMenu
                  trigger={<Button variant="ghost" size="sm" isIconOnly><Icon name="more-vertical" /></Button>}
                  items={[
                    { label: t('common:edit'), onClick: () => navigate(`/accounting/invoices/${row.id}/edit`) },
                    { label: t('common:print'), onClick: () => printInvoice(row.id) },
                    { label: t('common:delete'), onClick: () => deleteInvoice(row.id) },
                  ]}
                />
              ),
            },
          ]}
          data={invoices}
          sortBy={sortBy}
          sortOrder={sortOrder}
          onSort={(key, order) => { setSortBy(key); setSortOrder(order); }}
        />

        {/* Offene-Posten-Ampel */}
        <div className="accounting-summary grid grid-cols-3 gap-4 mt-6">
          <Card variant="success">
            <p>{t('accounting:paid')}</p>
            <p className="text-lg font-bold">{formatCurrency(summary.paid)}</p>
          </Card>
          <Card variant="warning">
            <p>{t('accounting:pending')}</p>
            <p className="text-lg font-bold">{formatCurrency(summary.pending)}</p>
          </Card>
          <Card variant="error">
            <p>{t('accounting:overdue')}</p>
            <p className="text-lg font-bold">{formatCurrency(summary.overdue)}</p>
          </Card>
        </div>
      </Card>
    </DashboardLayout>
  );
};
```

### 4.4 Kevin (Freelancer) — iPhone Job-Dashboard

```typescript
// src/features/freelancer/pages/JobsPage.tsx

export const JobsPage: React.FC = () => {
  const { data: jobs } = useMyJobs();

  return (
    <MobileLayout>
      <div className="freelancer-jobs">
        {/* Diese Woche */}
        <h2 className="section-title">{t('freelancer:thisWeek')}</h2>
        <div className="space-y-3">
          {jobs.filter((j) => isThisWeek(j.date)).map((job) => (
            <Card
              key={job.id}
              className="job-card"
              onClick={() => navigate(`/freelancer/jobs/${job.id}`)}
            >
              <div className="job-card-header">
                <h3 className="job-title">{job.title}</h3>
                <Badge variant={job.status === 'confirmed' ? 'success' : 'warning'}>
                  {job.status}
                </Badge>
              </div>

              <div className="job-card-meta space-y-1 text-sm text-gray-500">
                <p>📅 {formatDate(job.date)}</p>
                <p>🕐 {job.startTime} - {job.endTime}</p>
                <p>📍 {job.location}</p>
                <p>👤 {job.contact}</p>
              </div>

              <Button size="sm" className="w-full mt-3">
                {t('freelancer:viewDetails')}
              </Button>
            </Card>
          ))}
        </div>

        {/* Zeiterfassung Quick Access */}
        <div className="freelancer-time-entry mt-6">
          <Button size="lg" variant="primary" className="w-full" onClick={() => startTimeEntry()}>
            {t('freelancer:checkIn')}
          </Button>
        </div>
      </div>
    </MobileLayout>
  );
};
```

---

## 5. STATE MANAGEMENT ARCHITECTURE

### 5.1 Zustand Stores

```typescript
// src/stores/uiStore.ts

interface UIState {
  sidebarCollapsed: boolean;
  theme: 'light' | 'dark' | 'highcontrast';
  modalOpen: boolean;
  activeTab: string;
  toggleSidebar: () => void;
  setTheme: (theme: UIState['theme']) => void;
  setModalOpen: (open: boolean) => void;
  setActiveTab: (tab: string) => void;
}

export const useUIStore = create<UIState>((set) => ({
  sidebarCollapsed: false,
  theme: 'light',
  modalOpen: false,
  activeTab: 'overview',
  toggleSidebar: () => set((s) => ({ sidebarCollapsed: !s.sidebarCollapsed })),
  setTheme: (theme) => set({ theme }),
  setModalOpen: (open) => set({ modalOpen: open }),
  setActiveTab: (tab) => set({ activeTab: tab }),
}));

// src/stores/scannerStore.ts

interface ScanResult {
  barcode: string;
  timestamp: Date;
  valid: boolean;
}

interface ScannerState {
  currentJobId: string | null;
  scanMode: 'checkin' | 'checkout' | 'inventory' | 'locate';
  lastScans: ScanResult[];
  offlineQueue: OfflineScan[];
  isSyncing: boolean;
  setScanMode: (mode: ScannerState['scanMode']) => void;
  addScan: (result: ScanResult) => void;
  queueScan: (scan: OfflineScan) => void;
  syncQueue: () => Promise<void>;
}

export const useScannerStore = create<ScannerState>((set) => ({
  currentJobId: null,
  scanMode: 'checkin',
  lastScans: [],
  offlineQueue: [],
  isSyncing: false,
  setScanMode: (mode) => set({ scanMode: mode }),
  addScan: (result) => set((s) => ({ lastScans: [result, ...s.lastScans].slice(0, 50) })),
  queueScan: (scan) => set((s) => ({ offlineQueue: [...s.offlineQueue, scan] })),
  syncQueue: async () => {
    // Implementierung
  },
}));
```

### 5.2 React Query Patterns

```typescript
// src/api/equipment.queries.ts

export const equipmentKeys = {
  all: ['equipment'] as const,
  list: (filters: EquipmentFilters) => ['equipment', 'list', filters] as const,
  detail: (id: string) => ['equipment', 'detail', id] as const,
  availability: (id: string, from: Date, to: Date) => ['equipment', 'availability', id, from, to] as const,
};

export function useEquipmentList(filters: EquipmentFilters) {
  return useQuery({
    queryKey: equipmentKeys.list(filters),
    queryFn: () => api.equipment.list(filters),
    staleTime: 30_000,
    gcTime: 5 * 60_000,
  });
}

export function useCheckInEquipment() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: CheckInRequest) => api.equipment.checkIn(data),
    onMutate: async (newData) => {
      // Optimistic update
      await queryClient.cancelQueries({ queryKey: equipmentKeys.all });
      const previous = queryClient.getQueryData(equipmentKeys.all);

      queryClient.setQueryData(equipmentKeys.all, (old: any) => ({
        ...old,
        items: old.items.map((item: any) =>
          item.id === newData.equipmentId ? { ...item, status: 'in' } : item
        ),
      }));

      return { previous };
    },
    onError: (err, newData, context) => {
      // Rollback
      if (context?.previous) {
        queryClient.setQueryData(equipmentKeys.all, context.previous);
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: equipmentKeys.all });
    },
  });
}
```

---

## 6. RESPONSIVE STRATEGY

### 6.1 Breakpoint-Verhalten für Hauptseiten

```scss
// src/styles/responsive/dashboard.scss

.dashboard {
  // Desktop (lg+)
  @media (min-width: 1024px) {
    display: grid;
    grid-template-columns: 1fr 1fr 1fr 1fr;
    gap: var(--space-6);
  }

  // Tablet (md-lg)
  @media (min-width: 768px) and (max-width: 1023px) {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: var(--space-4);
  }

  // Mobile (sm-md)
  @media (max-width: 767px) {
    display: grid;
    grid-template-columns: 1fr;
    gap: var(--space-3);
  }
}

// Zebra TC21 Spezial (480×800px, Landscape)
.scanner-layout {
  display: flex;
  flex-direction: column;
  height: 100dvh; // Dynamic Viewport Height für iPhone Notch

  .scanner-header {
    flex: 0 0 auto;
    padding: var(--space-4);
    background: var(--color-gray-900);
  }

  .scanner-content {
    flex: 1 1 auto;
    overflow-y: auto;
  }

  .scanner-footer {
    flex: 0 0 auto;
    padding: var(--space-4);
  }
}
```

### 6.2 Container Queries (Komponenten-Responsive)

```scss
// src/components/molecules/DataTable/DataTable.module.scss

.dataTable {
  container-type: inline-size;
}

@container (min-width: 800px) {
  .dataTableCell {
    padding: var(--space-3) var(--space-4);
    font-size: var(--font-size-base);
  }
}

@container (max-width: 799px) {
  .dataTableCell {
    padding: var(--space-2) var(--space-2);
    font-size: var(--font-size-sm);
    overflow: hidden;
    text-overflow: ellipsis;
  }
}
```

### 6.3 Safe Area Insets (iPhone Notch)

```scss
// src/styles/responsive/safe-area.scss

.header {
  padding-top: max(var(--space-4), env(safe-area-inset-top));
  padding-left: max(var(--space-4), env(safe-area-inset-left));
  padding-right: max(var(--space-4), env(safe-area-inset-right));
}

.bottomNav {
  padding-bottom: max(var(--space-4), env(safe-area-inset-bottom));
}
```

---

## 7. PERFORMANCE BUDGET & OPTIMIZATION

### 7.1 Bundle Size Targets

```json
{
  "budgets": [
    {
      "type": "bundle",
      "name": "main",
      "baselineFile": "./baseline.json",
      "maxSize": "200kb"
    },
    {
      "type": "bundle",
      "name": "scanner",
      "maxSize": "150kb"
    },
    {
      "type": "bundle",
      "name": "accounting",
      "maxSize": "180kb"
    }
  ]
}
```

### 7.2 Code Splitting

```typescript
// src/App.tsx

const DashboardPage = lazy(() => import('./features/dashboard/pages/DashboardPage'));
const ScannerPage = lazy(() => import('./features/warehouse/pages/ScannerPage'));
const InvoiceListPage = lazy(() => import('./features/accounting/pages/InvoiceListPage'));
const FreelancerJobsPage = lazy(() => import('./features/freelancer/pages/JobsPage'));

// Route-basiertes Splitting
<Routes>
  <Route path="/dashboard" element={<Suspense fallback={<Spinner />}><DashboardPage /></Suspense>} />
  <Route path="/scanner" element={<Suspense fallback={<Spinner />}><ScannerPage /></Suspense>} />
</Routes>
```

### 7.3 Image Optimization

```typescript
// src/components/atoms/Image/Image.tsx

export const Image: React.FC<{ src: string; alt: string }> = ({ src, alt }) => {
  return (
    <img
      src={src}
      alt={alt}
      srcSet={`
        ${src}?w=320&q=75 320w,
        ${src}?w=640&q=80 640w,
        ${src}?w=1280&q=85 1280w
      `}
      sizes="(max-width: 640px) 100vw, 50vw"
      loading="lazy"
      decoding="async"
    />
  );
};
```

### 7.4 Virtual Scrolling (10.000+ Zeilen Tabelle)

```typescript
// Bereits in DataTable.tsx gezeigt
// useVirtualizer aus @tanstack/react-virtual
```

---

## 8. i18n IMPLEMENTATION

### 8.1 Struktur

```
src/i18n/
├── de/
│   ├── common.json
│   ├── dashboard.json
│   ├── warehouse.json
│   ├── accounting.json
│   ├── freelancer.json
│   └── errors.json
└── en/
    ├── common.json
    ├── dashboard.json
    └── ...
```

### 8.2 Keys (deutsch-first)

```json
{
  "common": {
    "save": "Speichern",
    "cancel": "Abbrechen",
    "delete": "Löschen",
    "new": "Neu",
    "edit": "Bearbeiten",
    "dashboard": "Dashboard",
    "settings": "Einstellungen",
    "logout": "Abmelden"
  },
  "warehouse": {
    "scan": "Scannen",
    "checkIn": "Einlagern",
    "checkOut": "Auslagern",
    "scan_success": "✓ Scan erfolgreich",
    "scan_error": "✗ Fehler beim Scan",
    "packingList": "Packliste",
    "mode_checkin": "Einlagerungsmodus",
    "mode_checkout": "Auslagerungsmodus",
    "mode_locate": "Suchen"
  },
  "accounting": {
    "invoices": "Rechnungen",
    "invoiceNumber": "Rechnungsnr.",
    "customer": "Kunde",
    "amount": "Betrag",
    "dueDate": "Fälligkeitsdatum",
    "status_paid": "Bezahlt",
    "status_overdue": "Überfällig",
    "status_pending": "Offen"
  }
}
```

### 8.3 Setup

```typescript
// src/i18n/i18n.ts

import i18n from 'i18next';
import { initReactI18next } from 'react-i18next';
import de_common from './de/common.json';
import de_warehouse from './de/warehouse.json';
import en_common from './en/common.json';

i18n.use(initReactI18next).init({
  lng: navigator.language.startsWith('de') ? 'de' : 'en',
  fallbackLng: 'en',
  resources: {
    de: {
      common: de_common,
      warehouse: de_warehouse,
    },
    en: {
      common: en_common,
    },
  },
  ns: ['common', 'warehouse'],
  defaultNS: 'common',
});

export default i18n;
```

### 8.4 Verwendung im Code

```typescript
import { useTranslation } from 'react-i18next';

export const MyComponent = () => {
  const { t } = useTranslation(['warehouse', 'common']);

  return (
    <Button onClick={handleSave}>
      {t('common:save')}
    </Button>
  );
};
```

---

## 9. ACCESSIBILITY (WCAG 2.1 AA)

### 9.1 Keyboard Navigation

```typescript
// src/components/molecules/DropdownMenu/DropdownMenu.tsx

useEffect(() => {
  if (!open) return;

  const handleKeyDown = (e: KeyboardEvent) => {
    if (e.key === 'Escape') setOpen(false);
    if (e.key === 'ArrowDown') focusNextItem();
    if (e.key === 'ArrowUp') focusPreviousItem();
    if (e.key === 'Enter') selectCurrentItem();
  };

  document.addEventListener('keydown', handleKeyDown);
  return () => document.removeEventListener('keydown', handleKeyDown);
}, [open]);
```

### 9.2 ARIA Labels

```typescript
// Alle interaktiven Komponenten müssen aria-label oder aria-labelledby haben

<button aria-label="Close modal" onClick={onClose}>
  <Icon name="x" />
</button>

<div role="alert" aria-live="polite">
  {message}
</div>

<table role="grid">
  <thead role="rowgroup">
    <tr role="row">
      <th role="columnheader">{header}</th>
    </tr>
  </thead>
</table>
```

### 9.3 Color Contrast

```scss
// WCAG AA: 4.5:1 für Text, 3:1 für große Elemente
// WCAG AAA: 7:1 für Text, 4.5:1 für große Elemente

--color-primary-700: #1d4ed8; // Kontrast gegen white: 8.5:1 ✓
--color-gray-700: #374151;    // Kontrast gegen white: 10:1 ✓
--color-error-600: #dc2626;   // Kontrast gegen white: 6.5:1 ✓
```

---

## 10. OFFLINE-SUPPORT (PWA)

### 10.1 Service Worker mit Workbox

```typescript
// vite.config.ts

import { VitePWA } from 'vite-plugin-pwa';

export default {
  plugins: [
    VitePWA({
      strategies: 'injectManifest',
      workbox: {
        runtimeCaching: [
          {
            urlPattern: /^https:\/\/api\.example\.com\/.*/i,
            handler: 'NetworkFirst',
            options: {
              cacheName: 'api-cache',
              expiration: { maxEntries: 50, maxAgeSeconds: 86400 },
            },
          },
          {
            urlPattern: /^https:\/\/images\.example\.com\/.*/i,
            handler: 'CacheFirst',
            options: {
              cacheName: 'image-cache',
              expiration: { maxEntries: 200, maxAgeSeconds: 604800 },
            },
          },
        ],
      },
      manifest: {
        name: 'CrateDesk',
        short_name: 'CrateDesk',
        description: 'Self-hosted Lagerverwaltung für Veranstaltungstechnik',
        theme_color: '#2563eb',
        background_color: '#ffffff',
        display: 'standalone',
        icons: [
          { src: '/icon-192.png', sizes: '192x192', type: 'image/png', purpose: 'any maskable' },
          { src: '/icon-512.png', sizes: '512x512', type: 'image/png', purpose: 'any maskable' },
        ],
      },
    }),
  ],
};
```

### 10.2 IndexedDB für Offline-Sync

```typescript
// src/features/warehouse/offline/db.ts

import Dexie, { Table } from 'dexie';

export interface OfflineScan {
  id?: number;
  barcode: string;
  jobId: string;
  timestamp: Date;
  synced: boolean;
}

export class OfflineDB extends Dexie {
  scans!: Table<OfflineScan>;

  constructor() {
    super('cratedesk-offline');
    this.version(1).stores({
      scans: '++id, synced, timestamp',
    });
  }
}

export const db = new OfflineDB();

// Sync-Service
export async function syncOfflineScans() {
  const unsynced = await db.scans.where('synced').equals(false).toArray();

  for (const scan of unsynced) {
    try {
      await api.equipment.checkIn({ barcode: scan.barcode, jobId: scan.jobId });
      await db.scans.update(scan.id!, { synced: true });
    } catch (error) {
      console.error('Sync failed:', error);
    }
  }
}
```

---

## 11. WEBSOKET ARCHITEKTUR (Echtzeit)

### 11.1 Socket.io Setup

```typescript
// src/services/socketService.ts

import { io, Socket } from 'socket.io-client';

export class SocketService {
  private socket: Socket | null = null;

  connect() {
    this.socket = io(import.meta.env.VITE_API_URL, {
      auth: { token: getAuthToken() },
      reconnection: true,
      reconnectionDelay: 1000,
      reconnectionDelayMax: 5000,
      reconnectionAttempts: 5,
    });

    this.socket.on('equipment:checkedin', (data) => {
      // Update UI in Echtzeit
      queryClient.invalidateQueries({ queryKey: equipmentKeys.all });
    });

    this.socket.on('notification:new', (notification) => {
      notificationStore.addNotification(notification);
    });
  }

  disconnect() {
    this.socket?.disconnect();
  }

  emit(event: string, data: any) {
    this.socket?.emit(event, data);
  }

  on(event: string, callback: (data: any) => void) {
    this.socket?.on(event, callback);
  }
}

export const socketService = new SocketService();
```

### 11.2 Real-Time Events pro Feature

```typescript
// Warehouse: Scan-Events in Echtzeit
socketService.on('warehouse:scanned', (data: { barcode: string; jobId: string; status: 'success' | 'error' }) => {
  // Alle Nutzer sehen sofort dass ein Item gescannt wurde
  useScannerStore.getState().addScan(data);
});

// Dashboard: Equipment-Status-Änderungen
socketService.on('equipment:statuschanged', (data: { equipmentId: string; status: 'in' | 'out' | 'defect' }) => {
  queryClient.setQueryData(equipmentKeys.detail(data.equipmentId), (old: any) => ({
    ...old,
    status: data.status,
  }));
});

// Notifications: Neue Zuweisung
socketService.on('freelancer:jobassigned', (data: JobAssignment) => {
  notificationStore.addNotification({
    id: nanoid(),
    type: 'job_assigned',
    title: `Neuer Job: ${data.jobTitle}`,
    timestamp: new Date(),
  });
});
```

---

## 12. TESTING-STRATEGIE

### 12.1 Unit Tests (Vitest + RTL)

```typescript
// src/components/atoms/Button/__tests__/Button.test.tsx

import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { Button } from '../Button';

describe('Button', () => {
  it('renders with correct text', () => {
    render(<Button>Click me</Button>);
    expect(screen.getByText('Click me')).toBeInTheDocument();
  });

  it('calls onClick handler', async () => {
    const onClick = vi.fn();
    render(<Button onClick={onClick}>Click</Button>);
    await userEvent.click(screen.getByText('Click'));
    expect(onClick).toHaveBeenCalled();
  });

  it('disables when disabled prop is true', () => {
    render(<Button disabled>Disabled</Button>);
    expect(screen.getByText('Disabled')).toBeDisabled();
  });

  it('applies correct variant styling', () => {
    const { container } = render(<Button variant="danger">Delete</Button>);
    expect(container.querySelector('.button--danger')).toBeInTheDocument();
  });
});
```

### 12.2 E2E Tests (Playwright)

```typescript
// tests/e2e/scanner.spec.ts

import { test, expect } from '@playwright/test';

test.describe('Scanner Flow (Lisa)', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/scanner');
    await page.waitForLoadState('networkidle');
  });

  test('should scan item and update packing list', async ({ page }) => {
    // Simuliere Barcode-Scan
    await page.fill('[data-testid="scanner-input"]', 'BARCODE123');
    await page.press('[data-testid="scanner-input"]', 'Enter');

    // Überprüfe dass Item in Packliste abgehakt ist
    await expect(page.getByTestId('packing-item-BARCODE123')).toHaveClass('packing-item--done');

    // Überprüfe Progress-Bar
    await expect(page.getByTestId('progress-bar')).toHaveAttribute('aria-valuenow', '1');
  });

  test('should show error on invalid barcode', async ({ page }) => {
    await page.fill('[data-testid="scanner-input"]', 'INVALID');
    await page.press('[data-testid="scanner-input"]', 'Enter');

    await expect(page.getByTestId('scanner-error')).toBeVisible();
    await expect(page.getByTestId('scanner-error')).toHaveClass('scanner-feedback--error');
  });
});
```

---

## SUMMARY: Design System Implementation Checklist

- [ ] SCSS Variablen für alle Tokens definieren (colors, spacing, typography, shadows)
- [ ] TypeScript Constants Mirror erstellen
- [ ] 18+ Atom-Komponenten bauen (Button, Input, Badge, Avatar, etc.)
- [ ] 12+ Molecule-Komponenten bauen (FormField, SearchBar, DatePicker, Alert, Toast, Tabs, etc.)
- [ ] 10+ Organism-Komponenten bauen (DataTable, Modal, Scanner-Overlay, Sidebar, Header)
- [ ] Responsive SCSS-Module schreiben (Breakpoints, Container Queries, Safe Area Insets)
- [ ] Zustand Stores für UI/Scanner/Notifications einrichten
- [ ] React Query Patterns für Server State (Equipment, Projects, Invoices)
- [ ] i18n Setup (DE/EN Namespacing)
- [ ] Dark Mode + High-Contrast Mode implementieren
- [ ] Accessibility (WCAG 2.1 AA): Keyboard Navigation, ARIA Labels, Color Contrast
- [ ] PWA Setup: Service Worker, Manifest, Icon Set
- [ ] IndexedDB für Offline-Sync (Scanner-Queue)
- [ ] WebSocket Integration für Echtzeit-Events
- [ ] Code Splitting: Route-basiert + Component-basiert
- [ ] Image Optimization: WebP, Lazy Loading, Responsive Images
- [ ] Virtual Scrolling für Tabellen (10.000+ Zeilen)
- [ ] Unit Tests (Vitest + RTL)
- [ ] E2E Tests (Playwright)
- [ ] Bundle Size Budget (<200KB Initial)
- [ ] Performance Metrics (LCP <2.5s, CLS <0.1)

**Alle Dateien sind modular, typsicher (TypeScript), barrierearm (WCAG 2.1 AA), offline-fähig und auf Rollen optimiert.**

