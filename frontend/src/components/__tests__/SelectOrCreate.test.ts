import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import SelectOrCreate from '../SelectOrCreate.vue'

const items = [
  { value: '1', label: 'Item One' },
  { value: '2', label: 'Item Two' },
]

const baseProps = {
  modelValue: '1',
  label: 'Select an item',
  items,
  emptyMessage: 'No items available.',
  createLabel: 'Create new item',
}

describe('SelectOrCreate', () => {
  it('renders dropdown when items are provided', () => {
    const wrapper = mount(SelectOrCreate, { props: baseProps })
    expect(wrapper.find('select').exists()).toBe(true)
    expect(wrapper.findAll('option')).toHaveLength(2)
  })

  it('shows empty message when no items', () => {
    const wrapper = mount(SelectOrCreate, { props: { ...baseProps, items: [] } })
    expect(wrapper.find('select').exists()).toBe(false)
    expect(wrapper.text()).toContain('No items available.')
  })

  it('shows create label in both states', () => {
    const withItems = mount(SelectOrCreate, { props: baseProps })
    expect(withItems.text()).toContain('Create new item')

    const empty = mount(SelectOrCreate, { props: { ...baseProps, items: [] } })
    expect(empty.text()).toContain('Create new item')
  })

  it('toggles to create form when create button clicked', async () => {
    const wrapper = mount(SelectOrCreate, {
      props: baseProps,
      slots: { form: '<div class="test-form">Form content</div>' },
    })
    expect(wrapper.find('.test-form').exists()).toBe(false)

    await wrapper.findAll('button').find(b => b.text().includes('Create new item'))!.trigger('click')

    expect(wrapper.find('.test-form').exists()).toBe(true)
    expect(wrapper.find('select').exists()).toBe(false)
  })

  it('toggles back from create form on cancel', async () => {
    const wrapper = mount(SelectOrCreate, {
      props: baseProps,
      slots: { form: '<div class="test-form">Form</div>' },
    })

    // Enter create mode
    await wrapper.findAll('button').find(b => b.text().includes('Create new item'))!.trigger('click')
    expect(wrapper.find('.test-form').exists()).toBe(true)

    // Cancel
    await wrapper.findAll('button').find(b => b.text() === 'Cancel')!.trigger('click')
    expect(wrapper.find('.test-form').exists()).toBe(false)
    expect(wrapper.find('select').exists()).toBe(true)
  })

  it('emits create event when Create button clicked', async () => {
    const wrapper = mount(SelectOrCreate, {
      props: baseProps,
      slots: { form: '<div>Form</div>' },
    })

    await wrapper.findAll('button').find(b => b.text().includes('Create new item'))!.trigger('click')
    await wrapper.findAll('button').find(b => b.text() === 'Create')!.trigger('click')

    expect(wrapper.emitted('create')).toHaveLength(1)
  })

  it('disables Create button when loading', async () => {
    const wrapper = mount(SelectOrCreate, {
      props: { ...baseProps, loading: true },
      slots: { form: '<div>Form</div>' },
    })

    await wrapper.findAll('button').find(b => b.text().includes('Create new item'))!.trigger('click')
    const createBtn = wrapper.findAll('button').find(b => b.text() === 'Create')!
    expect(createBtn.attributes('disabled')).toBeDefined()
  })

  it('renders heading slot', () => {
    const wrapper = mount(SelectOrCreate, {
      props: baseProps,
      slots: { heading: '<h6>My Heading</h6>' },
    })
    expect(wrapper.text()).toContain('My Heading')
  })

  it('renders detail slot when items present', () => {
    const wrapper = mount(SelectOrCreate, {
      props: baseProps,
      slots: { detail: '<div class="test-detail">Detail content</div>' },
    })
    expect(wrapper.find('.test-detail').exists()).toBe(true)
  })

  it('exposes creating state', async () => {
    const wrapper = mount(SelectOrCreate, { props: baseProps })
    expect(wrapper.vm.creating).toBe(false)

    await wrapper.findAll('button').find(b => b.text().includes('Create new item'))!.trigger('click')
    expect(wrapper.vm.creating).toBe(true)
  })
})
