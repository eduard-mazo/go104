// Barrel: every built-in descriptor in one array.
// Add a new built-in: drop a folder with index.ts, then add one line here.
import type { ElementDescriptor } from '../types'

import { GateValveDescriptor }     from './fluid/GateValve'
import { BallValveDescriptor }     from './fluid/BallValve'
import { ControlValveDescriptor }  from './fluid/ControlValve'
import { CheckValveDescriptor }    from './fluid/CheckValve'
import { PumpDescriptor }          from './fluid/Pump'
import { TankDescriptor }          from './fluid/Tank'
import { PipeSegmentDescriptor }   from './fluid/PipeSegment'
import { PctBarDescriptor }        from './display/PctBar'
import { CompressorDescriptor }    from './gas/Compressor'
import { CircuitBreakerDescriptor } from './electrical/CircuitBreaker'
import { MotorDescriptor }         from './electrical/Motor'
import { TransformerDescriptor }   from './electrical/Transformer'
import { IndicatorLampDescriptor } from './electrical/IndicatorLamp'
import { PressureGaugeDescriptor } from './instrument/PressureGauge'
import { FlowMeterDescriptor }     from './instrument/FlowMeter'
import { TempTxDescriptor }        from './instrument/TempTx'
import { HeatExchangerDescriptor } from './instrument/HeatExchanger'

export const builtinElements: ElementDescriptor[] = [
  // Fluid
  GateValveDescriptor,
  BallValveDescriptor,
  ControlValveDescriptor,
  CheckValveDescriptor,
  PumpDescriptor,
  TankDescriptor,
  PctBarDescriptor,
  PipeSegmentDescriptor,
  // Gas
  CompressorDescriptor,
  // Electrical
  CircuitBreakerDescriptor,
  MotorDescriptor,
  TransformerDescriptor,
  IndicatorLampDescriptor,
  // Instrument
  PressureGaugeDescriptor,
  FlowMeterDescriptor,
  TempTxDescriptor,
  HeatExchangerDescriptor,
]

// Re-export individual descriptors for selective registration.
export {
  GateValveDescriptor, BallValveDescriptor, ControlValveDescriptor, CheckValveDescriptor,
  PumpDescriptor, TankDescriptor, PipeSegmentDescriptor, PctBarDescriptor,
  CompressorDescriptor,
  CircuitBreakerDescriptor, MotorDescriptor, TransformerDescriptor, IndicatorLampDescriptor,
  PressureGaugeDescriptor, FlowMeterDescriptor, TempTxDescriptor, HeatExchangerDescriptor,
}
