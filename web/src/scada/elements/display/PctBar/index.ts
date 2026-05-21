import { markRaw } from 'vue'
import PctBar from '@/components/scada/elements/PctBar.vue'
import type { ElementDescriptor } from '../../../types'

export const PctBarDescriptor: ElementDescriptor = {
  // Sits in "fluid" palette group historically — keep label group there for UI parity.
  kind:  'pct_bar',
  label: '% Bar',
  group: 'fluid',
  component: markRaw(PctBar),

  defaultConfig: () => ({
    min:        0,
    max:        100,
    color_fill: '#22d3ee',
    color_bg:   '#1e293b',
    unit:       '%',
    label:      '',
    vertical:   true,
  }),
  defaultSize: () => ({ w: 56, h: 120 }),

  ports: [
    { id: 't', rx: 0.5, ry: 0 },
    { id: 'b', rx: 0.5, ry: 1 },
  ],

  palette: {
    pw: 16, ph: 32, vb: '0 0 16 32',
    svg: `<rect x="3" y="2" width="10" height="28" fill="#1e293b" rx="1" stroke="#475569" stroke-width="1"/>
<rect x="3" y="16" width="10" height="14" fill="#22d3ee" rx="1"/>`,
  },
}
