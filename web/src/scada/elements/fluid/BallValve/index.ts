import { markRaw } from 'vue'
import BallValve from '@/components/scada/elements/BallValve.vue'
import type { ElementDescriptor } from '../../../types'
import { digitalDefault } from '../../_shared'

export const BallValveDescriptor: ElementDescriptor = {
  kind:  'ball_valve',
  label: 'Ball valve',
  group: 'fluid',
  component: markRaw(BallValve),

  defaultConfig: () => digitalDefault(),
  defaultSize:   () => ({ w: 60, h: 60 }),

  ports: [
    { id: 'l', rx: 0, ry: 0.5 },
    { id: 'r', rx: 1, ry: 0.5 },
  ],

  palette: {
    pw: 28, ph: 22, vb: '0 0 28 22',
    svg: `<circle cx="14" cy="11" r="9" fill="#374151" stroke="#475569" stroke-width="1"/>
<line x1="5" y1="11" x2="23" y2="11" stroke="#0a0e14" stroke-width="5" stroke-linecap="round"/>`,
  },
}
