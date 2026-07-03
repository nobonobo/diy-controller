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

func (g Gains) Merge(p map[string]int32) Gains {
	for k, v := range p {
		switch k {
		case "TotalGain":
			g.TotalGain = q16.Fixed(v)
		case "ConstantGain":
			g.ConstantGain = q16.Fixed(v)
		case "RampGain":
			g.RampGain = q16.Fixed(v)
		case "SquareGain":
			g.SquareGain = q16.Fixed(v)
		case "SineGain":
			g.SineGain = q16.Fixed(v)
		case "TriangleGain":
			g.TriangleGain = q16.Fixed(v)
		case "SawtoothDownGain":
			g.SawtoothDownGain = q16.Fixed(v)
		case "SawtoothUpGain":
			g.SawtoothUpGain = q16.Fixed(v)
		case "SpringGain":
			g.SpringGain = q16.Fixed(v)
		case "DamperGain":
			g.DamperGain = q16.Fixed(v)
		case "InertiaGain":
			g.InertiaGain = q16.Fixed(v)
		case "FrictionGain":
			g.FrictionGain = q16.Fixed(v)
		case "CustomGain":
			g.CustomGain = q16.Fixed(v)
		}
	}
	return g
}

func (g Gains) ToMap() map[string]int32 {
	return map[string]int32{
		"TotalGain":        int32(g.TotalGain),
		"ConstantGain":     int32(g.ConstantGain),
		"RampGain":         int32(g.RampGain),
		"SquareGain":       int32(g.SquareGain),
		"SineGain":         int32(g.SineGain),
		"TriangleGain":     int32(g.TriangleGain),
		"SawtoothDownGain": int32(g.SawtoothDownGain),
		"SawtoothUpGain":   int32(g.SawtoothUpGain),
		"SpringGain":       int32(g.SpringGain),
		"DamperGain":       int32(g.DamperGain),
		"InertiaGain":      int32(g.InertiaGain),
		"FrictionGain":     int32(g.FrictionGain),
		"CustomGain":       int32(g.CustomGain),
	}
}

func NewGains() Gains {
	return Gains{
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
	KSpring           q16.Fixed // 仮想バネ定数 [N·m/rad/MaxTorque]
	KSpringLimit      q16.Fixed // バネ最大トルク比率 [0.0, 1.0]
	KSpringDeadBand   q16.Fixed // バネのデッドバンド [rad]
	KDamper           q16.Fixed // 仮想粘性係数 [N·m·s/rad/MaxTorque]
	KDamperDeadBand   q16.Fixed // ダンパーのデッドバンド [rad]
	KInertia          q16.Fixed // 仮想イナーシャ [N·m·s²/rad/MaxTorque]
	KInertiaDeadBand  q16.Fixed //イナーシャのデッドバンド [rad]
	KFriction         q16.Fixed // 仮想摩擦係数 [N·m·s/rad/MaxTorque]
	KFrictionDeadBand q16.Fixed // 摩擦のデッドバンド [rad]
	Backlash          q16.Fixed // 仮想バックラッシュ [rad]
	// 出力制限
	MinOut   q16.Fixed // 最低出力レート [0.0 - 1.0] [N·m/MaxTorque]
	MaxOut   q16.Fixed // 最高出力レート [0.0 - 1.0] [N·m/MaxTorque]
	MaxSpeed q16.Fixed // 最高速度 [rad/s]
	KBrake   q16.Fixed // 減速係数 [N·m·s/rad]
}

func (s Settings) Merge(p map[string]int32) Settings {
	for k, v := range p {
		switch k {
		case "Neutral":
			s.Neutral = q16.Fixed(v)
		case "HalfOfL2L":
			s.HalfOfL2L = q16.Fixed(v)
		case "KLock":
			s.KLock = q16.Fixed(v)
		case "KSpring":
			s.KSpring = q16.Fixed(v)
		case "KSpringLimit":
			s.KSpringLimit = q16.Fixed(v)
		case "KSpringDeadBand":
			s.KSpringDeadBand = q16.Fixed(v)
		case "KDamper":
			s.KDamper = q16.Fixed(v)
		case "KDamperDeadBand":
			s.KDamperDeadBand = q16.Fixed(v)
		case "KInertia":
			s.KInertia = q16.Fixed(v)
		case "KInertiaDeadBand":
			s.KInertiaDeadBand = q16.Fixed(v)
		case "KFriction":
			s.KFriction = q16.Fixed(v)
		case "KFrictionDeadBand":
			s.KFrictionDeadBand = q16.Fixed(v)
		case "Backlash":
			s.Backlash = q16.Fixed(v)
		case "MinOut":
			s.MinOut = q16.Fixed(v)
		case "MaxOut":
			s.MaxOut = q16.Fixed(v)
		case "MaxSpeed":
			s.MaxSpeed = q16.Fixed(v)
		case "KBrake":
			s.KBrake = q16.Fixed(v)
		}
	}
	return s
}

func (s Settings) ToMap() map[string]int32 {
	return map[string]int32{
		"Neutral":           int32(s.Neutral),
		"HalfOfL2L":         int32(s.HalfOfL2L),
		"KLock":             int32(s.KLock),
		"KSpring":           int32(s.KSpring),
		"KSpringLimit":      int32(s.KSpringLimit),
		"KSpringDeadBand":   int32(s.KSpringDeadBand),
		"KDamper":           int32(s.KDamper),
		"KDamperDeadBand":   int32(s.KDamperDeadBand),
		"KInertia":          int32(s.KInertia),
		"KInertiaDeadBand":  int32(s.KInertiaDeadBand),
		"KFriction":         int32(s.KFriction),
		"KFrictionDeadBand": int32(s.KFrictionDeadBand),
		"Backlash":          int32(s.Backlash),
		"MinOut":            int32(s.MinOut),
		"MaxOut":            int32(s.MaxOut),
		"MaxSpeed":          int32(s.MaxSpeed),
		"KBrake":            int32(s.KBrake),
	}
}
