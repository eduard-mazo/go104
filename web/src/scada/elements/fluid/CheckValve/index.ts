import { markRaw } from 'vue'
import CheckValve from '@/components/scada/elements/CheckValve.vue'
import type { ElementDescriptor } from '../../../types'
import { digitalDefault } from '../../_shared'

export const CheckValveDescriptor: ElementDescriptor = {
  kind:  'check_valve',
  label: 'Check valve',
  group: 'fluid',
  component: markRaw(CheckValve),

  defaultConfig: () => digitalDefault({ color_on: '#22d3ee', color_off: '#1e293b' }),
  defaultSize:   () => ({ w: 40, h: 32 }),

  ports: [
    { id: 'l', rx: 0, ry: 0.5 },
    { id: 'r', rx: 1, ry: 0.5 },
  ],

  palette: {
    pw: 30, ph: 20, vb: '0 0 30 20',
    svg: `<polygon points="0,0 29,10 0,20" fill="#374151" stroke="#475569" stroke-width="1"/>
<line x1="29" y1="0" x2="29" y2="20" stroke="#475569" stroke-width="2.5"/>`,
  },
}
