package settings

import (
	"github.com/nobonobo/q16"
)

// ゲイン設定
type Gains struct {
	TotalGain        q16.Fixed // 総合ゲイン [0.0 - 1.0]
	ConstantGain     q16.Fixed // 定常力エフェクトのゲイン [0.0 - 1.0]
	RampGain         q16.Fixed // ランプ力エフェクトのゲイン [0.0 - 1.0]
	SquareGain       q16.Fixed // 正方形波エフェクトのゲイン [0.0 - 1.0]
	SineGain         q16.Fixed // 正弦波エフェクトのゲイン [0.0 - 1.0]
	TriangleGain     q16.Fixed // 三角波エフェクトのゲイン [0.0 - 1.0]
	SawtoothDownGain q16.Fixed // ノコギリ波（下がり）エフェクトのゲイン [0.0 - 1.0]
	SawtoothUpGain   q16.Fixed // ノコギリ波（上がり）エフェクトのゲイン [0.0 - 1.0]
	SpringGain       q16.Fixed // ばねエフェクトのゲイン [0.0 - 1.0]
	DamperGain       q16.Fixed // ダンパーエフェクトのゲイン [0.0 - 1.0]
	InertiaGain      q16.Fixed // 慣性エフェクトのゲイン [0.0 - 1.0]
	FrictionGain     q16.Fixed // 摩擦エフェクトのゲイン [0.0 - 1.0]
	CustomGain       q16.Fixed // カスタム力エフェクトのゲイン [0.0 - 1.0]
}

func NewGains() *Gains {
	return &Gains{
		TotalGain:        q16.FromInt(1),
		ConstantGain:     q16.FromInt(1),
		RampGain:         q16.FromInt(1),
		SquareGain:       q16.FromInt(1),
		SineGain:         q16.FromInt(1),
		TriangleGain:     q16.FromInt(1),
		SawtoothDownGain: q16.FromInt(1),
		SawtoothUpGain:   q16.FromInt(1),
		SpringGain:       q16.FromInt(1),
		DamperGain:       q16.FromInt(1),
		InertiaGain:      q16.FromInt(1),
		FrictionGain:     q16.FromInt(1),
		CustomGain:       q16.FromInt(1),
	}
}

// Settings コントローラーの設定パラメータ（すべてQ16.16固定小数点）
type Settings struct {
	// ハードウェア特性
	Neutral   q16.Fixed // ニュートラルの位置 [rad]
	HalfOfL2L q16.Fixed // ロックトゥロック角度の半分 [rad]
	KLock     q16.Fixed // ロック時のトルク係数 [N·m/rad/MaxTorque]
	// システム物理特性パラメータ
	KSpring      q16.Fixed // 仮想バネ定数 [N·m/rad/MaxTorque]
	KSpringLimit q16.Fixed // バネ最大トルク比率 [0.0, 1.0]
	KDamper      q16.Fixed // 仮想粘性係数 [N·m·s/rad/MaxTorque]
	KInertia     q16.Fixed // 仮想イナーシャ [N·m·s²/rad/MaxTorque]
	KFriction    q16.Fixed // 仮想摩擦係数 [N·m·s/rad/MaxTorque]
	Backlash     q16.Fixed // 仮想バックラッシュ [rad]
	// 出力制限
	MinOut   q16.Fixed // 最低出力レート [0.0 - 1.0] [N·m/MaxTorque]
	MaxOut   q16.Fixed // 最高出力レート [0.0 - 1.0] [N·m/MaxTorque]
	MaxSpeed q16.Fixed // 最高速度 [rad/s]
	KBrake   q16.Fixed // 減速係数 [N·m·s/rad]
}
