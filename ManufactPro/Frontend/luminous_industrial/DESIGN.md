---
name: Luminous Industrial
colors:
  surface: '#fbf9f7'
  surface-dim: '#dbdad8'
  surface-bright: '#fbf9f7'
  surface-container-lowest: '#ffffff'
  surface-container-low: '#f5f3f1'
  surface-container: '#efeeeb'
  surface-container-high: '#e9e8e6'
  surface-container-highest: '#e4e2e0'
  on-surface: '#1b1c1b'
  on-surface-variant: '#414942'
  inverse-surface: '#30312f'
  inverse-on-surface: '#f2f0ee'
  outline: '#717972'
  outline-variant: '#c0c9c0'
  surface-tint: '#006c47'
  primary: '#006c47'
  on-primary: '#ffffff'
  primary-container: '#00885a'
  on-primary-container: '#000703'
  inverse-primary: '#60dda2'
  secondary: '#994700'
  on-secondary: '#ffffff'
  secondary-container: '#fe852b'
  on-secondary-container: '#632c00'
  tertiary: '#555e5b'
  on-tertiary: '#ffffff'
  tertiary-container: '#6e7673'
  on-tertiary-container: '#f6fffa'
  error: '#ba1a1a'
  on-error: '#ffffff'
  error-container: '#ffdad6'
  on-error-container: '#93000a'
  primary-fixed: '#7efabd'
  primary-fixed-dim: '#60dda2'
  on-primary-fixed: '#002113'
  on-primary-fixed-variant: '#005235'
  secondary-fixed: '#ffdbc8'
  secondary-fixed-dim: '#ffb68a'
  on-secondary-fixed: '#321300'
  on-secondary-fixed-variant: '#743400'
  tertiary-fixed: '#dbe4e1'
  tertiary-fixed-dim: '#bfc8c5'
  on-tertiary-fixed: '#151d1b'
  on-tertiary-fixed-variant: '#404946'
  background: '#fbf9f7'
  on-background: '#1b1c1b'
  surface-variant: '#e4e2e0'
typography:
  headline-xl:
    fontFamily: Inter
    fontSize: 32px
    fontWeight: '700'
    lineHeight: '1.2'
    letterSpacing: -0.02em
  headline-lg:
    fontFamily: Inter
    fontSize: 24px
    fontWeight: '600'
    lineHeight: '1.3'
    letterSpacing: -0.01em
  headline-md:
    fontFamily: Inter
    fontSize: 20px
    fontWeight: '600'
    lineHeight: '1.4'
    letterSpacing: 0.01em
  body-lg:
    fontFamily: Inter
    fontSize: 16px
    fontWeight: '400'
    lineHeight: '1.6'
    letterSpacing: 0em
  body-md:
    fontFamily: Inter
    fontSize: 14px
    fontWeight: '400'
    lineHeight: '1.5'
    letterSpacing: 0em
  label-sm:
    fontFamily: Inter
    fontSize: 12px
    fontWeight: '600'
    lineHeight: '1'
    letterSpacing: 0.05em
  headline-lg-mobile:
    fontFamily: Inter
    fontSize: 22px
    fontWeight: '600'
    lineHeight: '1.3'
rounded:
  sm: 0.25rem
  DEFAULT: 0.5rem
  md: 0.75rem
  lg: 1rem
  xl: 1.5rem
  full: 9999px
spacing:
  base: 4px
  xs: 8px
  sm: 16px
  md: 24px
  lg: 32px
  xl: 48px
  container-margin: 32px
  gutter: 24px
---

## Brand & Style

This design system merges industrial data density with a soft, contemporary dashboard aesthetic. The personality is efficient yet approachable, emphasizing clarity through modern visual metaphors. 

The style is a hybrid of **Glassmorphism** and **Minimalism**. It utilizes frosted surfaces and subtle background blurs to create a sense of depth and hierarchy without the clutter of heavy drop shadows. The overall mood is "Luminous," characterized by airy whites, vibrant seafoam-emerald accents, and a physical sense of layering that makes complex data feel manageable and inviting.

## Colors

The palette is anchored by a bright seafoam-emerald primary hue, used for key actions and focus areas. It provides a more modern and high-contrast look compared to muted industrial greens.

- **Primary:** Bright Seafoam Emerald (#34B780) used for high-intent actions and status indicators.
- **Secondary:** Warm Orange (#F17B21) used sparingly for attention-grabbing promotions or alerts.
- **Background:** A multi-layered approach using Off-White (#FBF9F7) for the application base and Pure White (#FFFFFF) for elevated cards.
- **Surface:** Semi-transparent white (Alpha 60-80%) is used for glassmorphic sidebars and overlays.

## Typography

This design system relies exclusively on **Inter** to maintain a clean, highly legible, and neutral foundation. To inject personality, headlines utilize generous letter spacing and bold weights.

The typographic hierarchy is structured to support data-heavy environments. Labels use a slightly tighter tracking and uppercase styling to differentiate them from body copy. Numeric data should prioritize clarity, often utilizing medium weights to stand out within card components.

## Layout & Spacing

The layout follows a **Fluid Grid** model with high-density content containers. It uses a 12-column grid for desktop and a 4-column grid for mobile devices.

- **Margins:** 32px outer margins ensure the content "breathes" against the edge of the viewport.
- **Gutters:** A consistent 24px gutter is maintained between cards to allow the background blur effects to remain visible.
- **Density:** While the outer layout is airy, internal card spacing remains "Industrial," utilizing tight 8px and 16px increments to maximize data visibility without feeling cramped.

## Elevation & Depth

Depth is achieved through **Glassmorphism** rather than traditional shadows. Surfaces are defined by three distinct tiers:

1.  **Canvas:** The base layer, typically a soft light gray/off-white.
2.  **Panels:** Large container surfaces (like sidebars) using a `backdrop-filter: blur(20px)` and a 1px solid white border at 40% opacity.
3.  **Cards:** Floating elements using pure white or primary gradients. These use "Ambient Shadows"—ultra-diffused (30px-50px blur), very low opacity (5-8%) shadows that match the hue of the card's background.

The 1px white border is critical for defining edges on light backgrounds.

## Shapes

The shape language is friendly and modern, characterized by large radii that soften the industrial nature of the data. 

- **Primary Cards:** Use a 24px radius (`rounded-xl` equivalent) for a distinct "dashboard" look.
- **Interactive Elements:** Buttons and input fields use a 12px-16px radius.
- **Indicators:** Small chips and status badges utilize fully rounded (pill) shapes to distinguish them from structural containers.

## Components

### Buttons
Primary buttons feature the seafoam emerald gradient with white text. They should have a subtle inner glow (1px top border, white 20%) to enhance the tactile feel. Secondary buttons are "Ghost" style with the primary border and text.

### Cards
Cards are the core of this system. Every card must have a 1px white border (Alpha 50%) and a 24px corner radius. Glassmorphic cards should have a background color of `rgba(255, 255, 255, 0.7)`.

### Side Navigation
The sidebar is a translucent vertical panel. Active states are indicated by a high-contrast pill shape that "floats" over the blurred background, using the primary seafoam emerald color (#34B780).

### Data Inputs
Inputs should be clean, using the secondary background color for the field with a 1px border that turns bright seafoam on focus. Use Inter-Medium for input text to ensure clarity.

### Chips & Badges
Small, rounded indicators used for status or categories. Use light tints of the primary color for backgrounds with full-strength primary color for the text to ensure accessibility.