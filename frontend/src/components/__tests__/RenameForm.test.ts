import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import RenameForm from '../RenameForm.vue'

describe('RenameForm', () => {
  it('renders with the initial value', () => {
    const wrapper = mount(RenameForm, { props: { value: 'Original' } })
    const input = wrapper.find('input')
    expect(input.element.value).toBe('Original')
  })

  it('renders "Settings" as the default section label', () => {
    const wrapper = mount(RenameForm, { props: { value: 'Original' } })
    expect(wrapper.find('h3').text()).toBe('Settings')
  })

  it('accepts a custom label prop', () => {
    const wrapper = mount(RenameForm, { props: { value: 'Original', label: 'Rename' } })
    expect(wrapper.find('h3').text()).toBe('Rename')
  })

  it('omits the label heading when label is an empty string', () => {
    const wrapper = mount(RenameForm, { props: { value: 'Original', label: '' } })
    expect(wrapper.find('h3').exists()).toBe(false)
  })

  it('disables Cancel and Save when the input matches the value', () => {
    const wrapper = mount(RenameForm, { props: { value: 'Original' } })
    const buttons = wrapper.findAll('button')
    for (const btn of buttons) {
      expect(btn.attributes('disabled')).toBeDefined()
    }
  })

  it('enables Cancel and Save once the input is edited', async () => {
    const wrapper = mount(RenameForm, { props: { value: 'Original' } })
    await wrapper.find('input').setValue('Changed')
    const buttons = wrapper.findAll('button')
    for (const btn of buttons) {
      expect(btn.attributes('disabled')).toBeUndefined()
    }
  })

  it('emits save with the new value on submit', async () => {
    const wrapper = mount(RenameForm, { props: { value: 'Original' } })
    await wrapper.find('input').setValue('Changed')
    await wrapper.find('form').trigger('submit')
    expect(wrapper.emitted('save')).toEqual([['Changed']])
  })

  it('does not emit save when input matches the value', async () => {
    const wrapper = mount(RenameForm, { props: { value: 'Original' } })
    await wrapper.find('form').trigger('submit')
    expect(wrapper.emitted('save')).toBeUndefined()
  })

  it('reverts the input to value when Cancel is clicked', async () => {
    const wrapper = mount(RenameForm, { props: { value: 'Original' } })
    const input = wrapper.find('input')
    await input.setValue('Changed')
    expect(input.element.value).toBe('Changed')

    const cancel = wrapper.findAll('button').find((b) => b.text() === 'Cancel')!
    await cancel.trigger('click')
    expect(input.element.value).toBe('Original')
  })

  it('shows "Saving..." on the submit button when saving is true', () => {
    const wrapper = mount(RenameForm, { props: { value: 'Original', saving: true } })
    const submit = wrapper.findAll('button').find((b) => b.attributes('type') === 'submit')!
    expect(submit.text()).toBe('Saving...')
    expect(submit.attributes('disabled')).toBeDefined()
  })

  it('resyncs the input when the value prop changes', async () => {
    const wrapper = mount(RenameForm, { props: { value: 'Original' } })
    await wrapper.setProps({ value: 'ChangedExternally' })
    const input = wrapper.find('input')
    expect(input.element.value).toBe('ChangedExternally')
  })

  it('after a save-and-resync cycle, Cancel and Save are disabled again', async () => {
    const wrapper = mount(RenameForm, { props: { value: 'Original' } })
    // User edits.
    await wrapper.find('input').setValue('New')
    // Parent accepts the save, updates its value, re-renders the component with the new value.
    await wrapper.setProps({ value: 'New' })
    // Input should still show "New" and buttons should be disabled.
    expect(wrapper.find('input').element.value).toBe('New')
    for (const btn of wrapper.findAll('button')) {
      expect(btn.attributes('disabled')).toBeDefined()
    }
  })
})
