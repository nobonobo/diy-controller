// Package pid はHID USB Force Feedbackプロトコルの実装を提供します。
// レースゲームの力覚フィードバック効果を制御するためのデータ構造とロジックを含みます。
package pid

const (
	MAX_EFFECTS        = 10 // 同時に管理できるエフェクトの最大数
	MAX_FFB_AXIS_COUNT = 2  // Force Feedback対応軸の最大数（X, Y）
)

// SIZE_EFFECT はTEffectState構造体のサイズです。
var (
	SIZE_EFFECT = uint16(128)
	// MEMORY_SIZE はForce Bufferで確保されるメモリ総サイズを計算します。
	MEMORY_SIZE = SIZE_EFFECT * MAX_EFFECTS
)

// EffectState はエフェクトの再生状態を表します。
type EffectState uint8

const (
	// MEFFECTSTATE_FREE はエフェクトブロックが未使用（解放済み）状態。
	MEFFECTSTATE_FREE EffectState = 0x00

	// MEFFECTSTATE_ALLOCATED はエフェクトブロックが割り当てられ、パラメータ設定待ちの状態。
	MEFFECTSTATE_ALLOCATED EffectState = 0x01

	// MEFFECTSTATE_PLAYING はエフェクトが再生中の状態。
	MEFFECTSTATE_PLAYING EffectState = 0x02
)

// HID Usage Code: 軸の識別子（HID Usage Tables仕様）
const (
	UsageX        = 0x30 // X軸（一般的な横方向の動き）
	UsageY        = 0x31 // Y軸（一般的な縦方向の動き）
	UsageRx       = 0x33 // X回転軸（ローテーション）
	UsageRy       = 0x34 // Y回転軸（ピッチ）
	UsageSteering = 0xc8 // ステアリング（HID Vendor Defined Usage 0xC8）
)

// HID Report ID: Force Feedbackデバイスとの通信で使用されるレポート識別子。
const (
	// ReportPIDStatusInputData はホストからのステータス入力データを表します。
	ReportPIDStatusInputData = 0x02

	// --- レポートタイプ（出力レポート）---

	// ReportSetEffect はエフェクトのパラメータを設定します。
	ReportSetEffect = 0x01
	// ReportSetEnvelope はエンベロープ（Attack/Fade）パラメータを設定します。
	ReportSetEnvelope = 0x02
	// ReportSetCondition はばね/摩擦/慣性の条件パラメータを設定します。
	ReportSetCondition = 0x03
	// ReportSetPeriodic は周期性エフェクト（正弦波など）のパラメータを設定します。
	ReportSetPeriodic = 0x04
	// ReportSetConstantForce は定常力の強さを設定します。
	ReportSetConstantForce = 0x05
	// ReportSetRampForce はランプ力（開始/終了値）を設定します。
	ReportSetRampForce = 0x06
	// ReportSetCustomForceData はカスタム力波形のデータブロックを設定します。
	ReportSetCustomForceData = 0x07
	// ReportSetDownloadForceSample はダウンロードされた力サンプルを設定します。
	ReportSetDownloadForceSample = 0x08
	// ReportEffectOperation はエフェクトの再生操作（開始/停止）を実行します。
	ReportEffectOperation = 0x0a
	// ReportBlockFree はエフェクトブロックを解放（削除）します。
	ReportBlockFree = 0x0b
	// ReportDeviceControl はデバイスの制御コマンドを送信します。
	ReportDeviceControl = 0x0c
	// ReportDeviceGain はデバイスゲインを設定します。
	ReportDeviceGain = 0x0d
	// ReportSetCustomForce はカスタム力エフェクトのパラメータを設定します。
	ReportSetCustomForce = 0x0e

	// ReportNewEffectBlock は新しいエフェクトブロックの作成をリクエストします。
	ReportNewEffectBlock = 0x11
	// ReportLoadEffectBlock はエフェクトブロックの読み込みをリクエストします。
	ReportLoadEffectBlock = 0x12
	// ReportPIDPool はPIDプール（メモリリソース）情報を取得します。
	ReportPIDPool = 0x13
)

// ControlType値。
const (
	// ControlEnableActuators はアクチュエータを有効にします。
	ControlEnableActuators ControlType = 0x01
	// ControlDisableActuators はアクチュエータを無効にします。
	ControlDisableActuators ControlType = 0x02
	// ControlStopAllEffects はすべてのエフェクトを停止します。
	ControlStopAllEffects ControlType = 0x03
	// ControlReset はデバイスをリセットし、すべてのエフェクトを削除します。
	ControlReset ControlType = 0x04
	// ControlPause はすべてのエフェクトを一時停止します。
	ControlPause ControlType = 0x05
	// ControlContinue は一時停止中のエフェクトを再開します。
	ControlContinue ControlType = 0x06
)

// EffectType値（HID Force Feedbackエフェクトタイプ）。
const (
	// USB_EFFECT_CONSTANT は定常力エフェクト（常に一定の力が加わります）。
	USB_EFFECT_CONSTANT EffectType = 0x01
	// USB_EFFECT_RAMP はランプ力エフェクト（線形に増加/減少する力）。
	USB_EFFECT_RAMP EffectType = 0x02
	// USB_EFFECT_SQUARE は正方形波エフェクト（最大値と0を切り替え）。
	USB_EFFECT_SQUARE EffectType = 0x03
	// USB_EFFECT_SINE は正弦波エフェクト。
	USB_EFFECT_SINE EffectType = 0x04
	// USB_EFFECT_TRIANGLE は三角波エフェクト。
	USB_EFFECT_TRIANGLE EffectType = 0x05
	// USB_EFFECT_SAWTOOTHDOWN はノコギリ波（下がり勾配）。
	USB_EFFECT_SAWTOOTHDOWN EffectType = 0x06
	// USB_EFFECT_SAWTOOTHUP はノコギリ波（上がり勾配）。
	USB_EFFECT_SAWTOOTHUP EffectType = 0x07
	// USB_EFFECT_SPRING はばねエフェクト（位置基準の復元力）。
	USB_EFFECT_SPRING EffectType = 0x08
	// USB_EFFECT_DAMPER はダンパーエフェクト（速度比例の減衰力）。
	USB_EFFECT_DAMPER EffectType = 0x09
	// USB_EFFECT_INERTIA は慣性エフェクト（加速度比例の力）。
	USB_EFFECT_INERTIA EffectType = 0x0A
	// USB_EFFECT_FRICTION は摩擦エフェクト（移動抵抗としての力）。
	USB_EFFECT_FRICTION EffectType = 0x0B
	// USB_EFFECT_CUSTOM はカスタム力波形エフェクト。
	USB_EFFECT_CUSTOM EffectType = 0x0C
)

// EffectOperation値。
const (
	// EOStart はエフェクトを通常開始します（他のエフェクトと同時に再生される可能性あり）。
	EOStart EffectOperation = 1
	// EOStartSolo は他のエフェクトを停止してこのエフェクトのみを単独で開始します。
	EOStartSolo EffectOperation = 2
	// EOStop は指定したエフェクトを停止します。
	EOStop EffectOperation = 3
)

// エフェクト軸有効化ビットマスク。
const (
	// X_AXIS_ENABLE **unused** はX軸のエフェクトを有効にするビットマスク。
	X_AXIS_ENABLE = 0x01
	// Y_AXIS_ENABLE **unused** はY軸のエフェクトを有効にするビットマスク。
	Y_AXIS_ENABLE = 0x02
	// DIRECTION_ENABLE **unused** は方向ベクトルによるエフェクト有効化。
	DIRECTION_ENABLE = 0x04
)

// USB_DURATION_INFINITE はエフェクトの無限再生を指定する値。
const USB_DURATION_INFINITE = 0x7fff
