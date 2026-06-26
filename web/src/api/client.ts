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
  // Protocol driver: 'iec104' (default) or 'dnp3'. The dnp3_* fields apply only
  // when protocol === 'dnp3'; host/port and gi_interval_s are shared (for DNP3,
  // gi_interval_s is the integrity-poll interval).
  protocol?: 'iec104' | 'dnp3'
  dnp3_outstation_addr?: number
  dnp3_master_addr?: number
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
  // DNP3 object group ('binary' | 'analog' | 'counter' | …); empty for IEC-104
  // lines. For DNP3, `ioa` is reused as the point index within its group.
  point_type?: string
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

// Shared config for analog gauge elements (tank, flow meter, pressure gauge, transformer)
export interface GaugeConfig {
  min:         number
  max:         number
  color_fill:  string
  color_alarm: string
  alarm_high:  number   // alarm when value >= alarm_high (0 = disabled)
  unit:        string
  label:       string
}

export interface PipeConfig {
  style: 'process' | 'utility' | 'instrument'
  color: string
  label: string
}

export type ElementKind =
  | 'valve'            // gate valve (digital)
  | 'pct_bar'          // percentage bar (analog)
  | 'circuit_breaker'  // electrical circuit breaker (digital)
  | 'motor'            // electric motor (digital)
  | 'transformer'      // transformer/load (analog)
  | 'indicator_lamp'   // indicator lamp (digital)
  | 'ball_valve'       // ball valve (digital)
  | 'pump'             // centrifugal pump (digital)
  | 'tank'             // storage tank with level (analog)
  | 'flow_meter'       // flow meter (analog)
  | 'pressure_gauge'   // pressure gauge (analog)
  | 'compressor'       // gas compressor (digital)
  | 'control_valve'    // actuated control valve (digital)
  | 'check_valve'      // non-return / check valve (static/digital)
  | 'heat_exchanger'   // shell & tube heat exchanger (analog)
  | 'temp_tx'          // ISA temperature transmitter TT (analog)
  | 'pipe_segment'     // static pipe segment (no signal)

export interface ScadaElement {
  id:        string         // crypto.randomUUID()
  kind:      ElementKind
  x:         number
  y:         number
  w:         number
  h:         number
  rotation:  0 | 90 | 180 | 270
  signal_id: number | null
  config:    ValveConfig | PctBarConfig | GaugeConfig | PipeConfig
}

export type ScadaLineStyle = 'pipe_water' | 'pipe_gas' | 'wire' | 'cable'

export interface ScadaLine {
  id:           string
  from_el:      string | null  // element id (null = free endpoint)
  to_el:        string | null
  from_port?:   string         // port id on from_el (live-tracked if present)
  to_port?:     string
  from_pt:      { x: number; y: number }  // fallback / free endpoint
  to_pt:        { x: number; y: number }
  style:        ScadaLineStyle
  color:        string
  stroke_width: number
  label:        string
}

export interface ScadaView {
  id:         number
  name:       string
  width:      number
  height:     number
  elements:   ScadaElement[]
  lines:      ScadaLine[]
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

export interface HistoryPoint {
  ts:      number   // Unix epoch seconds (float)
  value:   number
  quality: number
}

export const signalHistoryAPI = {
  query: (signalId: number, from: number, to: number, limit = 2000) =>
    http.get<HistoryPoint[]>(`/signals/${signalId}/history`, {
      params: { from, to, limit },
    }).then(r => r.data),
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

// ── DNP3 object groups (point types) ─────────────────────────────────────────
// What a DNP3 master reads from an outstation. `kind` derives the digital/analog
// classification go104 stores (mirrors internal/dnp3line measValue).
export const POINT_TYPES = [
  { value: 'binary',               label: 'Binary Input',         kind: 'digital' },
  { value: 'double_bit_binary',    label: 'Double-bit Binary',    kind: 'digital' },
  { value: 'binary_output_status', label: 'Binary Output Status', kind: 'digital' },
  { value: 'counter',              label: 'Counter',              kind: 'analog'  },
  { value: 'frozen_counter',       label: 'Frozen Counter',       kind: 'analog'  },
  { value: 'analog',               label: 'Analog Input',         kind: 'analog'  },
  { value: 'analog_output_status', label: 'Analog Output Status', kind: 'analog'  },
] as const

export function pointTypeKind(pt: string | undefined): 'digital' | 'analog' {
  return POINT_TYPES.find(p => p.value === pt)?.kind ?? 'analog'
}

export function pointTypeLabel(pt: string | undefined): string {
  return POINT_TYPES.find(p => p.value === pt)?.label ?? (pt || '—')
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
