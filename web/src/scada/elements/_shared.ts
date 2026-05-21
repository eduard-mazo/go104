// Shared default config factories — keep descriptors DRY.
export const digitalDefault = (over: Record<string, unknown> = {}) => ({
  color_on:    '#22c55e',
  color_off:   '#ef4444',
  color_fault: '#f59e0b',
  label:       '',
  ...over,
})

export const gaugeDefault = (over: Record<string, unknown> = {}) => ({
  min:         0,
  max:         100,
  color_fill:  '#22d3ee',
  color_alarm: '#ef4444',
  alarm_high:  0,
  unit:        '',
  label:       '',
  ...over,
})
