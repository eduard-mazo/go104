import { markRaw } from 'vue'
import ValveSVG from '@/components/scada/elements/ValveSVG.vue'
import type { ElementDescriptor } from '../../../types'
import { digitalDefault } from '../../_shared'

export const GateValveDescriptor: ElementDescriptor = {
  kind:  'valve',
  label: 'Gate valve',
  group: 'fluid',
  component: markRaw(ValveSVG),

  defaultConfig: () => digitalDefault(),
  defaultSize:   () => ({ w: 60, h: 60 }),

  ports: [
    { id: 'l', rx: 0, ry: 0.5 },
    { id: 'r', rx: 1, ry: 0.5 },
  ],

  palette: {
    pw: 30, ph: 20, vb: '0 0 30 20',
    svg: `<polygon points="0,0 15,10 0,20" fill="#374151" stroke="#475569" stroke-width="1"/>
<polygon points="30,0 15,10 30,20" fill="#374151" stroke="#475569" stroke-width="1"/>`,
  },
}
