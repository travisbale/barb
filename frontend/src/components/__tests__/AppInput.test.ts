import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import AppInput from '../AppInput.vue'

describe('AppInput', () => {
  it('renders an input by default', () => {
    const wrapper = mount(AppInput, { props: { modelValue: '' } })
    expect(wrapper.find('input').exists()).toBe(true)
    expect(wrapper.find('textarea').exists()).toBe(false)
  })

  it('renders a textarea when multiline', () => {
    const wrapper = mount(AppInput, { props: { modelValue: '', multiline: true } })
    expect(wrapper.find('textarea').exists()).toBe(true)
    expect(wrapper.find('input').exists()).toBe(false)
  })

  it('renders the placeholder as a label', () => {
    const wrapper = mount(AppInput, { props: { modelValue: '', placeholder: 'Email' } })
    expect(wrapper.find('label').text()).toContain('Email')
  })

  it('shows required indicator', () => {
    const wrapper = mount(AppInput, { props: { modelValue: '', placeholder: 'Name', required: true } })
    expect(wrapper.find('label').text()).toContain('*')
  })

  it('emits update:modelValue on input', async () => {
    const wrapper = mount(AppInput, { props: { modelValue: '' } })
    await wrapper.find('input').setValue('hello')
    expect(wrapper.emitted('update:modelValue')![0]).toEqual(['hello'])
  })

  it('shows password toggle for password type', () => {
    const wrapper = mount(AppInput, { props: { modelValue: '', type: 'password' } })
    expect(wrapper.find('button').exists()).toBe(true)
  })

  it('does not show password toggle for text type', () => {
    const wrapper = mount(AppInput, { props: { modelValue: '', type: 'text' } })
    expect(wrapper.find('button').exists()).toBe(false)
  })

  it('toggles password visibility on button click', async () => {
    const wrapper = mount(AppInput, { props: { modelValue: '', type: 'password' } })
    expect(wrapper.find('input').attributes('type')).toBe('password')
    await wrapper.find('button').trigger('click')
    expect(wrapper.find('input').attributes('type')).toBe('text')
    await wrapper.find('button').trigger('click')
    expect(wrapper.find('input').attributes('type')).toBe('password')
  })

  it('uses default rows for multiline', () => {
    const wrapper = mount(AppInput, { props: { modelValue: '', multiline: true } })
    expect(wrapper.find('textarea').attributes('rows')).toBe('4')
  })

  it('uses custom rows for multiline', () => {
    const wrapper = mount(AppInput, { props: { modelValue: '', multiline: true, rows: 8 } })
    expect(wrapper.find('textarea').attributes('rows')).toBe('8')
  })
})
