import { markRaw } from 'vue'
import Tank from '@/components/scada/elements/Tank.vue'
import type { ElementDescriptor } from '../../../types'
import { gaugeDefault } from '../../_shared'

export const TankDescriptor: ElementDescriptor = {
  kind:  'tank',
  label: 'Tank',
  group: 'fluid',
  component: markRaw(Tank),

  defaultConfig: () => gaugeDefault({ unit: 'm³', alarm_high: 90 }),
  defaultSize:   () => ({ w: 64, h: 140 }),

  ports: [
    { id: 't', rx: 0.5, ry: 0   },
    { id: 'b', rx: 0.5, ry: 1   },
    { id: 'l', rx: 0,   ry: 0.5 },
    { id: 'r', rx: 1,   ry: 0.5 },
  ],

  palette: {
    pw: 18, ph: 32, vb: '0 0 18 32',
    svg: `<rect x="2" y="2" width="14" height="28" fill="#0f1923" rx="1" stroke="#475569" stroke-width="1"/>
<line x1="3" y1="20" x2="15" y2="20" stroke="#22d3ee" stroke-width="1.5"/>`,
  },
}
