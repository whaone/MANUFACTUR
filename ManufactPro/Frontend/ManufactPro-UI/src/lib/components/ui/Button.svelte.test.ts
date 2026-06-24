import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/svelte'
import { createRawSnippet } from 'svelte'
import Button from './Button.svelte'

const label = (text: string) =>
  createRawSnippet(() => ({ render: () => `<span>${text}</span>` }))

describe('Button', () => {
  it('renders children and a button element', () => {
    render(Button, { children: label('Save') })
    const btn = screen.getByRole('button')
    expect(btn).toBeInTheDocument()
    expect(btn).toHaveTextContent('Save')
  })

  it('fires onclick when clicked', async () => {
    const onclick = vi.fn()
    render(Button, { onclick, children: label('Go') })
    await fireEvent.click(screen.getByRole('button'))
    expect(onclick).toHaveBeenCalledTimes(1)
  })

  it('renders the disabled attribute when disabled', () => {
    render(Button, { disabled: true, children: label('Nope') })
    expect(screen.getByRole('button')).toBeDisabled()
  })

  it('shows the loading spinner', () => {
    const { container } = render(Button, { loading: true, children: label('Loading') })
    expect(container.querySelector('svg.animate-spin')).toBeInTheDocument()
  })

  it('applies the danger variant classes', () => {
    render(Button, { variant: 'danger', children: label('Delete') })
    expect(screen.getByRole('button').className).toContain('bg-error')
  })

  it('sets the button type', () => {
    render(Button, { type: 'submit', children: label('Submit') })
    expect(screen.getByRole('button')).toHaveAttribute('type', 'submit')
  })
})
