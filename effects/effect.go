package effects

import (
	"github.com/nobonobo/q16"

	"github.com/nobonobo/diy-controller/settings"
)

type Params struct {
	Delta    q16.Fixed // Delta Time [sec]
	Angle    q16.Fixed // Angle [rad]
	Velocity q16.Fixed // Velocity [rad/sec]
	Accel    q16.Fixed // Acceleration [rad/sec^2]
}

type EffectParam struct {
	EffectType            EffectType // エフェクトタイプ（定数参照）
	Duration              q16.Fixed  // エフェクトの継続時間（秒数、0は無限）
	TriggerRepeatInterval q16.Fixed  // トリガー再発生インターバル（秒数、0は無限）
	SamplePeriod          q16.Fixed  // サンプリング期間（秒数、0は無限）
	Gain                  q16.Fixed  // ゲイン値（0-1.0、物理的には0-10000にマッピング可能） // 単位: ratio (Torque = Gain * MaxTorque [Nm])
	TriggerButton         uint8      // トリガーボタンID（0-8、0は常時有効）
	EnableAxis            uint8      // 軸有効化ビットマスク: Bit0=X, Bit1=Y, Bit2=DirectionEnable
	DirectionX            uint8      // X方向角度（0=0度 .. 255=360度）
	DirectionY            uint8      // Y方向角度（0=0度 .. 255=360度）
	StartDelay            q16.Fixed  // 開始遅延時間（0-1.0s） // 単位: sec
}

type PeriodicParam struct {
	Magnitude q16.Fixed // 力覚の振幅（0-1.0） // 単位: fraction of MaxTorque [Nm]
	Offset    q16.Fixed // オフセット（中心位置） // 単位: fraction of MaxTorque [Nm]
	Phase     q16.Fixed // 位相（0-2π rad） // 単位: rad
	Period    q16.Fixed // 周期（秒） // 単位: sec
}

type Envelope struct {
	AttackLevel q16.Fixed // Attackレベル（0-1.0力覚の開始強度） // 単位: fraction of MaxTorque [Nm]
	FadeLevel   q16.Fixed // Fadeレベル（-1.0-1.0力覚の終了強度、符号あり） // 単位: fraction of MaxTorque [Nm]
	AttackTime  q16.Fixed // Attack時間（秒） // 単位: sec
	FadeTime    q16.Fixed // Fade時間（秒） // 単位: sec
}

type RampParam struct {
	StartMagnitude q16.Fixed // 0.0-1.0力覚の開始強度
	EndMagnitude   q16.Fixed // 0.0-1.0力覚の終了強度
}

type Condition struct {
	CpOffset            q16.Fixed // Cpオフセット（-1.0-1.0）
	PositiveCoefficient q16.Fixed // 正方向係数（0-1.0） // 単位: ratio (Torque = coeff * MaxTorque [Nm])
	NegativeCoefficient q16.Fixed // 負方向係数（0-1.0） // 単位: ratio (Torque = coeff * MaxTorque [Nm])
	PositiveSaturation  q16.Fixed // 正方向飽和（0-1.0） // 単位: ratio (max Torque = sat * MaxTorque [Nm])
	NegativeSaturation  q16.Fixed // 負方向飽和（-1.0-0.0） // 単位: ratio (min Torque = sat * MaxTorque [Nm])
	DeadBand            q16.Fixed // デッドバンド（0-2πラジアン） // 単位: rad
}

type Effect struct {
	param         EffectParam
	periodicParam PeriodicParam
	rampParam     RampParam
	envelope      Envelope
	conditions    [MaxAxisCount]Condition

	state         EffectState
	elapsedTime   q16.Fixed
	totalDuration q16.Fixed
	gains         *settings.Gains
}

func (e *Effect) TotalDuration() q16.Fixed {
	if e.totalDuration == q16.Zero {
		return e.param.Duration
	}
	return e.totalDuration
}

func (e *Effect) SetTotalDuration(totalDuration q16.Fixed) {
	e.totalDuration = totalDuration
}

func (e *Effect) Start() {
	if e.state == EffAllocated {
		e.elapsedTime = q16.Zero
		e.state = EffPlaying
	}
}

func (e *Effect) Stop() {
	if e.state == EffPlaying {
		e.state = EffStopping
	}
}

func (e *Effect) Free() {
	e.state = EffFree
}

func (e *Effect) Clear() {
	e.elapsedTime = q16.Zero
	e.totalDuration = q16.Zero // 無効
	e.param.EffectType = EffNone
	e.param.Gain = q16.FromInt(1)               // 1.0
	e.param.Duration = q16.Zero                 // 無効
	e.param.EnableAxis = 1                      // b0:X, b1:Y
	e.param.DirectionX = 0                      // 0..255:0..360deg
	e.param.DirectionY = 0                      // 0..255:0..360deg
	e.param.StartDelay = q16.Zero               // 0 ms
	e.param.TriggerButton = 0                   // 常に有効
	e.param.SamplePeriod = q16.Zero             // 0 ms
	e.param.TriggerRepeatInterval = q16.Zero    // 0 ms
	e.periodicParam.Magnitude = q16.FromInt(1)  // 1.0
	e.periodicParam.Offset = q16.Zero           // 0.0
	e.periodicParam.Phase = q16.Zero            // 0.0
	e.periodicParam.Period = q16.Zero           // 0.0
	e.rampParam.StartMagnitude = q16.FromInt(1) // 1.0
	e.rampParam.EndMagnitude = q16.FromInt(1)   // 1.0
	e.envelope.AttackLevel = q16.FromInt(1)     // 1.0
	e.envelope.FadeLevel = q16.Zero             // 0.0
	e.envelope.AttackTime = q16.Zero            // 0 ms
	e.envelope.FadeTime = q16.Zero              // 0 ms
	for i := range e.conditions {
		e.conditions[i] = Condition{
			CpOffset:            q16.Zero,
			PositiveCoefficient: q16.Zero,
			NegativeCoefficient: q16.Zero,
			PositiveSaturation:  q16.Zero,
			NegativeSaturation:  q16.Zero,
			DeadBand:            q16.Zero,
		}
	}
}

func (e *Effect) IsPlaying() bool {
	return e.state == EffPlaying || e.state == EffStopping
}

func (e *Effect) State() EffectState {
	return e.state
}

func (e *Effect) Duration() q16.Fixed {
	return e.totalDuration
}

func (e *Effect) SetDuration(totalDuration q16.Fixed) {
	e.totalDuration = totalDuration
}

func (e *Effect) SetEffectParam(param EffectParam) {
	e.param = param
	e.state = EffAllocated
}

func (e *Effect) EffectParam() EffectParam {
	return e.param
}

func (e *Effect) ElapsedTime() q16.Fixed {
	return e.elapsedTime
}

func (e *Effect) AddElapsedTime(delta q16.Fixed) {
	e.elapsedTime += delta
}

func (e *Effect) SetPeriodicParam(param PeriodicParam) {
	e.periodicParam = param
}

func (e *Effect) PeriodicParam() PeriodicParam {
	return e.periodicParam
}

func (e *Effect) SetRampParam(param RampParam) {
	e.rampParam = param
}

func (e *Effect) RampParam() RampParam {
	return e.rampParam
}

func (e *Effect) SetEnvelope(envelope Envelope) {
	e.envelope = envelope
}

func (e *Effect) Envelope() Envelope {
	return e.envelope
}

func (e *Effect) SetCondition(axis uint8, condition Condition) {
	e.conditions[axis] = condition
}

func (e *Effect) Condition(axis uint8) Condition {
	return e.conditions[axis]
}

func (e *Effect) ApplyEnvelope(torque q16.Fixed) q16.Fixed {
	if e.elapsedTime < e.envelope.AttackTime {
		return q16.Mul(torque, e.envelope.AttackLevel)
	}
	if e.totalDuration > 0 {
		if e.elapsedTime > e.totalDuration-e.envelope.FadeTime {
			return q16.Mul(torque, e.envelope.FadeLevel)
		}
	}
	return torque
}

func (e *Effect) Calc(params *Params, axis int) (torque q16.Fixed) {
	if !e.IsPlaying() {
		return q16.Zero
	}
	e.elapsedTime += params.Delta
	if e.state == EffStopping {
		if e.elapsedTime >= e.totalDuration+e.envelope.FadeTime {
			e.state = EffAllocated
			return q16.Zero
		}
	}
	switch e.param.EffectType {
	case EffConstant: // Periodic
		// なぜかConstantGainだけ符号反転を期待されている
		return q16.Mul(-e.gains.ConstantGain, q16.Mul(e.param.Gain, e.ApplyEnvelope(e.periodicParam.Magnitude)))
	case EffRamp:
		torque = e.rampParam.StartMagnitude + q16.Mul(q16.Div(e.rampParam.EndMagnitude-e.rampParam.StartMagnitude, e.param.Duration), e.elapsedTime)
		return q16.Mul(e.gains.RampGain, q16.Mul(e.param.Gain, torque))
	case EffSquare: // Periodic
		wave := q16.Sin(q16.Div(q16.Mul(q16.Period, e.elapsedTime), e.periodicParam.Period))
		torque = q16.Mul(e.periodicParam.Magnitude, q16.Sign(wave))
		return q16.Mul(e.gains.SquareGain, q16.Mul(e.param.Gain, e.ApplyEnvelope(torque)))
	case EffSine: // Periodic
		wave := q16.Sin(q16.Div(q16.Mul(q16.Period, e.elapsedTime), e.periodicParam.Period))
		torque = q16.Mul(e.periodicParam.Magnitude, wave)
		return q16.Mul(e.gains.SineGain, q16.Mul(e.param.Gain, e.ApplyEnvelope(torque)))
	case EffTriangle: // Periodic
		_, m := q16.DivMod(q16.Mul(q16.Period, e.elapsedTime), e.Duration())
		torque = q16.Zero
		switch {
		case m < q16.Pi:
			torque = q16.Div(m, q16.Pi)
		case m < q16.Period:
			torque = q16.FromInt(2) - q16.Div(m, q16.Pi)
		case m < q16.Period+q16.Pi:
			torque = -(q16.Div(m, q16.Pi) - q16.FromInt(2))
		default:
			torque = q16.Div(m, q16.Pi) - q16.FromInt(4)
		}
		return q16.Mul(e.gains.TriangleGain, q16.Mul(e.param.Gain, e.ApplyEnvelope(torque)))
	case EffSawtoothDown: // Periodic
		_, m := q16.DivMod(q16.Mul(q16.Period, e.elapsedTime), e.periodicParam.Period)
		torque = q16.Mul(e.periodicParam.Magnitude, q16.FromInt(1)-q16.Div(m, q16.Period))
		return q16.Mul(e.gains.SawtoothDownGain, q16.Mul(e.param.Gain, e.ApplyEnvelope(torque)))
	case EffSawtoothUp: // Periodic
		_, m := q16.DivMod(q16.Mul(q16.Period, e.elapsedTime), e.periodicParam.Period)
		torque = q16.Mul(e.periodicParam.Magnitude, q16.Div(m, q16.Period))
		return q16.Mul(e.gains.SawtoothUpGain, q16.Mul(e.param.Gain, e.ApplyEnvelope(torque)))
	case EffSpring:
		cond := e.conditions[axis]
		angle := params.Angle - cond.CpOffset
		if q16.Abs(angle) < cond.DeadBand {
			torque = q16.Zero
		} else if angle > q16.Zero {
			torque = q16.Mul(-cond.PositiveCoefficient, angle)
		} else {
			torque = q16.Mul(-cond.NegativeCoefficient, angle)
		}
		if torque > cond.PositiveSaturation {
			torque = cond.PositiveSaturation
		} else if torque < -cond.NegativeSaturation {
			torque = -cond.NegativeSaturation
		}
		return q16.Mul(e.gains.SpringGain, torque)
	case EffDamper:
		cond := e.conditions[axis]
		vel := params.Velocity - cond.CpOffset
		if q16.Abs(vel) < cond.DeadBand {
			torque = q16.Zero
		} else if vel > q16.Zero {
			torque = q16.Mul(-cond.PositiveCoefficient, vel)
		} else {
			torque = q16.Mul(-cond.NegativeCoefficient, vel)
		}
		if torque > cond.PositiveSaturation {
			torque = cond.PositiveSaturation
		} else if torque < -cond.NegativeSaturation {
			torque = -cond.NegativeSaturation
		}
		return q16.Mul(e.gains.DamperGain, torque)
	case EffInertia:
		cond := e.conditions[axis]
		acc := params.Accel - cond.CpOffset
		if q16.Abs(acc) < cond.DeadBand {
			torque = q16.Zero
		} else if acc > q16.Zero {
			torque = q16.Mul(-cond.PositiveCoefficient, acc)
		} else {
			torque = q16.Mul(-cond.NegativeCoefficient, acc)
		}
		if torque > cond.PositiveSaturation {
			torque = cond.PositiveSaturation
		} else if torque < -cond.NegativeSaturation {
			torque = -cond.NegativeSaturation
		}
		return q16.Mul(e.gains.InertiaGain, torque)
	case EffFriction:
		cond := e.conditions[axis]
		vel := params.Velocity - cond.CpOffset
		if q16.Abs(vel) < cond.DeadBand {
			torque = q16.Zero
		} else if vel > q16.Zero {
			torque = q16.Mul(-cond.PositiveCoefficient, vel)
		} else {
			torque = q16.Mul(-cond.NegativeCoefficient, vel)
		}
		if torque > cond.PositiveSaturation {
			torque = cond.PositiveSaturation
		} else if torque < -cond.NegativeSaturation {
			torque = -cond.NegativeSaturation
		}
		return q16.Mul(e.gains.FrictionGain, torque)
	case EffCustom:
		return q16.Zero
	}
	return q16.Zero
}
