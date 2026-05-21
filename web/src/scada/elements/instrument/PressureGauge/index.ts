import { markRaw } from 'vue'
import PressureGauge from '@/components/scada/elements/PressureGauge.vue'
import type { ElementDescriptor } from '../../../types'
import { gaugeDefault } from '../../_shared'

export const PressureGaugeDescriptor: ElementDescriptor = {
  kind:  'pressure_gauge',
  label: 'Pressure gauge',
  group: 'instrument',
  component: markRaw(PressureGauge),

  defaultConfig: () => gaugeDefault({ unit: 'bar', alarm_high: 80 }),
  defaultSize:   () => ({ w: 72, h: 72 }),

  ports: [
    { id: 'b', rx: 0.5, ry: 1 },
  ],

  palette: {
    pw: 24, ph: 24, vb: '0 0 24 24',
    svg: `<circle cx="12" cy="12" r="10" fill="#0f1923" stroke="#475569" stroke-width="1"/>
<line x1="2" y1="12" x2="22" y2="12" stroke="#475569" stroke-width="0.8" opacity="0.5"/>
<text x="12" y="10" text-anchor="middle" dominant-baseline="middle" font-size="5.5" fill="#22d3ee" font-family="monospace" font-weight="700">PT</text>
<text x="12" y="16" text-anchor="middle" dominant-baseline="middle" font-size="5" fill="#22d3ee" font-family="monospace">—</text>`,
  },
}
