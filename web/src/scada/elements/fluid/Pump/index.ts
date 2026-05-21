import { markRaw } from 'vue'
import Pump from '@/components/scada/elements/Pump.vue'
import type { ElementDescriptor } from '../../../types'

export const PumpDescriptor: ElementDescriptor = {
  kind:  'pump',
  label: 'Pump',
  group: 'fluid',
  component: markRaw(Pump),

  defaultConfig: () => ({
    color_on:    '#22c55e',
    color_off:   '#ef4444',
    color_fault: '#f59e0b',
    label:       '',
  }),
  defaultSize: () => ({ w: 60, h: 60 }),

  ports: [
    { id: 'in',  rx: 0,   ry: 0.5 },
    { id: 'out', rx: 0.5, ry: 0   },
  ],

  palette: {
    pw: 24, ph: 24, vb: '0 0 24 24',
    svg: `<circle cx="12" cy="12" r="10" fill="#374151" stroke="#475569" stroke-width="1"/>
<polygon points="12,7 19,16.5 5,16.5" fill="none" stroke="white" stroke-width="1.2" stroke-linejoin="round" opacity="0.7"/>`,
  },
}
