import { markRaw } from 'vue'
import IndicatorLamp from '@/components/scada/elements/IndicatorLamp.vue'
import type { ElementDescriptor } from '../../../types'
import { digitalDefault } from '../../_shared'

export const IndicatorLampDescriptor: ElementDescriptor = {
  kind:  'indicator_lamp',
  label: 'Lamp',
  group: 'electrical',
  component: markRaw(IndicatorLamp),

  defaultConfig: () => digitalDefault({ color_off: '#1e293b' }),
  defaultSize:   () => ({ w: 60, h: 60 }),

  ports: [
    { id: 'b', rx: 0.5, ry: 1 },
  ],

  palette: {
    pw: 22, ph: 22, vb: '0 0 22 22',
    svg: `<circle cx="11" cy="11" r="9" fill="#22c55e" stroke="#475569" stroke-width="1" opacity="0.7"/>
<line x1="5" y1="5" x2="17" y2="17" stroke="white" stroke-width="1.5" stroke-linecap="round" opacity="0.7"/>
<line x1="17" y1="5" x2="5" y2="17" stroke="white" stroke-width="1.5" stroke-linecap="round" opacity="0.7"/>`,
  },
}
