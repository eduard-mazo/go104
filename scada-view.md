SCADA Expansion — Detailed Plan
                                                                                                         1. New Element Library
                                                                                                         Electrical:
                                                                                                         ┌─────────────────┬──────────────────────────────┬────────────────────┐                                │     Symbol      │         SVG approach         │    Signal type     │
  ├─────────────────┼──────────────────────────────┼────────────────────┤
  │ Circuit breaker │ Two squares + contact lines  │ Digital (CLOSED=1) │
  ├─────────────────┼──────────────────────────────┼────────────────────┤
  │ Motor/pump      │ Circle + M glyph             │ Digital (RUN=1)    │
  ├─────────────────┼──────────────────────────────┼────────────────────┤                                │ Transformer     │ Two coils (sine arcs)        │ Analog (kVA load)  │
  ├─────────────────┼──────────────────────────────┼────────────────────┤                                │ Bus bar         │ Thick horizontal line + taps │ Analog (V/A)       │
  ├─────────────────┼──────────────────────────────┼────────────────────┤                                │ Indicator lamp  │ Circle + glow filter         │ Digital            │
  └─────────────────┴──────────────────────────────┴────────────────────┘                              
  Water/Fluid:

  ┌───────────────────────────────┬────────────────────────┬─────────────┐
  │            Symbol             │      SVG approach      │ Signal type │
  ├───────────────────────────────┼────────────────────────┼─────────────┤
  │ Gate valve (existing, rename) │ Bowtie triangles       │ Digital     │
  ├───────────────────────────────┼────────────────────────┼─────────────┤
  │ Ball valve                    │ Circle + bar           │ Digital     │
  ├───────────────────────────────┼────────────────────────┼─────────────┤                               │ Pump (centrifugal)            │ Circle + impeller arcs │ Digital     │
  ├───────────────────────────────┼────────────────────────┼─────────────┤
  │ Tank                          │ Rect + fill level      │ Analog      │
  ├───────────────────────────────┼────────────────────────┼─────────────┤                               │ Flow meter                    │ Diamond + arrows       │ Analog      │
  ├───────────────────────────────┼────────────────────────┼─────────────┤                               │ Pressure gauge                │ Circle + needle        │ Analog      │
  └───────────────────────────────┴────────────────────────┴─────────────┘

  Gas:

  ┌────────────────┬───────────────────────────┬─────────────┐
  │     Symbol     │       SVG approach        │ Signal type │                                           ├────────────────┼───────────────────────────┼─────────────┤
  │ Compressor     │ Circle + wave glyph       │ Digital     │
  ├────────────────┼───────────────────────────┼─────────────┤
  │ Separator      │ Tall rect + internal line │ Analog      │
  ├────────────────┼───────────────────────────┼─────────────┤                                           │ Relief valve   │ Triangle + spring         │ Digital     │
  ├────────────────┼───────────────────────────┼─────────────┤
  │ Heat exchanger │ Rect + crossed arrows     │ Analog      │
  └────────────────┴───────────────────────────┴─────────────┘                                         
  All: same ScadaElement shape — add kind string, extend ElementKind union type.                       
  ---
  2. Connection Lines / Pipes

  Data model:
  interface ScadaLine {
    id: string
    from_el: string | null   // element id or null (floating endpoint)
    to_el:   string | null
    from_pt: { x: number, y: number }  // abs canvas coords                                                to_pt:   { x: number, y: number }
    style: 'pipe_water' | 'pipe_gas' | 'wire' | 'cable'                                                    color: string
    stroke_width: number                                                                                   dashed: boolean
    label: string                                                                                        }

  Storage: Add lines: ScadaLine[] alongside elements in ScadaView. Stored as second JSON blob column or   merged into single layout column: { elements: [...], lines: [...] }.                                
  Rendering: SVG overlay layer behind elements. Each line renders as <path> with orthogonal routing      (L-shape, two segments) or diagonal direct line. Animated dashes on active flow (CSS
  stroke-dashoffset animation).

  Editor UX:                                                                                             - Palette: "Connect" tool toggle button
  - Click element → blue port dot appears at edges
  - Click first port → drag → click second port = creates line
  - Click on line to select → props panel shows style/color/label
  - Delete line: right-click → delete, or select + Delete key

  Port attachment: Each element kind declares port positions as relative fractions (e.g. valve:
  left={0,0.5}, right={1,0.5}). When element moves, lines recompute endpoints from current el.x +        port.rx * el.w.

  ---
  3. Signal History Graph Modal

  Backend — new endpoint:
  GET /api/signals/{id}/history?from=<unix>&to=<unix>&limit=1000
  Response: [{ ts: number, value: number, quality: number }]   Store: recent history lives in an
  IN-MEMORY per-signal ring buffer (internal/store/history_mem.go), NOT SQLite — per-sample history
  writes were the single-writer bottleneck. Each datapoint appends O(1) to the ring; QueryHistory
  serves the recent window. Bounded by GO104_HISTORY_POINTS (default 2000/signal, 0=off);
  non-persistent (lost on restart). Long-term archival is the upstream historian's job (Sparkplug/Timescale).

  Frontend — HistoryModal.vue:

  Layout:
  ┌─────────────────────────────────────────────────┐
  │ Signal History          [+ Compare]  [×]        │
  │                                                 │                                                    │ Primary: FCV-101 ───────── [Line 1][Last 1h ▾] │
  │ Compare: ─ none ─  ───────── [Line 2]           │
  │                                                 │
  │  ┌───────────────────────────────────────────┐  │                                                    │  │  SVG/Canvas chart — dual Y axes           │  │
  │  │  Tooltip crosshair on hover               │  │
  │  └───────────────────────────────────────────┘  │
  │                                                 │                                                    │ [1h] [6h] [24h] [7d]   from: [datetime] to: [datetime] │
  └─────────────────────────────────────────────────┘

  Chart library choice: Use uPlot (2.8kB gzipped, canvas-based, fastest for time series) — no heavy
  deps. Single canvas, dual Y-axes when compare signal present.                                        
  State:
  const modal = ref<{
    open: boolean
    primarySignalId: number
    compareSignalId: number | null
    range: '1h'|'6h'|'24h'|'7d'|'custom'
    from: number
    to: number
  }>

  Right-click trigger (both design + live mode):
  - Add menu item: { icon: '📈', label: 'View history…', action: () => openHistoryModal(el.signal_id) }  - Modal Teleports to <body>, backdrop blur
                                                                                                         ---                                                                                                    4. File / Component Changes
                                                                                                         Backend (Go):
  internal/store/store.go          UpsertDatapoint → in-memory RTDB (dpCache) + dirty mark, no hot-path write
  internal/store/snapshot.go       batched periodic persist of current values (GO104_SNAPSHOT_SEC, default 30s)
  internal/store/history_mem.go    in-memory history ring (InsertHistory/QueryHistory), QueryHistory handler
  internal/api/scada.go            + getSignalHistory handler
  internal/api/router.go           + GET /api/signals/{id}/history                                       internal/iec104/line_worker.go   + insert row on every decoded datapoint
  internal/models/models.go        + HistoryPoint struct
                                                                                                         Frontend (Vue/TS):                                                                                     web/src/api/client.ts            + signalHistoryAPI, ScadaLine interface, updated ScadaView
  web/src/stores/scada.ts          + addLine, removeLine, updateLine, lines in save payload
  web/src/components/scada/                                                                                elements/
      CircuitBreaker.vue           new
      MotorPump.vue                new
      Transformer.vue              new
      BallValve.vue                new
      Tank.vue                     new
      FlowMeter.vue                new                                                                       Compressor.vue               new
      PressureGauge.vue            new
      HeatExchanger.vue            new
      IndicatorLamp.vue            new                                                                     ConnectionLayer.vue            new  (SVG overlay, renders all ScadaLine[])
    ConnectionHandle.vue           new  (port dot during connect mode)
    HistoryModal.vue               new  (uPlot chart, compare picker)
  web/src/views/ScadaDesign.vue    + connect-mode toggle, palette groups, line selection                 web/src/views/ScadaLive.vue      + history modal trigger on right-click

  Palette sidebar reorganizes into collapsible groups:
  - ▾ Fluid — valve, ball valve, pump, tank, flow meter                                                  - ▾ Gas — compressor, separator, relief valve, heat exchanger
  - ▾ Electrical — breaker, motor, transformer, bus bar, lamp
  - ▾ Connect — wire, pipe_water, pipe_gas tool buttons                                                
  ---                                                                                                    5. SVG Design Approach (all new elements)
                                                                                                         Same pattern as existing ValveSVG.vue:
  - Pure <svg> with viewBox="0 0 100 100" (or element-appropriate aspect)                                - Props: config, signal (Datapoint|null), selected, w, h                                               - Computed fill from signal value + quality
  - transition: fill 0.35s ease on animated parts
  - Moving parts (pump impeller, needle) use transform: rotate() driven by signal value
  - IEC/ISA-5.1 standard shapes where possible
                                                                                                         ---
  6. Implementation Order                                                                              
  1. Backend history store — table + insert + query endpoint (no UI dep)
  2. Update ScadaView data model — add lines array, migrate existing records
  3. HistoryModal.vue — standalone, test with real data before wiring context menu
  4. Right-click → history in both Live + Design views
  5. New element SVGs — batch implement (10 components, ~3 similar groups)
  6. ElementKind expansion — add to client.ts, element registry, palette
  7. ConnectionLayer.vue + port system — most complex, do last
                                                                                                         ---
  7. Key Decisions

  ┌────────────────┬───────────────────────────────────┬──────────────────────────────────────────┐
  │    Question    │             Decision              │                  Reason                  │
  ├────────────────┼───────────────────────────────────┼──────────────────────────────────────────┤
  │ Chart lib      │ uPlot                             │ Fastest, smallest, dual Y-axis, no D3    │
  │                │                                   │ overhead                                 │
  ├────────────────┼───────────────────────────────────┼──────────────────────────────────────────┤
  │ Line routing   │ Orthogonal (2-segment L)          │ Cleaner for P&ID; diagonal as fallback   │      ├────────────────┼───────────────────────────────────┼──────────────────────────────────────────┤
  │ History        │ In-memory ring (recent window),   │ Per-sample SQLite history was the        │      │ retention      │ default 2000 pts/signal           │ single-writer bottleneck; archival is    │      │                │ (GO104_HISTORY_POINTS)            │ the upstream historian's job             │
  ├────────────────┼───────────────────────────────────┼──────────────────────────────────────────┤      │ Port model     │ Declarative per-kind              │ Recompute on move = no stale coords      │
  ├────────────────┼───────────────────────────────────┼──────────────────────────────────────────┤      │ Layout column  │ Merge {elements,lines} into       │ Atomic save, backward compatible         │
  │                │ single JSON                       │ (default lines=[])                       │      └────────────────┴───────────────────────────────────┴──────────────────────────────────────────┘

  ---
  Ready to implement? Start with backend history table + endpoint, or dive into SVG elements first?
