package effects

import (
	"github.com/nobonobo/q16"

	"github.com/nobonobo/diy-controller/settings"
)

// NewEffects は指定された最大サイズのエフェクト配列をヒープ上に確保して作成します。
// この関数は、指定された最大数 (max) の効果 (Effect) を保持する構造体 (*EffectPool) を初期化し、
// ヒープ上にメモリを確保します。具体的には、*Effect型のスライスを動的に確保し、
// 各要素には個別にヒープ割り当てされた *Effect ポインタを設定します。
func NewEffectPool(max int) *EffectPool {
	gains := settings.NewGains()
	effs := make([]*Effect, max)
	for i := range effs {
		effs[i] = new(Effect)
		effs[i].state = EffFree
		effs[i].Clear()
		effs[i].gains = gains
	}
	return &EffectPool{
		effects: effs,
		gains:   gains,
		gain:    q16.FromInt(1),
	}
}

// EffectPool はエフェクト管理構造体。ヒープ上に確保されたスライスを介して Effect ポインタを管理します。
type EffectPool struct {
	effects []*Effect // ヒープ上のスライス
	gain    q16.Fixed
	gains   *settings.Gains
}

func (e *EffectPool) Gain() q16.Fixed {
	return e.gain
}

func (e *EffectPool) SetGain(gain q16.Fixed) {
	e.gain = gain
}

// Clear はすべてのエフェクトをリセットします。
func (e *EffectPool) Clear() {
	for _, effect := range e.effects {
		effect.state = EffFree
		effect.Clear()
	}
}

func (e *EffectPool) Len() int {
	return len(e.effects)
}

// Add は空いたスロットに新しいエフェクトを追加します。
// 返値は0:空きなし、1以上:エフェクトID
func (e *EffectPool) Allocate() uint8 {
	for i, effect := range e.effects {
		if effect.state == EffFree {
			effect.Clear()
			effect.state = EffAllocated
			//println("Allocated Effect:", i+1)
			return uint8(i + 1)
		}
	}
	//println("Allocate Failed: No Free Effects")
	return 0
}

// Get は指定インデックスのエフェクトを返します。
func (e *EffectPool) Get(idx uint8) *Effect {
	if idx <= 0 || int(idx) > len(e.effects) {
		return nil
	}
	res := e.effects[idx-1]
	return res
}

func (e *EffectPool) Start(id uint8) {
	id = id - 1
	if int(id) >= 0 && int(id) < len(e.effects) {
		e.effects[id].Start()
	}
}

func (e *EffectPool) Stop(id uint8) {
	id = id - 1
	if int(id) >= 0 && int(id) < len(e.effects) {
		e.effects[id].Stop()
	}
}

func (e *EffectPool) StopAll() {
	for _, effect := range e.effects {
		effect.Stop()
	}
}

func (e *EffectPool) Free(id uint8) {
	id = id - 1
	if int(id) >= 0 && int(id) < len(e.effects) {
		e.effects[id].Free()
	}
}

func (e *EffectPool) FreeAll() {
	for _, effect := range e.effects {
		effect.Free()
	}
}

func (e *EffectPool) Gains() *settings.Gains {
	return e.gains
}

// Calc はすべてのエフェクトのトルクを計算して合計します。
func (e *EffectPool) Calc(params *Params, axis int) q16.Fixed {
	output := q16.Zero
	for _, effect := range e.effects {
		for i := range MaxAxisCount {
			output += effect.Calc(params, i)
		}
	}
	return q16.Mul(e.gains.TotalGain, q16.Mul(e.gain, output))
}
