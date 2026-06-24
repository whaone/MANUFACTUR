import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/svelte'
import { createRawSnippet } from 'svelte'
import Badge from './Badge.svelte'

// Helper: build a children snippet rendering plain text (Svelte 5 requires a snippet).
const label = (text: string) =>
  createRawSnippet(() => ({ render: () => `<span>${text}</span>` }))

describe('Badge', () => {
  it('renders the children content', () => {
    render(Badge, { children: label('Active') })
    expect(screen.getByText('Active')).toBeInTheDocument()
  })

  it('applies the variant classes', () => {
    const { container } = render(Badge, { variant: 'success', children: label('OK') })
    const badge = container.querySelector('span')
    expect(badge?.className).toContain('text-status-success')
    expect(badge?.className).toContain('rounded-full')
  })

  it('defaults to the neutral variant', () => {
    const { container } = render(Badge, { children: label('Draft') })
    const badge = container.querySelector('span')
    expect(badge?.className).toContain('text-on-surface-variant')
  })

  it('applies the md size when requested', () => {
    const { container } = render(Badge, { size: 'md', children: label('Big') })
    const badge = container.querySelector('span')
    expect(badge?.className).toContain('text-sm')
  })
})
