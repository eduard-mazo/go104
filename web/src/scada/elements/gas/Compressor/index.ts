import { markRaw } from 'vue'
import Compressor from '@/components/scada/elements/Compressor.vue'
import type { ElementDescriptor } from '../../../types'
import { digitalDefault } from '../../_shared'

export const CompressorDescriptor: ElementDescriptor = {
  kind:  'compressor',
  label: 'Compressor',
  group: 'gas',
  component: markRaw(Compressor),

  defaultConfig: () => digitalDefault(),
  defaultSize:   () => ({ w: 60, h: 60 }),

  ports: [
    { id: 'l', rx: 0, ry: 0.5 },
    { id: 'r', rx: 1, ry: 0.5 },
  ],

  palette: {
    pw: 24, ph: 24, vb: '0 0 24 24',
    svg: `<circle cx="12" cy="12" r="10" fill="#374151" stroke="#475569" stroke-width="1"/>
<path d="M6 9 L10 12 L6 15 M14 9 L18 12 L14 15" fill="none" stroke="white" stroke-width="1.2" stroke-linecap="round" stroke-linejoin="round" opacity="0.8"/>`,
  },
}
