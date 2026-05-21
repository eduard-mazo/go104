import { markRaw } from 'vue'
import CircuitBreaker from '@/components/scada/elements/CircuitBreaker.vue'
import type { ElementDescriptor } from '../../../types'
import { digitalDefault } from '../../_shared'

export const CircuitBreakerDescriptor: ElementDescriptor = {
  kind:  'circuit_breaker',
  label: 'Breaker',
  group: 'electrical',
  component: markRaw(CircuitBreaker),

  defaultConfig: () => digitalDefault({ color_on: '#22c55e', color_off: '#64748b' }),
  defaultSize:   () => ({ w: 72, h: 40 }),

  ports: [
    { id: 'l', rx: 0, ry: 0.5 },
    { id: 'r', rx: 1, ry: 0.5 },
  ],

  palette: {
    pw: 32, ph: 18, vb: '0 0 32 18',
    svg: `<line x1="0" y1="9" x2="9" y2="9" stroke="#475569" stroke-width="2"/>
<rect x="9" y="3" width="14" height="12" fill="#374151" stroke="#475569" stroke-width="1" rx="1"/>
<line x1="11" y1="9" x2="21" y2="9" stroke="white" stroke-width="1.5"/>
<line x1="23" y1="9" x2="32" y2="9" stroke="#475569" stroke-width="2"/>`,
  },
}
