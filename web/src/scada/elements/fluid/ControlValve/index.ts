import { markRaw } from 'vue'
import ControlValve from '@/components/scada/elements/ControlValve.vue'
import type { ElementDescriptor } from '../../../types'
import { digitalDefault } from '../../_shared'

export const ControlValveDescriptor: ElementDescriptor = {
  kind:  'control_valve',
  label: 'Control valve',
  group: 'fluid',
  component: markRaw(ControlValve),

  defaultConfig: () => digitalDefault(),
  defaultSize:   () => ({ w: 48, h: 48 }),

  ports: [
    { id: 'l', rx: 0, ry: 0.5 },
    { id: 'r', rx: 1, ry: 0.5 },
  ],

  palette: {
    pw: 30, ph: 20, vb: '0 0 30 20',
    svg: `<polygon points="0,0 15,10 0,20" fill="#374151" stroke="#475569" stroke-width="1"/>
<polygon points="30,0 15,10 30,20" fill="#374151" stroke="#475569" stroke-width="1"/>
<circle cx="15" cy="10" r="2" fill="#475569"/>
<line x1="15" y1="0" x2="15" y2="-4" stroke="#475569" stroke-width="1.5"/>
<circle cx="15" cy="-10" r="5" fill="none" stroke="#475569" stroke-width="1"/>`,
  },
}
