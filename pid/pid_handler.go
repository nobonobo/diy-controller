// Package pid はHID USB Force Feedbackプロトコルの実装を提供します。
// PIDHandler はホスト（PC）からのForce Feedbackコマンドを受信・処理し、
// エフェクトの管理、力覚計算、USB HIDレポートの送受信を制御します。
package pid

import (
	"errors"
	"machine"
	"machine/usb"
	"machine/usb/hid"

	"github.com/nobonobo/q16"

	"github.com/nobonobo/diy-controller/effects"
)

// ============================================================================
// グローバル変数
// ============================================================================

// buffer はdump関数の出力バッファです。
// 1024バイトの静的メモリを確保し、ヒープ割り当てを回避します。
var buffer = make([]byte, 1024)

// dump はバイナリデータを16進文字列に変換してダンプします。
// NOTE: 現在はこの関数が呼び出されていません（SetEffect内のmachine.Serial.Writeがコメントアウトされています）。
//
// Parameters:
//   - src: ダンプ対象のバイナリデータ
//
// Returns:
//   - 16進文字列に変換されたバッファスライス（末尾に改行付き）
//
// Example: []byte{0x01, 0x0a} -> "010a\n"
func dump(src []byte) []byte {
	const hex = "0123456789abcdef"
	for i, b := range src {
		buffer[i*2] = hex[b>>4]
		buffer[i*2+1] = hex[b&0x0f]
	}
	buffer[len(src)*2] = '\n'
	return buffer[:len(src)*2+1]
}

// ============================================================================
// PIDHandler 構造体
// ============================================================================

// PIDHandler はForce Feedbackデバイスの主要制御構造体です。
// ホストからのHIDコマンドを受信し、エフェクトの状態管理、
// 力覚フィードバックの計算、USB HIDレポートの送受信を処理します。
type PIDHandler struct {
	// effectPool はエフェクト状態を管理するプールです。
	effectPool *effects.EffectPool

	// pidBlockLoad はHIDレポート0x06（PID Block Load Feature Report）のデータです。
	// ホストがエフェクトブロックの読み込みステータスをクエリする際に使用します。
	pidBlockLoad PIDBlockLoadFeatureData

	// pidPool はHIDレポート0x07（PID Pool Feature Report）のデータです。
	// デバイスのPIDメモリプールの情報をホストに返します。
	pidPool PIDPoolFeatureData

	// enabled はアクチュエータが有効かどうかの状態を示すフラグです。
	// trueの場合、力覚フィードバックが出力されます。
	enabled bool

	// paused はエフェクトが一時停止中かどうかを示すフラグです。
	// trueの場合、CalcForces()はゼロ力を返します。
	paused bool

	cnt int // for debug
}

// NewPIDHandler はPIDHandlerの新しいインスタンスを生成して返します。
// 全てのエフェクト状態をMEFFECTSTATE_FREEに初期化し、
// PIDプール情報とデフォルトゲイン値を設定します。
//
// Parameters:
//   - effectPool: エフェクト状態を管理するプール
//
// Returns:
//   - 初期化されたPIDHandlerポインタ
func NewPIDHandler(effectPool *effects.EffectPool) *PIDHandler {
	handler := &PIDHandler{
		effectPool: effectPool,
		pidBlockLoad: PIDBlockLoadFeatureData{
			RamPoolAvailable: MEMORY_SIZE,
			b:                make([]byte, 5),
		},
		pidPool: PIDPoolFeatureData{
			ReportID:               7,
			RamPoolSize:            MEMORY_SIZE,
			MaxSimultaneousEffects: MAX_EFFECTS,
			MemoryManagement:       3,
			b:                      make([]byte, 5),
		},
	}
	return handler
}

// ============================================================================
// HIDレポート処理 — ホストからのコマンド受信
// ============================================================================

// RxHandler はホストからの出力レポート（HID Output Report）を受信した際に呼び出されます。
// レポートIDに基づいて対応するハンドラメソッドをディスパッチします。
// これはInterruptOut（USBインターフェースのOUTエンドポイント）からのデータを処理します。
//
// Parameters:
//   - b: ホストから受信したバイナリデータ
func (m *PIDHandler) RxHandler(b []byte) {
	if len(b) == 0 {
		return
	}
	// machine.Serial.Write(dump(b))
	reportId := b[0]
	switch reportId {
	case ReportSetEffect: // 0x01 — エフェクトパラメータ設定
		m.SetEffect(b)
	case ReportSetEnvelope: // 0x02 — エンベロープ（Attack/Fade）パラメータ設定
		m.SetEnvelope(b)
	case ReportSetCondition: // 0x03 — ばね/摩擦/慣性条件パラメータ設定
		m.SetCondition(b)
	case ReportSetPeriodic: // 0x04 — 周期性エフェクトパラメータ設定
		m.SetPeriodic(b)
	case ReportSetConstantForce: // 0x05 — 定常力パラメータ設定
		m.SetConstantForce(b)
	case ReportSetRampForce: // 0x06 — ランプ力パラメータ設定
		m.SetRampForce(b)
	case ReportSetCustomForceData: // 0x07 — カスタム力波形データ設定
		m.SetCustomForceData(b)
	case ReportSetDownloadForceSample: // 0x08 — ダウンロード力サンプル設定
		m.SetDownloadForceSample(b)
	case ReportEffectOperation: // 0x0a — エフェクト操作（開始/停止）
		m.EffectOperation(b)
	case ReportBlockFree: // 0x0b — エフェクトブロック解放
		m.BlockFree(b)
	case ReportDeviceControl: // 0x0c — デバイス制御コマンド
		m.DeviceControl(b)
	case ReportDeviceGain: // 0x0d — デバイスゲイン設定
		m.DeviceGain(b)
	case ReportSetCustomForce: // 0x0e — カスタム力エフェクトパラメータ設定
		m.SetCustomForce(b)
	}
}

// ============================================================================
// USB HID制御要求ハンドラ
// ============================================================================

var ErrorOutOfMemory = errors.New("out of memory")

// CreateNewEffect は新しいエフェクトブロックの作成を処理します。
// ホストからのCreate New Effect Feature Report（レポートID 0x11）に対応し、
// 空いているエフェクトブロックを割り当てて状態をMEFFECTSTATE_ALLOCATEDに設定します。
//
// Parameters:
//   - data: ホストから受信したCreateNewEffectFeatureData
//
// Returns:
//   - エラーが発生した場合（既に満杯など）
func (m *PIDHandler) CreateNewEffect(data CreateNewEffectFeatureData) error {
	m.pidBlockLoad.ReportID = 6
	m.pidBlockLoad.EffectBlockIndex = m.effectPool.Allocate()
	if m.pidBlockLoad.EffectBlockIndex == 0 {
		m.pidBlockLoad.LoadStatus = 2 // 1=Success, 2=Full, 3=Error
		return ErrorOutOfMemory
	}
	if m.pidBlockLoad.RamPoolAvailable > SIZE_EFFECT {
		m.pidBlockLoad.RamPoolAvailable -= SIZE_EFFECT
		m.pidBlockLoad.LoadStatus = 1 // 1=Success, 2=Full, 3=Error
	} else {
		m.pidBlockLoad.LoadStatus = 2 // 1=Success, 2=Full, 3=Error
		return ErrorOutOfMemory
	}
	return nil
}

// GetReport はHID INPUTレポートおよびFEATUREレポートのリクエストに応答してデータを返します。
// USB制御エンドポイントからのGET_REPORT要求を処理します。
//
// Parameters:
//   - setup: USBセットアップパケット（リクエスト情報を含む）
//
// Returns:
//   - true: 処理が成功した場合、false: 未対応のレポートIDなど
func (m *PIDHandler) GetReport(setup usb.Setup) bool {
	reportId := setup.WValueL
	switch setup.WValueH {
	case hid.REPORT_TYPE_INPUT:
	case hid.REPORT_TYPE_OUTPUT:
	case hid.REPORT_TYPE_FEATURE:
		switch reportId {
		case 6: // PID Block Load Feature Report
			b, _ := m.pidBlockLoad.MarshalBinary()
			machine.SendUSBInPacket(0, b)
			return true
		case 7: // PID Pool Feature Report
			b, _ := m.pidPool.MarshalBinary()
			machine.SendUSBInPacket(0, b)
			return true
		}
	}
	return false
}

// GetIdle はHID SET_IDLE/GET_IDLE リクエストを処理します。
// IDLE状態のインターバル値を取得しますが、現在は単純にZLP（ゼロ長さパケット）を返すだけです。
//
// Parameters:
//   - setup: USBセットアップパケット
//
// Returns:
//   - true: 常にtrue
func (m *PIDHandler) GetIdle(setup usb.Setup) bool {
	machine.SendZlp()
	return true
}

// GetProtocol はHID GET_PROTOCOL リクエストを処理します。
// プロトコルモード（BOOT MODEまたはREPORT PROTOCOL）を取得しますが、
// 現在は単純にZLPを返すだけです。
//
// Parameters:
//   - setup: USBセットアップパケット
//
// Returns:
//   - true: 常にtrue
func (m *PIDHandler) GetProtocol(setup usb.Setup) bool {
	machine.SendZlp()
	return true
}

// SetReport はHID SET_REPORT リクエストを処理します。
// ホストからのエフェクトブロック作成リクエスト（レポートID 0x05）を解析し、
// CreateNewEffectを呼び出します。
//
// Parameters:
//   - setup: USBセットアップパケット
//
// Returns:
//   - true: 処理が成功した場合
func (m *PIDHandler) SetReport(setup usb.Setup) bool {
	//println("set report:", setup.WValueL, setup.WValueH, setup.WIndex, setup.WLength)
	reportId := setup.WValueL
	switch setup.WValueH {
	case hid.REPORT_TYPE_INPUT:
		machine.SendZlp()
		return true
	case hid.REPORT_TYPE_OUTPUT:
		machine.SendZlp()
		return true
	case hid.REPORT_TYPE_FEATURE:
		if setup.WLength == 0 {
			machine.ReceiveUSBControlPacket()
			machine.SendZlp()
			return true
		}
		if reportId == 5 {
			b, err := machine.ReceiveUSBControlPacket()
			if err != nil {
				//println("ReceiveUSBControlPacket Failed:", err.Error())
				return false
			}
			v := CreateNewEffectFeatureData{}
			v.UnmarshalBinary(b[:4])
			if err := m.CreateNewEffect(v); err != nil {
				//println("CreateNewEffect Failed:", err.Error())
				return false
			}
			machine.SendZlp()
			return true
		}
	}
	return false
}

// SetIdle はHID SET_IDLE リクエストを処理します。
// IDLE状態のインターバル値を設定しますが、現在は単純にZLPを返すだけです。
//
// Parameters:
//   - setup: USBセットアップパケット
//
// Returns:
//   - true: 常にtrue
func (m *PIDHandler) SetIdle(setup usb.Setup) bool {
	machine.SendZlp()
	return true
}

// SetProtocol はHID SET_PROTOCOL リクエストを処理します。
// プロトコルモードを設定しますが、現在は単純にZLPを返すだけです。
//
// Parameters:
//   - setup: USBセットアップパケット
//
// Returns:
//   - true: 常にtrue
func (m *PIDHandler) SetProtocol(setup usb.Setup) bool {
	machine.SendZlp()
	return true
}

// SetupHandler はUSB制御エンドポイントでのSETUPパケットを処理します。
// リクエストタイプ（DEVICE_TO_HOST/HOST_TO_DEVICE）とBRequestに基づいて、
// 対応するハンドラメソッド（GetReport, SetReport, GetIdle, SetIdle, etc.）にディスパッチします。
//
// Parameters:
//   - setup: USBセットアップパケット
//
// Returns:
//   - true: 処理が成功した場合、false: 未対応のリクエスト
func (m *PIDHandler) SetupHandler(setup usb.Setup) bool {
	//println("setup:", setup.BmRequestType, setup.BRequest)
	switch setup.BmRequestType {
	case usb.REQUEST_DEVICETOHOST_CLASS_INTERFACE: //usb.GET_REPORT: // デバイスからホストへ（クラス指定、インターフェース）
		switch setup.BRequest {
		case usb.GET_REPORT:
			return m.GetReport(setup)
		case usb.GET_IDLE:
			return m.GetIdle(setup)
		case usb.GET_PROTOCOL:
			return m.GetProtocol(setup)
		}
	case usb.REQUEST_HOSTTODEVICE_CLASS_INTERFACE: // usb.SET_REPORT: // ホストからデバイスへ（クラス指定、インターフェース）
		switch setup.BRequest {
		case usb.SET_REPORT:
			return m.SetReport(setup)
		case usb.SET_IDLE:
			return m.SetIdle(setup)
		case usb.SET_PROTOCOL:
			return m.SetProtocol(setup)
		}
	}
	return false
}

// ============================================================================
// エフェクト管理 — 割り当て・開始・停止・解放
// ============================================================================

// StopAllEffects はすべてのエフェクトを停止します。
// 各エフェクトのStopEffect()を呼び出します。
func (m *PIDHandler) StopAllEffects() {
	m.effectPool.StopAll()
}

// StartEffect は指定されたインデックスのエフェクトを開始します。
// エフェクトの状態にMEFFECTSTATE_PLAYINGをセットし、
// 経過時間をリセットして開始時刻を更新します。
//
// Parameters:
//   - id: 開始するエフェクトのブロックインデックス
func (m *PIDHandler) StartEffect(id uint8) {
	m.effectPool.Start(id)
}

// StopEffect は指定されたインデックスのエフェクトを停止します。
// エフェクトの状態からMEFFECTSTATE_PLAYINGフラグをクリアし、
// RAMプールの利用可能量を元に戻します。
//
// Parameters:
//   - id: 停止するエフェクトのブロックインデックス
func (m *PIDHandler) StopEffect(id uint8) {
	m.effectPool.Stop(id)
}

// FreeAllEffects はすべてのエフェクトを解放します。
// 全エフェクト状態をMEFFECTSTATE_FREEにリセットし、
// RAMプールの利用可能量を総サイズに戻し、nextEIDを1にリセットします。
func (m *PIDHandler) FreeAllEffects() {
	m.effectPool.FreeAll()
	m.pidBlockLoad.RamPoolAvailable = MEMORY_SIZE
	//println("FreeAllEffects", MEMORY_SIZE)
}

// FreeEffect は指定されたインデックスのエフェクトを解放します。
// エフェクト状態をMEFFECTSTATE_FREEに設定し、
// nextEIDを必要に応じて更新して再利用可能にします。
//
// Parameters:
//   - id: 解放するエフェクトのブロックインデックス
func (m *PIDHandler) FreeEffect(id uint8) {
	m.effectPool.Free(id)
	if m.pidBlockLoad.RamPoolAvailable != 0 {
		m.pidBlockLoad.RamPoolAvailable += SIZE_EFFECT
	}
}

// ============================================================================
// エフェクトパラメータ設定 — 各HIDレポート対応
// ============================================================================

// SetEffect はHIDレポート0x01（Set Effect Output）を処理します。
// エフェクトの基本パラメータ（持続時間、方向、ゲイン、タイプ、軸有効化など）を設定します。
//
// Parameters:
//   - b: ホストから受信したバイナリデータ
func (m *PIDHandler) SetEffect(b []byte) {
	var v SetEffectOutputData
	_ = v.UnmarshalBinary(b)
	effect := m.effectPool.Get(v.EffectBlockIndex)
	if effect == nil {
		return
	}
	effect.SetEffectParam(effects.EffectParam{
		EffectType:            effects.EffectType(v.EffectType),
		Duration:              q16.Fixed(int32(v.Duration) * q16.Scale / 1000),
		TriggerRepeatInterval: q16.Fixed(int32(v.TriggerRepeatInterval) * q16.Scale / 1000),
		SamplePeriod:          q16.Fixed(int32(v.SamplePeriod) * q16.Scale / 1000),
		Gain:                  q16.Fixed(int32(v.Gain) * q16.Scale / 255),
		TriggerButton:         v.TriggerButton,
		EnableAxis:            v.EnableAxis,
		DirectionX:            v.DirectionX,
		DirectionY:            v.DirectionY,
		StartDelay:            q16.Fixed(int32(v.StartDelay) * q16.Scale / 1000),
	})
	//println("SetEffect:", v.EffectBlockIndex, v.EffectType, v.Gain)
}

// SetEnvelope はHIDレポート0x02（Set Envelope Output）を処理します。
// エフェクトのエンベロープパラメータ（Attackレベル/時間、Fadeレベル/時間）を設定します。
//
// Parameters:
//   - b: ホストから受信したバイナリデータ
func (m *PIDHandler) SetEnvelope(b []byte) {
	var v SetEnvelopeOutputData
	_ = v.UnmarshalBinary(b)
	effect := m.effectPool.Get(v.EffectBlockIndex)
	if effect == nil {
		return
	}
	effect.SetEnvelope(effects.Envelope{
		AttackLevel: q16.Fixed(int32(v.AttackLevel) * q16.Scale / 255),
		FadeLevel:   q16.Fixed(int32(v.FadeLevel) * q16.Scale / 255),
		AttackTime:  q16.Fixed(int32(v.AttackTime) * q16.Scale / 1000),
		FadeTime:    q16.Fixed(int32(v.FadeTime) * q16.Scale / 1000),
	})
}

// SetCondition はHIDレポート0x03（Set Condition Output）を処理します。
// ばね・摩擦・慣性エフェクトの条件パラメータ（オフセット、係数、飽和値、デッドバンド）を設定します。
//
// Parameters:
//   - b: ホストから受信したバイナリデータ
func (m *PIDHandler) SetCondition(b []byte) {
	var v SetConditionOutputData
	_ = v.UnmarshalBinary(b)
	axis := v.ParameterBlockOffset
	effect := m.effectPool.Get(v.EffectBlockIndex)
	if effect == nil {
		return
	}
	//println("SetCondition:", v.CpOffset, v.PositiveCoefficient, v.NegativeCoefficient, v.PositiveSaturation, v.NegativeSaturation, v.DeadBand)
	effect.SetCondition(axis, effects.Condition{
		CpOffset:            q16.Fixed(int32(v.CpOffset) * q16.Scale / 127),
		PositiveCoefficient: q16.Fixed(int32(v.PositiveCoefficient) * q16.Scale / 10000),
		NegativeCoefficient: q16.Fixed(int32(v.NegativeCoefficient) * q16.Scale / 10000),
		PositiveSaturation:  q16.Fixed(int32(v.PositiveSaturation) * q16.Scale / 10000),
		NegativeSaturation:  q16.Fixed(int32(v.NegativeSaturation) * q16.Scale / 10000),
		DeadBand:            q16.Fixed(int32(v.DeadBand) * q16.Scale / 10000),
	})
}

// SetPeriodic はHIDレポート0x04（Set Periodic Force Output）を処理します。
// 周期性エフェクトのパラメータ（振幅、オフセット、位相、周期）を設定します。
//
// Parameters:
//   - b: ホストから受信したバイナリデータ
func (m *PIDHandler) SetPeriodic(b []byte) {
	var v SetPeriodicOutputData
	_ = v.UnmarshalBinary(b)
	effect := m.effectPool.Get(v.EffectBlockIndex)
	if effect == nil {
		return
	}
	effect.SetPeriodicParam(effects.PeriodicParam{
		Magnitude: q16.Fixed(int32(v.Magnitude) * q16.Scale / 32767),
		Offset:    q16.Fixed(int32(v.Offset) * q16.Scale / 127),
		Phase:     q16.Mul(q16.Fixed(int32(v.Phase)*q16.Scale/255), q16.Period),
		Period:    q16.Fixed(int32(v.Period) * q16.Scale / 1000),
	})
}

// SetConstantForce はHIDレポート0x05（Set Constant Force Output）を処理します。
// 定常力エフェクトのパラメータ（力覚値）を設定します。
//
// Parameters:
//   - b: ホストから受信したバイナリデータ
func (m *PIDHandler) SetConstantForce(b []byte) {
	var v SetConstantForceOutputData
	_ = v.UnmarshalBinary(b)
	effect := m.effectPool.Get(v.EffectBlockIndex)
	if effect == nil {
		return
	}
	p := effect.PeriodicParam()
	p.Magnitude = q16.Fixed(int32(v.Magnitude) * q16.Scale / 32767)
	effect.SetPeriodicParam(p)
}

// SetRampForce はHIDレポート0x06（Set Ramp Force Output）を処理します。
// ランプ力エフェクトのパラメータ（開始/終了の力覚値）を設定します。
//
// Parameters:
//   - b: ホストから受信したバイナリデータ
func (m *PIDHandler) SetRampForce(b []byte) {
	var v SetRampForceOutputData
	_ = v.UnmarshalBinary(b)
	effect := m.effectPool.Get(v.EffectBlockIndex)
	if effect == nil {
		return
	}
	effect.SetRampParam(effects.RampParam{
		StartMagnitude: q16.Fixed(int32(v.StartMagnitude) * q16.Scale / 32767),
		EndMagnitude:   q16.Fixed(int32(v.EndMagnitude) * q16.Scale / 32767),
	})
}

// SetCustomForceData はHIDレポート0x07（Set Custom Force Data Output）を処理します。
// カスタム力波形のデータブロックを設定します。（未実装）
//
// Parameters:
//   - b: ホストから受信したバイナリデータ
func (m *PIDHandler) SetCustomForceData(b []byte) {
	var v SetCustomForceDataOutputData
	_ = v.UnmarshalBinary(b)
	// TODO: implement
}

// SetDownloadForceSample はHIDレポート0x08（Set Download Force Sample Output）を処理します。
// ダウンロードされた力サンプルデータを設定します。（未実装）
//
// Parameters:
//   - b: ホストから受信したバイナリデータ
func (m *PIDHandler) SetDownloadForceSample(b []byte) {
	var v SetDownloadForceSampleOutputData
	_ = v.UnmarshalBinary(b)
	// TODO: implement
}

// EffectOperation はHIDレポート0x0A（Effect Operation Output）を処理します。
// エフェクトの再生操作（開始、単独開始、停止）を実行します。
// LoopCountが0xFFの場合は無限ループとして扱います。
//
// Parameters:
//   - b: ホストから受信したバイナリデータ
func (m *PIDHandler) EffectOperation(b []byte) {
	var v EffectOperationOutputData
	_ = v.UnmarshalBinary(b)
	switch v.Operation {
	case EOStart:
		effect := m.effectPool.Get(v.EffectBlockIndex)
		if effect == nil {
			return
		}
		switch v.LoopCount {
		case 0xff:
			effect.SetDuration(q16.Zero)
		default:
			effect.SetDuration(q16.Mul(effect.EffectParam().Duration, q16.FromInt(int(v.LoopCount))))
		}
		m.StartEffect(v.EffectBlockIndex)
	case EOStartSolo:
		m.StopAllEffects()
		m.StartEffect(v.EffectBlockIndex)
	case EOStop:
		m.StopEffect(v.EffectBlockIndex)
	}
}

// BlockFree はHIDレポート0x0B（Block Free Output）を処理します。
// 指定されたエフェクトブロックを解放します。
// EffectBlockIndexが0xFFの場合は全エフェクトを解放します。
//
// Parameters:
//   - b: ホストから受信したバイナリデータ
func (m *PIDHandler) BlockFree(b []byte) {
	var v BlockFreeOutputData
	_ = v.UnmarshalBinary(b)
	if v.EffectBlockIndex == 0xff {
		m.FreeAllEffects()
		return
	}
	m.effectPool.Free(v.EffectBlockIndex)
}

// DeviceControl はHIDレポート0x0C（Device Control Output）を処理します。
// デバイスの制御コマンド（アクチュエータ有効/無効、全停止、リセット、一時停止/再開）を実行します。
//
// Parameters:
//   - b: ホストから受信したバイナリデータ
func (m *PIDHandler) DeviceControl(b []byte) {
	var v DeviceControlOutputData
	_ = v.UnmarshalBinary(b)
	switch v.Control {
	case ControlEnableActuators:
		m.enabled = true
	case ControlDisableActuators:
		m.enabled = false
	case ControlStopAllEffects:
		m.effectPool.StopAll()
	case ControlReset:
		m.FreeAllEffects()
	case ControlPause:
		m.paused = true
	case ControlContinue:
		m.paused = false
	}
}

// DeviceGain はHIDレポート0x0D（Device Gain Output）を処理します。
// デバイスのゲイン値を設定します。この値はCalcForces()で力覚計算時に使用されます。
//
// Parameters:
//   - b: ホストから受信したバイナリデータ
func (m *PIDHandler) DeviceGain(b []byte) {
	var v DeviceGainOutputData
	_ = v.UnmarshalBinary(b)
	m.effectPool.SetGain(q16.Fixed(int32(v.Gain) * q16.Scale / 255))
}

// SetCustomForce はHIDレポート0x0E（Set Custom Force Output）を処理します。
// カスタム力エフェクトのパラメータを設定します。（未実装）
//
// Parameters:
//   - b: ホストから受信したバイナリデータ
func (m *PIDHandler) SetCustomForce(b []byte) {
	var v SetCustomForceOutputData
	_ = v.UnmarshalBinary(b)
	// TODO: implement
}

// ============================================================================
// 力覚計算 — コア機能
// ============================================================================

// Calc はすべてのアクティブなエフェクトの力を合計して計算します。
// この関数はメインループから定期的に呼び出されることを想定しています。
//
// 処理フロー:
//  1. pausedがtrueの場合はゼロ力を返す
//  2. 各エフェクトについて、ALLOCATEDかつPLAYING状態か確認
//  3. 期限切れのエフェクトは自動停止
//  4. 各エフェクトのForce()メソッドを呼び出し、X/Y軸の力を合計
//
// Returns:
//   - q16.Fixed: 軸力
func (m *PIDHandler) Calc(params *effects.Params, axis int) q16.Fixed {
	if !m.enabled || m.paused {
		return q16.Zero
	}
	return m.effectPool.Calc(params, axis)
}
