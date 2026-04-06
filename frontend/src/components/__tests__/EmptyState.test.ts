import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import EmptyState from '../EmptyState.vue'

describe('EmptyState', () => {
  it('renders the message', () => {
    const wrapper = mount(EmptyState, { props: { message: 'Nothing here yet.' } })
    expect(wrapper.text()).toBe('Nothing here yet.')
  })

  it('applies dashed border styling', () => {
    const wrapper = mount(EmptyState, { props: { message: 'Empty' } })
    expect(wrapper.find('div').classes()).toContain('border-dashed')
  })
})
