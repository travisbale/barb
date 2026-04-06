import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import SettingsSection from '../SettingsSection.vue'

describe('SettingsSection', () => {
  it('renders the label', () => {
    const wrapper = mount(SettingsSection, { props: { label: 'Template' } })
    expect(wrapper.text()).toContain('Template')
  })

  it('shows summary slot when not expanded', () => {
    const wrapper = mount(SettingsSection, {
      props: { label: 'SMTP' },
      slots: { summary: '<div>Gmail SMTP</div>' },
    })
    expect(wrapper.text()).toContain('Gmail SMTP')
  })

  it('hides Change button when not editable', () => {
    const wrapper = mount(SettingsSection, { props: { label: 'Template' } })
    expect(wrapper.text()).not.toContain('Change')
  })

  it('shows Change button when editable and not expanded', () => {
    const wrapper = mount(SettingsSection, { props: { label: 'Template', editable: true } })
    expect(wrapper.text()).toContain('Change')
  })

  it('hides Change button when expanded', () => {
    const wrapper = mount(SettingsSection, { props: { label: 'Template', editable: true, expanded: true } })
    expect(wrapper.findAll('button').map(b => b.text())).not.toContain('Change')
  })

  it('shows editor slot when expanded', () => {
    const wrapper = mount(SettingsSection, {
      props: { label: 'Template', expanded: true },
      slots: { editor: '<div>Editor form</div>' },
    })
    expect(wrapper.text()).toContain('Editor form')
  })

  it('shows Save and Cancel buttons when expanded', () => {
    const wrapper = mount(SettingsSection, { props: { label: 'Template', expanded: true } })
    const buttons = wrapper.findAll('button').map(b => b.text())
    expect(buttons).toContain('Cancel')
    expect(buttons).toContain('Save')
  })

  it('shows Saving... when saving', () => {
    const wrapper = mount(SettingsSection, { props: { label: 'Template', expanded: true, saving: true } })
    expect(wrapper.text()).toContain('Saving...')
  })

  it('emits change when Change is clicked', async () => {
    const wrapper = mount(SettingsSection, { props: { label: 'Template', editable: true } })
    await wrapper.findAll('button').find(b => b.text() === 'Change')!.trigger('click')
    expect(wrapper.emitted('change')).toHaveLength(1)
  })

  it('emits cancel when Cancel is clicked', async () => {
    const wrapper = mount(SettingsSection, { props: { label: 'Template', expanded: true } })
    await wrapper.findAll('button').find(b => b.text() === 'Cancel')!.trigger('click')
    expect(wrapper.emitted('cancel')).toHaveLength(1)
  })

  it('emits save when Save is clicked', async () => {
    const wrapper = mount(SettingsSection, { props: { label: 'Template', expanded: true } })
    await wrapper.findAll('button').find(b => b.text() === 'Save')!.trigger('click')
    expect(wrapper.emitted('save')).toHaveLength(1)
  })

  it('renders create-new slot when expanded', () => {
    const wrapper = mount(SettingsSection, {
      props: { label: 'Template', expanded: true },
      slots: { 'create-new': '<button>+ New</button>' },
    })
    expect(wrapper.text()).toContain('+ New')
  })
})
