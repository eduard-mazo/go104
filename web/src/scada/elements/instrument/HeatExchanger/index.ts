import { markRaw } from 'vue'
import HeatExchanger from '@/components/scada/elements/HeatExchanger.vue'
import type { ElementDescriptor } from '../../../types'
import { gaugeDefault } from '../../_shared'

export const HeatExchangerDescriptor: ElementDescriptor = {
  kind:  'heat_exchanger',
  label: 'Heat exchanger',
  group: 'instrument',
  component: markRaw(HeatExchanger),

  defaultConfig: () => gaugeDefault({ unit: '°C', alarm_high: 80 }),
  defaultSize:   () => ({ w: 100, h: 60 }),

  ports: [
    { id: 'sl', rx: 0, ry: 0.35 },
    { id: 'sr', rx: 1, ry: 0.35 },
    { id: 'tl', rx: 0, ry: 0.65 },
    { id: 'tr', rx: 1, ry: 0.65 },
  ],

  palette: {
    pw: 36, ph: 22, vb: '0 0 36 22',
    svg: `<rect x="2" y="2" width="32" height="18" fill="#0f1923" stroke="#475569" stroke-width="1" rx="2"/>
<path d="M4 7 C 12 7 12 15 18 15 S 24 7 32 7" fill="none" stroke="#22d3ee" stroke-width="1.2" stroke-linecap="round"/>`,
  },
}
