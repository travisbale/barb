import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import DataTable from '../DataTable.vue'
import DataTableRow from '../DataTableRow.vue'

describe('DataTable', () => {
  it('renders column headers', () => {
    const columns = [{ label: 'Name' }, { label: 'Email' }]
    const wrapper = mount(DataTable, { props: { columns } })
    const headers = wrapper.findAll('th')
    expect(headers).toHaveLength(2)
    expect(headers[0].text()).toBe('Name')
    expect(headers[1].text()).toBe('Email')
  })

  it('applies width class to columns', () => {
    const columns = [{ label: 'Actions', width: 'w-16' }]
    const wrapper = mount(DataTable, { props: { columns } })
    expect(wrapper.find('th').classes()).toContain('w-16')
  })

  it('renders slot content in tbody', () => {
    const columns = [{ label: 'Name' }]
    const wrapper = mount(DataTable, {
      props: { columns },
      slots: { default: '<tr><td>Row content</td></tr>' },
    })
    expect(wrapper.find('tbody').text()).toContain('Row content')
  })
})

describe('DataTableRow', () => {
  it('renders slot content', () => {
    const wrapper = mount(DataTableRow, {
      slots: { default: '<td>Cell</td>' },
    })
    expect(wrapper.text()).toContain('Cell')
  })

  it('applies cursor-pointer when clickable', () => {
    const wrapper = mount(DataTableRow, { props: { clickable: true } })
    expect(wrapper.find('tr').classes()).toContain('cursor-pointer')
  })

  it('does not apply cursor-pointer by default', () => {
    const wrapper = mount(DataTableRow)
    expect(wrapper.find('tr').classes()).not.toContain('cursor-pointer')
  })

  it('sets animation delay based on index', () => {
    const wrapper = mount(DataTableRow, { props: { index: 3 } })
    expect(wrapper.find('tr').attributes('style')).toContain('animation-delay: 60ms')
  })
})
