import type { App } from 'vue'
import { scadaRegistry } from './registry'
import { builtinElements } from './elements'

export { scadaRegistry } from './registry'
export type { ElementRegistry } from './registry'
export * from './types'
export { builtinElements } from './elements'

export interface ScadaPluginOptions {
  /** Skip auto-registration of built-ins. Use to swap in a custom pack. */
  includeBuiltins?: boolean
}

export const ScadaPlugin = {
  install(app: App, opts: ScadaPluginOptions = {}) {
    const includeBuiltins = opts.includeBuiltins ?? true
    if (includeBuiltins) scadaRegistry.register(builtinElements)
    app.config.globalProperties.$scada = scadaRegistry
  },
}
