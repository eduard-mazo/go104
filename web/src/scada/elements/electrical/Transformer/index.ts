import { markRaw } from 'vue'
import Transformer from '@/components/scada/elements/Transformer.vue'
import type { ElementDescriptor } from '../../../types'
import { gaugeDefault } from '../../_shared'

export const TransformerDescriptor: ElementDescriptor = {
  kind:  'transformer',
  label: 'Transformer',
  group: 'electrical',
  component: markRaw(Transformer),

  defaultConfig: () => gaugeDefault({ unit: 'kVA' }),
  defaultSize:   () => ({ w: 80, h: 80 }),

  ports: [
    { id: 'tl', rx: 0.3, ry: 0 },
    { id: 'tr', rx: 0.7, ry: 0 },
    { id: 'bl', rx: 0.3, ry: 1 },
    { id: 'br', rx: 0.7, ry: 1 },
  ],

  palette: {
    pw: 32, ph: 24, vb: '0 0 32 24',
    svg: `<circle cx="10" cy="12" r="8" fill="#0f172a" stroke="#475569" stroke-width="1"/>
<circle cx="22" cy="12" r="8" fill="#0f172a" stroke="#475569" stroke-width="1"/>`,
  },
}
