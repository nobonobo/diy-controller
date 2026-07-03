package controller

import (
	"sync"
	"time"

	"github.com/nobonobo/q16"

	"github.com/nobonobo/diy-controller/effects"
	"github.com/nobonobo/diy-controller/motor"
	"github.com/nobonobo/diy-controller/settings"
)

const (
	MaxSystemEffects   = 4
	DeltaTime          = 1.0 / 1000.0
	VelCutOffFrequency = 10.0 // [Hz]
)

var (
	DeltaTimeQ16      = q16.FromFloat64(DeltaTime)
	VelCutOffLPFAlpha = CalcAlpha(VelCutOffFrequency, DeltaTime)
)

type Input = motor.State

type Output struct {
	Angle    q16.Fixed // クランプ後の角度
	Velocity q16.Fixed // 角速度
	Power    q16.Fixed // 出力比率 [-1.0 - 1.0]
}

// Controller FFB仮想物理モデルコントローラー
// settings.Settingsをコンポジションで組み込み、dtベースの物理シミュレーションで出力を算出する。
type Controller struct {
	mu            sync.RWMutex
	settings      settings.Settings // 設定パラメータ（プライベートフィールド）
	prevTime      time.Time         // 前回のUpdate呼び出し時刻
	prevVel       q16.Fixed         // 前回の仮想角速度 [rad/s]
	userEffects   *effects.EffectPool
	systemEffects *effects.EffectPool
	velocityLPF   *LPF
}

// New settings.Settings付きでControllerを新規作成する
func New(effectPool *effects.EffectPool) *Controller {
	c := &Controller{
		prevTime: time.Now(),
		prevVel:  q16.FromInt(0),
		// userEffects は MaxUserEffects サイズのヒープ上にスライスを確保し、Effectポインタ群を初期化する。
		userEffects: effectPool,
		// systemEffects は MaxSystemEffects サイズのヒープ上にスライスを確保し、Effectポインタ群を初期化する。
		systemEffects: effects.NewEffectPool(MaxSystemEffects),
		velocityLPF:   NewLPF(VelCutOffLPFAlpha),
	}
	return c
}

func (c *Controller) Gains() settings.Gains {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.userEffects.Gains()
}

func (c *Controller) SetGains(gains settings.Gains) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.userEffects.SetGains(gains)
}

// Settings settingsへのアクセサ（設定読み取り用）
func (c *Controller) Settings() settings.Settings {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.settings
}

// SetSettings settingsへのアクセサ（設定変更用）
func (c *Controller) SetSettings(s settings.Settings) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.settings = s
	c.systemEffects.Clear()
	idx := c.systemEffects.Allocate()
	param := c.systemEffects.Get(idx).EffectParam()
	param.EffectType = effects.EffInertia
	effect := c.systemEffects.Get(idx)
	effect.SetEffectParam(param)
	for i := uint8(0); i < effects.MaxAxisCount; i++ {
		effect.SetCondition(i, effects.Condition{
			PositiveCoefficient: s.KInertia,
			NegativeCoefficient: s.KInertia,
			PositiveSaturation:  q16.FromInt(1),
			NegativeSaturation:  q16.FromInt(1),
			DeadBand:            s.KInertiaDeadBand,
		})
	}
	effect.Start()
	idx = c.systemEffects.Allocate()
	param = c.systemEffects.Get(idx).EffectParam()
	param.EffectType = effects.EffDamper
	effect = c.systemEffects.Get(idx)
	effect.SetEffectParam(param)
	for i := uint8(0); i < effects.MaxAxisCount; i++ {
		effect.SetCondition(i, effects.Condition{
			PositiveCoefficient: s.KDamper,
			NegativeCoefficient: s.KDamper,
			PositiveSaturation:  q16.FromInt(1),
			NegativeSaturation:  q16.FromInt(1),
			DeadBand:            s.KDamperDeadBand,
		})
	}
	effect.Start()
	idx = c.systemEffects.Allocate()
	param = c.systemEffects.Get(idx).EffectParam()
	param.EffectType = effects.EffSpring
	effect = c.systemEffects.Get(idx)
	effect.SetEffectParam(param)
	for i := uint8(0); i < effects.MaxAxisCount; i++ {
		effect.SetCondition(i, effects.Condition{
			PositiveCoefficient: s.KSpring,
			NegativeCoefficient: s.KSpring,
			PositiveSaturation:  s.KSpringLimit,
			NegativeSaturation:  s.KSpringLimit,
			DeadBand:            s.KSpringDeadBand,
		})
	}
	effect.Start()
	idx = c.systemEffects.Allocate()
	param = c.systemEffects.Get(idx).EffectParam()
	param.EffectType = effects.EffFriction
	effect = c.systemEffects.Get(idx)
	effect.SetEffectParam(param)
	for i := uint8(0); i < effects.MaxAxisCount; i++ {
		effect.SetCondition(i, effects.Condition{
			PositiveCoefficient: s.KFriction,
			NegativeCoefficient: s.KFriction,
			PositiveSaturation:  q16.FromInt(1),
			NegativeSaturation:  q16.FromInt(1),
			DeadBand:            s.KFrictionDeadBand,
		})
	}
	effect.Start()
}

// Update 前回の呼び出しからの時差dtに基づき、仮想パラメータに従って出力値を算出する。
// stateは現在のモーターセンサフィードバック（角度・角速度）。
func (c *Controller) Update(state *Input, axis int) *Output {
	now := time.Now()
	dt := q16.FromDuration(now.Sub(c.prevTime))
	c.prevTime = now

	angle := state.Angle - c.settings.Neutral
	if q16.Abs(angle) < c.settings.Backlash {
		angle = 0
	} else if angle > q16.Zero {
		angle -= c.settings.Backlash // バックラッシュ範囲外（正）: angle - backlash
	} else {
		angle += c.settings.Backlash // バックラッシュ範囲外（負）: angle + backlash
	}

	// dtが異常に大きい場合は、そのまま現在の状態を返す
	if dt > q16.FromDuration(200*time.Millisecond) {
		c.velocityLPF.Reset(q16.Zero)
		return &Output{
			Angle:    angle,
			Velocity: state.Velocity,
			Power:    q16.Zero,
		}
	}
	if dt < DeltaTimeQ16 {
		dt = DeltaTimeQ16 // 最速1KHz相当
	}
	velocity := c.velocityLPF.Update(state.Velocity)
	// 加速度の計算
	accel := q16.Zero
	if dt > q16.Zero {
		accel = q16.Div(velocity-c.prevVel, dt)
	}
	c.prevVel = velocity

	// ロックトゥロック反力の生成
	lockPower := q16.Zero
	if angle > c.settings.HalfOfL2L {
		lockPower = q16.Mul(-c.settings.KLock, (angle - c.settings.HalfOfL2L))
		if state.Velocity > 0 {
			lockPower -= q16.Mul(c.settings.KBrake, state.Velocity)
		}
	} else if angle < -c.settings.HalfOfL2L {
		lockPower = q16.Mul(-c.settings.KLock, (angle + c.settings.HalfOfL2L))
		if state.Velocity < 0 {
			lockPower -= q16.Mul(c.settings.KBrake, state.Velocity)
		}
	}
	// 速度が MaxSpeed を超えている場合は速度に比例してブレーキトルク生成
	brakePower := q16.Zero
	excess := q16.Abs(velocity) - c.settings.MaxSpeed
	if excess > q16.Zero {
		// 逆トルク = -sign(Velocity) * (KBrakeGain * excess)
		brakePower = q16.Mul(-q16.Sign(velocity), q16.Mul(c.settings.KBrake, excess))
	}

	params := &effects.Params{
		Delta:    dt,
		Angle:    angle,
		Velocity: state.Velocity,
		Accel:    accel,
	}
	// トルク合計
	sysOutput := c.systemEffects.Calc(params, axis)
	userOutput := c.userEffects.Calc(params, axis)
	output := lockPower + brakePower + sysOutput + userOutput
	// 出力を [-MinTorque, MinTorque] 分ずらす
	if output > q16.Zero {
		output += c.settings.MinOut
	} else if output < q16.Zero {
		output -= c.settings.MinOut
	}
	// 出力を [-MaxTorque, MaxTorque] にクランプ
	if output >= c.settings.MaxOut {
		output = c.settings.MaxOut
	} else if output <= -c.settings.MaxOut {
		output = -c.settings.MaxOut
	}
	return &Output{
		Angle:    angle,
		Velocity: state.Velocity,
		Power:    output,
	}
}
