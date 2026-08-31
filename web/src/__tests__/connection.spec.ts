import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { createTestingPinia } from '@pinia/testing'
import { createRouter, createWebHistory } from 'vue-router'
import CloudConnection from '../components/CloudConnection.vue'
import SetupGuide from '../components/SetupGuide.vue'

vi.mock('../api/client', () => ({
  getConnectionStatus: vi.fn(),
}))

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', name: 'dashboard', component: { template: '<div>Dashboard</div>' } },
    { path: '/setup/aws', name: 'setup-aws', component: { template: '<div>Setup</div>' } },
  ],
})

describe('CloudConnection', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('shows loading state initially', async () => {
    const { getConnectionStatus } = await import('../api/client')
    vi.mocked(getConnectionStatus).mockReturnValue(new Promise(() => {}))

    const wrapper = mount(CloudConnection, {
      global: {
        plugins: [createTestingPinia(), router],
      },
    })

    expect(wrapper.text()).toContain('Checking connection...')
  })

  it('shows connected state with account info', async () => {
    const { getConnectionStatus } = await import('../api/client')
    vi.mocked(getConnectionStatus).mockResolvedValue({
      state: 'connected',
      accountId: '123456789012',
      arn: 'arn:aws:iam::123456789012:role/WeaveLensScanner',
      region: 'us-east-1',
      credentialSource: 'profile',
      message: '',
    })

    const wrapper = mount(CloudConnection, {
      global: {
        plugins: [createTestingPinia(), router],
      },
    })

    await vi.waitFor(() => {
      expect(wrapper.text()).toContain('Connected')
      expect(wrapper.text()).toContain('123456789012')
      expect(wrapper.text()).toContain('us-east-1')
    })
  })

  it('shows not connected state with setup link', async () => {
    const { getConnectionStatus } = await import('../api/client')
    vi.mocked(getConnectionStatus).mockResolvedValue({
      state: 'not_connected',
      accountId: '',
      arn: '',
      region: '',
      credentialSource: '',
      message: 'No credentials found',
    })

    const wrapper = mount(CloudConnection, {
      global: {
        plugins: [createTestingPinia(), router],
      },
    })

    await vi.waitFor(() => {
      expect(wrapper.text()).toContain('Not Connected')
      expect(wrapper.text()).toContain('View Setup Guide')
    })
  })

  it('shows error state with setup link', async () => {
    const { getConnectionStatus } = await import('../api/client')
    vi.mocked(getConnectionStatus).mockResolvedValue({
      state: 'authentication_error',
      accountId: '',
      arn: '',
      region: '',
      credentialSource: '',
      message: 'Invalid credentials',
    })

    const wrapper = mount(CloudConnection, {
      global: {
        plugins: [createTestingPinia(), router],
      },
    })

    await vi.waitFor(() => {
      expect(wrapper.text()).toContain('Authentication Error')
      expect(wrapper.text()).toContain('Invalid credentials')
      expect(wrapper.text()).toContain('View Setup Guide')
    })
  })

  it('shows error message when API fails', async () => {
    const { getConnectionStatus } = await import('../api/client')
    vi.mocked(getConnectionStatus).mockRejectedValue(new Error('Network error'))

    const wrapper = mount(CloudConnection, {
      global: {
        plugins: [createTestingPinia(), router],
      },
    })

    await vi.waitFor(() => {
      expect(wrapper.text()).toContain('Network error')
      expect(wrapper.text()).toContain('Retry')
    })
  })
})

describe('SetupGuide', () => {
  it('renders setup guide with all sections', () => {
    const wrapper = mount(SetupGuide, {
      global: {
        plugins: [createTestingPinia(), router],
      },
    })

    expect(wrapper.text()).toContain('AWS Setup Guide')
    expect(wrapper.text()).toContain('Local Development')
    expect(wrapper.text()).toContain('Cross-Account Access')
    expect(wrapper.text()).toContain('Required IAM Permissions')
    expect(wrapper.text()).toContain('Troubleshooting')
  })

  it('renders IAM permissions for all services', () => {
    const wrapper = mount(SetupGuide, {
      global: {
        plugins: [createTestingPinia(), router],
      },
    })

    expect(wrapper.text()).toContain('DescribeVpcs')
    expect(wrapper.text()).toContain('DescribeDBInstances')
    expect(wrapper.text()).toContain('DescribeLoadBalancers')
    expect(wrapper.text()).toContain('GetCallerIdentity')
  })

  it('renders back to dashboard link', () => {
    const wrapper = mount(SetupGuide, {
      global: {
        plugins: [createTestingPinia(), router],
      },
    })

    expect(wrapper.text()).toContain('Back to Dashboard')
  })

  it('renders check connection button', () => {
    const wrapper = mount(SetupGuide, {
      global: {
        plugins: [createTestingPinia(), router],
      },
    })

    expect(wrapper.text()).toContain('Check Connection')
  })
})
