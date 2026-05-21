import { markRaw } from 'vue'
import Motor from '@/components/scada/elements/Motor.vue'
import type { ElementDescriptor } from '../../../types'
import { digitalDefault } from '../../_shared'

export const MotorDescriptor: ElementDescriptor = {
  kind:  'motor',
  label: 'Motor',
  group: 'electrical',
  component: markRaw(Motor),

  defaultConfig: () => digitalDefault(),
  defaultSize:   () => ({ w: 60, h: 60 }),

  ports: [
    { id: 'l', rx: 0,   ry: 0.5 },
    { id: 'r', rx: 1,   ry: 0.5 },
    { id: 't', rx: 0.5, ry: 0   },
  ],

  palette: {
    pw: 24, ph: 24, vb: '0 0 24 24',
    svg: `<circle cx="12" cy="12" r="10" fill="#374151" stroke="#475569" stroke-width="1"/>
<text x="12" y="13" text-anchor="middle" dominant-baseline="middle" font-size="9" fill="white" font-family="monospace" font-weight="700">M</text>`,
  },
}
