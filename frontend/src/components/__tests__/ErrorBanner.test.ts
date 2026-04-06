import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import ErrorBanner from '../ErrorBanner.vue'

describe('ErrorBanner', () => {
  it('renders nothing when message is empty', () => {
    const wrapper = mount(ErrorBanner, { props: { message: '' } })
    expect(wrapper.find('div').exists()).toBe(false)
  })

  it('renders the error message', () => {
    const wrapper = mount(ErrorBanner, { props: { message: 'Something went wrong' } })
    expect(wrapper.text()).toContain('Something went wrong')
  })

  it('applies danger styling', () => {
    const wrapper = mount(ErrorBanner, { props: { message: 'Error' } })
    expect(wrapper.find('div').classes()).toContain('border-danger/30')
  })
})
