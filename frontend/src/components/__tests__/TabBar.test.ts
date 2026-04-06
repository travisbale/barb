import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import TabBar from '../TabBar.vue'

describe('TabBar', () => {
  const tabs = ['Results', 'Settings']

  it('renders all tabs', () => {
    const wrapper = mount(TabBar, { props: { tabs, modelValue: 'Results' } })
    const buttons = wrapper.findAll('button')
    expect(buttons).toHaveLength(2)
    expect(buttons[0].text()).toBe('Results')
    expect(buttons[1].text()).toBe('Settings')
  })

  it('highlights the active tab', () => {
    const wrapper = mount(TabBar, { props: { tabs, modelValue: 'Results' } })
    const buttons = wrapper.findAll('button')
    expect(buttons[0].classes()).toContain('text-amber')
    expect(buttons[1].classes()).not.toContain('text-amber')
  })

  it('emits update:modelValue when a tab is clicked', async () => {
    const wrapper = mount(TabBar, { props: { tabs, modelValue: 'Results' } })
    await wrapper.findAll('button')[1].trigger('click')
    expect(wrapper.emitted('update:modelValue')).toEqual([['Settings']])
  })
})
