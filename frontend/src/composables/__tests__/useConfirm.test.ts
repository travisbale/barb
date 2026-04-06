import { describe, it, expect } from 'vitest'
import { useConfirm } from '../useConfirm'

describe('useConfirm', () => {
  it('starts hidden', () => {
    const { visible } = useConfirm()
    expect(visible.value).toBe(false)
  })

  it('confirm sets visible, message, and label', () => {
    const { confirm, visible, message, confirmLabel, confirmVariant } = useConfirm()
    confirm('Are you sure?', { label: 'Yes', variant: 'primary' })
    expect(visible.value).toBe(true)
    expect(message.value).toBe('Are you sure?')
    expect(confirmLabel.value).toBe('Yes')
    expect(confirmVariant.value).toBe('primary')
  })

  it('confirm uses defaults for label and variant', () => {
    const { confirm, confirmLabel, confirmVariant } = useConfirm()
    confirm('Delete this?')
    expect(confirmLabel.value).toBe('Delete')
    expect(confirmVariant.value).toBe('danger')
  })

  it('respond(true) resolves the promise with true', async () => {
    const { confirm, respond } = useConfirm()
    const result = confirm('Delete?')
    respond(true)
    expect(await result).toBe(true)
  })

  it('respond(false) resolves the promise with false', async () => {
    const { confirm, respond } = useConfirm()
    const result = confirm('Delete?')
    respond(false)
    expect(await result).toBe(false)
  })

  it('respond hides the dialog', () => {
    const { confirm, respond, visible } = useConfirm()
    confirm('Delete?')
    expect(visible.value).toBe(true)
    respond(false)
    expect(visible.value).toBe(false)
  })
})
