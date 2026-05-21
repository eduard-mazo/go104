import { markRaw } from 'vue'
import TempTransmitter from '@/components/scada/elements/TempTransmitter.vue'
import type { ElementDescriptor } from '../../../types'
import { gaugeDefault } from '../../_shared'

export const TempTxDescriptor: ElementDescriptor = {
  kind:  'temp_tx',
  label: 'Temp TX',
  group: 'instrument',
  component: markRaw(TempTransmitter),

  defaultConfig: () => gaugeDefault({ unit: '°C', alarm_high: 80 }),
  defaultSize:   () => ({ w: 56, h: 56 }),

  ports: [
    { id: 'b', rx: 0.5, ry: 1 },
  ],

  palette: {
    pw: 24, ph: 24, vb: '0 0 24 24',
    svg: `<circle cx="12" cy="12" r="10" fill="#0f1923" stroke="#475569" stroke-width="1"/>
<line x1="2" y1="12" x2="22" y2="12" stroke="#475569" stroke-width="0.8" opacity="0.5"/>
<text x="12" y="10" text-anchor="middle" dominant-baseline="middle" font-size="5.5" fill="#22d3ee" font-family="monospace" font-weight="700">TT</text>
<text x="12" y="16" text-anchor="middle" dominant-baseline="middle" font-size="5" fill="#22d3ee" font-family="monospace">—</text>
<rect x="10" y="22" width="4" height="6" fill="#64748b" rx="1"/>`,
  },
}
