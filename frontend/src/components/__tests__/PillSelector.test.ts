import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import PillSelector from '../PillSelector.vue'

describe('PillSelector', () => {
  it('renders one button per option', () => {
    const wrapper = mount(PillSelector, {
      props: { modelValue: [], options: ['a', 'b', 'c'] },
    })
    const buttons = wrapper.findAll('button')
    expect(buttons).toHaveLength(3)
    expect(buttons.map((b) => b.text())).toEqual(['a', 'b', 'c'])
  })

  it('marks selected options with aria-pressed=true', () => {
    const wrapper = mount(PillSelector, {
      props: { modelValue: ['b'], options: ['a', 'b', 'c'] },
    })
    const buttons = wrapper.findAll('button')
    expect(buttons[0].attributes('aria-pressed')).toBe('false')
    expect(buttons[1].attributes('aria-pressed')).toBe('true')
    expect(buttons[2].attributes('aria-pressed')).toBe('false')
  })

  it('emits update:modelValue with the option added when clicked unselected', async () => {
    const wrapper = mount(PillSelector, {
      props: { modelValue: ['a'], options: ['a', 'b', 'c'] },
    })
    await wrapper.findAll('button')[1].trigger('click')
    expect(wrapper.emitted('update:modelValue')).toEqual([[['a', 'b']]])
  })

  it('emits update:modelValue with the option removed when clicked selected', async () => {
    const wrapper = mount(PillSelector, {
      props: { modelValue: ['a', 'b'], options: ['a', 'b', 'c'] },
    })
    await wrapper.findAll('button')[0].trigger('click')
    expect(wrapper.emitted('update:modelValue')).toEqual([[['b']]])
  })

  it('preserves the order of existing selections when adding a new one', async () => {
    const wrapper = mount(PillSelector, {
      props: { modelValue: ['c', 'a'], options: ['a', 'b', 'c'] },
    })
    await wrapper.findAll('button')[1].trigger('click')
    expect(wrapper.emitted('update:modelValue')).toEqual([[['c', 'a', 'b']]])
  })

  it('renders nothing when options list is empty', () => {
    const wrapper = mount(PillSelector, {
      props: { modelValue: [], options: [] },
    })
    expect(wrapper.findAll('button')).toHaveLength(0)
  })

  it('uses button type="button" to prevent form submission when used inside a form', () => {
    const wrapper = mount(PillSelector, {
      props: { modelValue: [], options: ['a'] },
    })
    expect(wrapper.find('button').attributes('type')).toBe('button')
  })
})
