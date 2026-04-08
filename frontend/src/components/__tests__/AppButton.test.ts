import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import AppButton from '../AppButton.vue'

describe('AppButton', () => {
  it('renders slot content', () => {
    const wrapper = mount(AppButton, { slots: { default: 'Click me' } })
    expect(wrapper.text()).toBe('Click me')
  })

  it('defaults to primary variant styling', () => {
    const wrapper = mount(AppButton)
    expect(wrapper.classes()).toContain('bg-amber')
  })

  it('applies secondary variant classes', () => {
    const wrapper = mount(AppButton, { props: { variant: 'secondary' } })
    expect(wrapper.classes()).toContain('border')
    expect(wrapper.classes()).not.toContain('bg-amber')
  })

  it('applies ghost variant classes', () => {
    const wrapper = mount(AppButton, { props: { variant: 'ghost' } })
    expect(wrapper.classes()).toContain('bg-transparent')
  })

  it('applies danger variant classes', () => {
    const wrapper = mount(AppButton, { props: { variant: 'danger' } })
    expect(wrapper.classes()).toContain('hover:text-danger')
  })

  it('sets disabled attribute', () => {
    const wrapper = mount(AppButton, { props: { disabled: true } })
    expect(wrapper.attributes('disabled')).toBeDefined()
  })

  it('applies link variant classes', () => {
    const wrapper = mount(AppButton, { props: { variant: 'link' } })
    expect(wrapper.classes()).toContain('text-amber')
    expect(wrapper.classes()).toContain('text-xs')
    expect(wrapper.classes()).not.toContain('px-3')
  })

  it('emits click event', async () => {
    const wrapper = mount(AppButton)
    await wrapper.trigger('click')
    expect(wrapper.emitted('click')).toHaveLength(1)
  })
})
