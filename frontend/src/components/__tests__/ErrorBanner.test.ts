import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import ErrorBanner from '../ErrorBanner.vue'

describe('ErrorBanner', () => {
  it('renders nothing when message is empty', () => {
    const wrapper = mount(ErrorBanner, { props: { modelValue: '' } })
    expect(wrapper.find('div').exists()).toBe(false)
  })

  it('renders the error message', () => {
    const wrapper = mount(ErrorBanner, { props: { modelValue: 'Something went wrong' } })
    expect(wrapper.text()).toContain('Something went wrong')
  })

  it('applies danger styling', () => {
    const wrapper = mount(ErrorBanner, { props: { modelValue: 'Error' } })
    expect(wrapper.find('div').classes()).toContain('border-danger/30')
  })

  it('emits empty string when dismiss is clicked', async () => {
    const wrapper = mount(ErrorBanner, { props: { modelValue: 'Error' } })
    await wrapper.find('button').trigger('click')
    expect(wrapper.emitted('update:modelValue')).toEqual([['']])
  })
})
