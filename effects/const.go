package effects

const MaxAxisCount = 2

// EffectType エフェクトの種類
type EffectType uint8

const (
	EffNone         EffectType = iota // 0:は予約
	EffConstant                       // 1:一定力
	EffRamp                           // 2:ランプ（勾配）
	EffSquare                         // 3:矩形波
	EffSine                           // 4:周期性（サイン波）
	EffTriangle                       // 5:三角波
	EffSawtoothDown                   // 6:ノコギリ波
	EffSawtoothUp                     // 7:ノコギリ波
	EffSpring                         // 8:バネ
	EffDamper                         // 9:ダンパー
	EffInertia                        // 10:慣性
	EffFriction                       // 11:摩擦
	EffCustom                         // 12:カスタム
)

type EffectState uint8

const (
	EffFree      EffectState = iota // フリー
	EffAllocated                    // アロケート
	EffPlaying                      // 生成中
	EffStopping                     // 停止中
)
