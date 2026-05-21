import { markRaw } from 'vue'
import PipeSegment from '@/components/scada/elements/PipeSegment.vue'
import type { ElementDescriptor } from '../../../types'

export const PipeSegmentDescriptor: ElementDescriptor = {
  kind:  'pipe_segment',
  label: 'Pipe',
  group: 'fluid',
  component: markRaw(PipeSegment),

  defaultConfig: () => ({ style: 'process', color: '#64748b', label: '' }),
  defaultSize:   () => ({ w: 120, h: 16 }),

  ports: [
    { id: 'l', rx: 0, ry: 0.5 },
    { id: 'r', rx: 1, ry: 0.5 },
  ],

  palette: {
    pw: 36, ph: 10, vb: '0 0 36 10',
    svg: `<line x1="0" y1="5" x2="36" y2="5" stroke="#64748b" stroke-width="4" stroke-linecap="round"/>`,
  },
}
