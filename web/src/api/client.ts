import axios from 'axios'

const http = axios.create({ baseURL: '/api', timeout: 10_000 })

// ── Types ──────────────────────────────────────────────────────────────────

export interface Line {
  id: number
  name: string
  host: string
  port: number
  common_address: number
  k: number
  w: number
  t1_ms: number
  t2_ms: number
  t3_ms: number
  gi_interval_s: number
  enabled: boolean
  created_at: string
  // runtime
  state: 'ACTIVE' | 'CONNECTING' | 'DISCONNECTED' | 'STOPPED'
  rx_count: number
  tx_count: number
  addr: string
}

export interface Signal {
  id: number
  line_id: number
  name: string
  ioa: number
  type_id: number
  signal_type: 'digital' | 'analog'
  unit: string
  scale: number
  offset: number
  description: string
  line_name?: string  // populated by /api/signals/all
}

export interface Datapoint {
  signal_id: number
  line_id: number
  ioa: number
  name: string
  signal_type: 'digital' | 'analog'
  type_id: number
  unit: string
  value: number
  raw_value: number
  quality: number
  quality_ok: boolean
  timestamp: string
  received_at: string
}

export interface Command {
  line_id: number
  ioa: number
  type_id: number
  value: number
  select: boolean
}

// WebSocket push messages
export interface WSDatapoint extends Datapoint { type: 'datapoint' }
export interface WSLineStatus {
  type: 'line_status'
  line_id: number
  state: string
  rx_count: number
  tx_count: number
  addr: string
}
export interface WSGIComplete  { type: 'gi_complete';   line_id: number }
export interface WSGIStart     { type: 'gi_start';      line_id: number }
export interface WSSignalStale { type: 'signal_stale';  signal_id: number; line_id: number }
export type WSMessage = WSDatapoint | WSLineStatus | WSGIComplete | WSGIStart | WSSignalStale

// ── API calls ──────────────────────────────────────────────────────────────

export const linesAPI = {
  list: ()             => http.get<Line[]>('/lines').then(r => r.data),
  create: (l: Partial<Line>) => http.post<Line>('/lines', l).then(r => r.data),
  update: (id: number, l: Partial<Line>) => http.put<Line>(`/lines/${id}`, l).then(r => r.data),
  remove: (id: number) => http.delete(`/lines/${id}`).then(r => r.data),
  start:  (id: number) => http.post(`/lines/${id}/start`).then(r => r.data),
  stop:   (id: number) => http.post(`/lines/${id}/stop`).then(r => r.data),
  gi:     (id: number) => http.post(`/lines/${id}/gi`).then(r => r.data),
}

export const signalsAPI = {
  list:   (lineID: number) => http.get<Signal[]>('/signals', { params: { line_id: lineID } }).then(r => r.data),
  create: (s: Partial<Signal>) => http.post<Signal>('/signals', s).then(r => r.data),
  update: (id: number, s: Partial<Signal>) => http.put<Signal>(`/signals/${id}`, s).then(r => r.data),
  remove: (id: number) => http.delete(`/signals/${id}`).then(r => r.data),
}

export const monitorAPI = {
  list: (lineID: number) => http.get<Datapoint[]>('/monitor', { params: { line_id: lineID } }).then(r => r.data),
}

export const commandsAPI = {
  send: (cmd: Command) => http.post('/commands', cmd).then(r => r.data),
}

// ── SCADA types ────────────────────────────────────────────────────────────

export interface ValveConfig {
  color_on:    string
  color_off:   string
  color_fault: string
  label:       string
}

export interface PctBarConfig {
  min:        number
  max:        number
  color_fill: string
  color_bg:   string
  unit:       string
  label:      string
  vertical:   boolean
}

export type ElementKind = 'valve' | 'pct_bar'

export interface ScadaElement {
  id:        string         // crypto.randomUUID()
  kind:      ElementKind
  x:         number
  y:         number
  w:         number
  h:         number
  rotation:  0 | 90 | 180 | 270
  signal_id: number | null
  config:    ValveConfig | PctBarConfig
}

export interface ScadaView {
  id:         number
  name:       string
  width:      number
  height:     number
  elements:   ScadaElement[]
  updated_at: string
}

export const scadaAPI = {
  list:   ()                      => http.get<ScadaView[]>('/scada/views').then(r => r.data),
  get:    (id: number)            => http.get<ScadaView>(`/scada/views/${id}`).then(r => r.data),
  create: (v: Partial<ScadaView>) => http.post<ScadaView>('/scada/views', v).then(r => r.data),
  update: (id: number, v: Partial<ScadaView>) =>
    http.put<ScadaView>(`/scada/views/${id}`, v).then(r => r.data),
  remove: (id: number)            => http.delete(`/scada/views/${id}`).then(r => r.data),
}

export const allSignalsAPI = {
  list: () => http.get<Signal[]>('/signals/all').then(r => r.data),
}

// ── TYPE ID catalogue ──────────────────────────────────────────────────────

export const TYPE_IDS = [
  { id: 1,  label: 'M_SP_NA_1 — Single point',          kind: 'digital' },
  { id: 3,  label: 'M_DP_NA_1 — Double point',          kind: 'digital' },
  { id: 9,  label: 'M_ME_NA_1 — Normalized (-1..1)',    kind: 'analog'  },
  { id: 11, label: 'M_ME_NB_1 — Scaled (int16)',        kind: 'analog'  },
  { id: 13, label: 'M_ME_NC_1 — Float',                 kind: 'analog'  },
  { id: 21, label: 'M_ME_ND_1 — Normalized no quality', kind: 'analog'  },
  { id: 30, label: 'M_SP_TB_1 — Single point + time',   kind: 'digital' },
  { id: 31, label: 'M_DP_TB_1 — Double point + time',   kind: 'digital' },
  { id: 34, label: 'M_ME_TA_1 — Normalized + time',     kind: 'analog'  },
  { id: 35, label: 'M_ME_TB_1 — Scaled + time',         kind: 'analog'  },
  { id: 36, label: 'M_ME_TF_1 — Float + time',          kind: 'analog'  },
]

export const CMD_TYPE_IDS = [
  { id: 45, label: 'C_SC_NA_1 — Single command (ON/OFF)' },
  { id: 46, label: 'C_DC_NA_1 — Double command (0-3)'    },
  { id: 48, label: 'C_SE_NA_1 — Setpoint normalized'     },
  { id: 49, label: 'C_SE_NB_1 — Setpoint scaled'         },
  { id: 50, label: 'C_SE_NC_1 — Setpoint float'          },
]

export function typeIDName(id: number): string {
  return TYPE_IDS.find(t => t.id === id)?.label.split(' — ')[0] ?? `TypeID ${id}`
}

export function formatValue(dp: Datapoint): string {
  if (dp.signal_type === 'digital') {
    if (dp.type_id === 3 || dp.type_id === 31) {
      const labels = ['INDET', 'OFF', 'ON', 'INDET']
      return labels[dp.value & 3] ?? String(dp.value)
    }
    return dp.value ? 'ON' : 'OFF'
  }
  return dp.value.toFixed(3) + (dp.unit ? ' ' + dp.unit : '')
}
