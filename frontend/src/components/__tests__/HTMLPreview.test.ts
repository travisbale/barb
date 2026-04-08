import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import HTMLPreview from '../HTMLPreview.vue'

describe('HTMLPreview', () => {
  it('renders a sandboxed iframe with srcdoc', () => {
    const wrapper = mount(HTMLPreview, { props: { srcdoc: '<p>Hello</p>' } })
    const iframe = wrapper.find('iframe')
    expect(iframe.exists()).toBe(true)
    expect(iframe.attributes('srcdoc')).toBe('<p>Hello</p>')
    expect(iframe.attributes('sandbox')).toBe('')
  })
})
