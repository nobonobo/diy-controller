// Package pid はHID USB Force Feedbackプロトコルの実装を提供します。
// レースゲームの力覚フィードバック効果を制御するためのデータ構造とロジックを含みます。
package pid

import "encoding/binary"

// ReportID はHIDレポートの種類を識別します。
type ReportID uint8

// ControlType はデバイスの制御コマンドを表します。
type ControlType uint8

// EffectType はForce Feedbackエフェクトのタイプを表します（12種類）。
type EffectType uint8

// EffectOperation はエフェクトの操作（開始/停止）を表します。
type EffectOperation uint8

// --- 主要構造体 ---

// PIDStatusInputData はHIDレポート0x02（PID Status Input Data）の構造体です。
// ホストがデバイスの現在のステータスをクエリするために使用します。
type PIDStatusInputData struct {
	ReportID         ReportID // レポートID（常に0x02）
	Status           uint8    // ステータスビット: Bit0=デバイス一時停止中, Bit1=アクチュエータ有効, Bit2=安全スイッチ, Bit3=アクチュエータオーバーライドスイッチ, Bit4=アクチュエータ電源
	EffectBlockIndex uint8    // Bit7=エフェクト再生中, Bit0-7=エフェクトブロックインデックス（1-40）
}

// SetEffectOutputData はHIDレポート0x01（Set Effect）の構造体です。
// エフェクトのパラメータをデバイスに設定するために使用します。
type SetEffectOutputData struct {
	ReportID              ReportID   // レポートID（常に0x01）
	EffectBlockIndex      uint8      // エフェクトブロックインデックス（1-40）
	EffectType            EffectType // エフェクトタイプ（定数参照）
	Duration              uint16     // エフェクトの継続時間（0-32767 ms、0x7FFFは無限）
	TriggerRepeatInterval uint16     // トリガー再発生インターバル（0-32767 ms）
	SamplePeriod          uint16     // サンプリング期間（0-32767 ms）
	Gain                  uint8      // ゲイン値（0-255、物理的には0-10000にマッピング可能）
	TriggerButton         uint8      // トリガーボタンID（0-8、0は常時有効）
	EnableAxis            uint8      // 軸有効化ビットマスク: Bit0=X, Bit1=Y, Bit2=DirectionEnable
	DirectionX            uint8      // X方向角度（0=0度 .. 255=360度）
	DirectionY            uint8      // Y方向角度（0=0度 .. 255=360度）
	StartDelay            uint16     // 開始遅延時間（0-32767 ms）
}

// UnmarshalBinary はSetEffectOutputDataをバイナリからパースします。
func (s *SetEffectOutputData) UnmarshalBinary(b []byte) error {
	s.ReportID = ReportID(b[0])
	s.EffectBlockIndex = b[1]
	s.EffectType = EffectType(b[2])
	s.Duration = binary.LittleEndian.Uint16(b[3:5])
	s.TriggerRepeatInterval = binary.LittleEndian.Uint16(b[5:7])
	s.SamplePeriod = binary.LittleEndian.Uint16(b[7:9])
	s.Gain = b[9]
	s.TriggerButton = b[10]
	s.EnableAxis = b[11]
	s.DirectionX = b[12]
	s.DirectionY = b[13]
	s.StartDelay = binary.LittleEndian.Uint16(b[14:16])
	return nil
}

// SetEnvelopeOutputData はHIDレポート0x02（Set Envelope）の構造体です。
// エフェクトのエンベロープ（Attack/Fade）パラメータを設定します。
type SetEnvelopeOutputData struct {
	ReportID         ReportID // レポートID（常に0x02）
	EffectBlockIndex uint8    // エフェクトブロックインデックス（1-40）
	AttackLevel      uint16   // Attackレベル（力覚の開始強度）
	FadeLevel        int16    // Fadeレベル（力覚の終了強度、符号あり）
	AttackTime       uint32   // Attack時間（ms、力が最大値まで増加するまでの時間）
	FadeTime         uint32   // Fade時間（ms、力が最終値まで減少する時間）
}

// UnmarshalBinary はSetEnvelopeOutputDataをバイナリからパースします。
func (s *SetEnvelopeOutputData) UnmarshalBinary(b []byte) error {
	s.ReportID = ReportID(b[0])
	s.EffectBlockIndex = b[1]
	s.AttackLevel = binary.LittleEndian.Uint16(b[2:4])
	s.FadeLevel = int16(binary.LittleEndian.Uint16(b[4:6]))
	s.AttackTime = binary.LittleEndian.Uint32(b[6:10])
	s.FadeTime = binary.LittleEndian.Uint32(b[10:14])
	return nil
}

// SetConditionOutputData はHIDレポート0x03（Set Condition）の構造体です。
// ばね・摩擦・慣性エフェクトの条件パラメータを設定します。
type SetConditionOutputData struct {
	ReportID             ReportID // レポートID（常に0x03）
	EffectBlockIndex     uint8    // エフェクトブロックインデックス（1-40）
	ParameterBlockOffset uint8    // パラメータブロックオフセット: Bit0-3=オフセット, Bit4-5=Instance1, Bit6-7=Instance2
	CpOffset             int16    // Center Point Offset（中心点オフセット、-128-127）
	PositiveCoefficient  int16    // 正の方向係数（-128-127、位置に対する力の比率）
	NegativeCoefficient  int16    // 負の方向係数（-128-127）
	PositiveSaturation   int16    // 正の飽和値（-128-127）
	NegativeSaturation   int16    // 負の飽和値（-128-127）
	DeadBand             uint16   // デッドバンド幅（0-255、中心付近の無感領域）
}

// UnmarshalBinary はSetConditionOutputDataをバイナリからパースします。
func (s *SetConditionOutputData) UnmarshalBinary(b []byte) error {
	s.ReportID = ReportID(b[0])
	s.EffectBlockIndex = b[1]
	s.ParameterBlockOffset = b[2]
	s.CpOffset = int16(binary.LittleEndian.Uint16(b[3:5]))
	s.PositiveCoefficient = int16(binary.LittleEndian.Uint16(b[5:7]))
	s.NegativeCoefficient = int16(binary.LittleEndian.Uint16(b[7:9]))
	s.PositiveSaturation = int16(binary.LittleEndian.Uint16(b[9:11]))
	s.NegativeSaturation = int16(binary.LittleEndian.Uint16(b[11:13]))
	s.DeadBand = binary.LittleEndian.Uint16(b[13:15])
	return nil
}

// SetPeriodicOutputData はHIDレポート0x04（Set Periodic Force）の構造体です。
// 周期性エフェクトのパラメータを設定します。
type SetPeriodicOutputData struct {
	ReportID         ReportID // レポートID（常に0x04）
	EffectBlockIndex uint8    // エフェクトブロックインデックス（1-40）
	Magnitude        int16    // 力覚の振幅（-32767-32767）
	Offset           int16    // オフセット（中心位置）
	Phase            uint16   // 位相（0-255、0-359度の範囲でexp-2）
	Period           uint32   // 周期（ms、0-32767）
}

// UnmarshalBinary はSetPeriodicOutputDataをバイナリからパースします。
func (s *SetPeriodicOutputData) UnmarshalBinary(b []byte) error {
	s.ReportID = ReportID(b[0])
	s.EffectBlockIndex = b[1]
	return nil
}

// SetConstantForceOutputData はHIDレポート0x05（Set Constant Force）の構造体です。
// 定常力の強さを設定します。
type SetConstantForceOutputData struct {
	ReportID         ReportID // レポートID（常に0x05）
	EffectBlockIndex uint8    // エフェクトブロックインデックス（1-40）
	Magnitude        int16    // 力覚値（-255-255）
}

// UnmarshalBinary はSetConstantForceOutputDataをバイナリからパースします。
func (s *SetConstantForceOutputData) UnmarshalBinary(b []byte) error {
	s.ReportID = ReportID(b[0])
	s.EffectBlockIndex = b[1]
	s.Magnitude = int16(binary.LittleEndian.Uint16(b[2:4]))
	return nil
}

// SetRampForceOutputData はHIDレポート0x06（Set Ramp Force）の構造体です。
// ランプ力エフェクトのパラメータを設定します。
type SetRampForceOutputData struct {
	ReportID         ReportID // レポートID（常に0x06）
	EffectBlockIndex uint8    // エフェクトブロックインデックス（1-40）
	StartMagnitude   int16    // 開始力覚値
	EndMagnitude     int16    // 終了力覚値
}

// UnmarshalBinary はSetRampForceOutputDataをバイナリからパースします。
func (s *SetRampForceOutputData) UnmarshalBinary(b []byte) error {
	s.ReportID = ReportID(b[0])
	s.EffectBlockIndex = b[1]
	s.StartMagnitude = int16(binary.LittleEndian.Uint16(b[2:4]))
	s.EndMagnitude = int16(binary.LittleEndian.Uint16(b[4:6]))
	return nil
}

// SetCustomForceDataOutputData はHIDレポート0x07（Set Custom Force Data）の構造体です。
// カスタム力波形のデータブロックを設定します。
type SetCustomForceDataOutputData struct {
	ReportID         ReportID // レポートID（常に0x07）
	EffectBlockIndex uint8    // エフェクトブロックインデックス（1-40）
	DataOffset       uint16   // データオフセット（波形データの開始位置）
	Data             [12]byte // カスタム力波形データ（int8配列、12バイト）
}

// UnmarshalBinary はSetCustomForceDataOutputDataをバイナリからパースします。
func (s *SetCustomForceDataOutputData) UnmarshalBinary(b []byte) error {
	s.ReportID = ReportID(b[0])
	s.EffectBlockIndex = b[1]
	s.DataOffset = binary.LittleEndian.Uint16(b[2:4])
	copy(s.Data[:], b[4:])
	return nil
}

// SetDownloadForceSampleOutputData はHIDレポート0x08（Set Download Force Sample）の構造体です。
// ダウンロードされた力サンプルデータを設定します。
type SetDownloadForceSampleOutputData struct {
	ReportID ReportID // レポートID（常に0x08）
	X        int8     // X軸方向の力サンプル値
	Y        int8     // Y軸方向の力サンプル値
}

// UnmarshalBinary はSetDownloadForceSampleOutputDataをバイナリからパースします。
func (s *SetDownloadForceSampleOutputData) UnmarshalBinary(b []byte) error {
	s.ReportID = ReportID(b[0])
	s.X = int8(b[1])
	s.Y = int8(b[2])
	return nil
}

// EffectOperationOutputData はHIDレポート0x0A（Effect Operation）の構造体です。
// エフェクトの再生操作（開始/停止）を実行します。
type EffectOperationOutputData struct {
	ReportID         ReportID        // レポートID（常に0x0A）
	EffectBlockIndex uint8           // エフェクトブロックインデックス（1-40）
	Operation        EffectOperation // 操作タイプ: 1=Start, 2=StartSolo, 3=Stop
	LoopCount        uint8           // ループ回数（0xFFは無限ループ）
}

// UnmarshalBinary はEffectOperationOutputDataをバイナリからパースします。
func (s *EffectOperationOutputData) UnmarshalBinary(b []byte) error {
	s.ReportID = ReportID(b[0])
	s.EffectBlockIndex = b[1]
	s.Operation = EffectOperation(b[2])
	s.LoopCount = b[3]
	return nil
}

// BlockFreeOutputData はHIDレポート0x0B（Block Free）の構造体です。
// エフェクトブロックを解放（削除）します。
type BlockFreeOutputData struct {
	ReportID         ReportID // レポートID（常に0x0B）
	EffectBlockIndex uint8    // 解放するエフェクトブロックインデックス（1-40）
}

// UnmarshalBinary はBlockFreeOutputDataをバイナリからパースします。
func (s *BlockFreeOutputData) UnmarshalBinary(b []byte) error {
	s.ReportID = ReportID(b[0])
	s.EffectBlockIndex = b[1]
	return nil
}

// DeviceControlOutputData はHIDレポート0x0C（Device Control）の構造体です。
// デバイスの制御コマンドを送信します。
type DeviceControlOutputData struct {
	ReportID ReportID    // レポートID（常に0x0C）
	Control  ControlType // 制御コマンド（定数参照）
}

// UnmarshalBinary はDeviceControlOutputDataをバイナリからパースします。
func (s *DeviceControlOutputData) UnmarshalBinary(b []byte) error {
	s.ReportID = ReportID(b[0])
	s.Control = ControlType(b[1])
	return nil
}

// DeviceGainOutputData はHIDレポート0x0D（Device Gain）の構造体です。
// デバイスのゲイン値を設定します。
type DeviceGainOutputData struct {
	ReportID ReportID // レポートID（常に0x0D）
	Gain     uint8    // ゲイン値（0-255、物理的には0-10000にマッピング可能）
}

// UnmarshalBinary はDeviceGainOutputDataをバイナリからパースします。
func (s *DeviceGainOutputData) UnmarshalBinary(b []byte) error {
	s.ReportID = ReportID(b[0])
	s.Gain = b[1]
	return nil
}

// SetCustomForceOutputData はHIDレポート0x0E（Set Custom Force）の構造体です。
// カスタム力エフェクトのパラメータを設定します。
type SetCustomForceOutputData struct {
	ReportID         ReportID // レポートID（常に0x0E）
	EffectBlockIndex uint8    // エフェクトブロックインデックス（1-40）
	SampleCount      uint8    // サンプル数
	SamplePeriod     uint16   // サンプリング期間（ms、0-32767）
}

// UnmarshalBinary はSetCustomForceOutputDataをバイナリからパースします。
func (s *SetCustomForceOutputData) UnmarshalBinary(b []byte) error {
	s.ReportID = ReportID(b[0])
	s.EffectBlockIndex = b[1]
	s.SampleCount = b[2]
	s.SamplePeriod = binary.LittleEndian.Uint16(b[3:5])
	return nil
}

// CreateNewEffectFeatureData はHIDレポート0x05（Create New Effect Feature Report）の構造体です。
// 新しいエフェクトブロックの作成をリクエストします。
type CreateNewEffectFeatureData struct {
	ReportID   ReportID   // レポートID（常に0x05）
	EffectType EffectType // エフェクトタイプ（定数参照）
	ByteCount  uint16     // データバイト数（0-511）
}

// UnmarshalBinary はCreateNewEffectFeatureDataをバイナリからパースします。
func (s *CreateNewEffectFeatureData) UnmarshalBinary(b []byte) error {
	s.ReportID = ReportID(b[0])
	s.EffectType = EffectType(b[1])
	s.ByteCount = binary.LittleEndian.Uint16(b[2:4])
	return nil
}

// PIDBlockLoadFeatureData はHIDレポート0x06（PID Block Load Feature Report）の構造体です。
// エフェクトブロックの読み込みステータスを取得します。
type PIDBlockLoadFeatureData struct {
	ReportID         ReportID // レポートID（常に0x06）
	EffectBlockIndex uint8    // エフェクトブロックインデックス（1-40）
	LoadStatus       uint8    // 読み込みステータス: 1=Success, 2=Full, 3=Error
	RamPoolAvailable uint16   // RAMプール利用可能サイズ（0または0xFFFF）
	b                []byte   // シリアライズ用の内部バッファ
}

// MarshalBinary はPIDBlockLoadFeatureDataをバイナリにシリアライズします。
func (s *PIDBlockLoadFeatureData) MarshalBinary() ([]byte, error) {
	b := s.b[:0]
	b = append(b, byte(s.ReportID))
	b = append(b, s.EffectBlockIndex)
	b = append(b, s.LoadStatus)
	b = binary.LittleEndian.AppendUint16(b, s.RamPoolAvailable)
	return b, nil
}

// PIDPoolFeatureData はHIDレポート0x07（PID Pool Feature Report）の構造体です。
// デバイスのPIDメモリプール情報を取得します。
type PIDPoolFeatureData struct {
	ReportID               ReportID // レポートID（常に0x07）
	RamPoolSize            uint16   // RAMプール総サイズ（バイト）
	MaxSimultaneousEffects uint8    // 同時再生可能エフェクト最大数（推測: 40）
	MemoryManagement       uint8    // メモリ管理モード: Bit0=DeviceManagedPool, Bit1=SharedParameterBlocks
	b                      []byte   // シリアライズ用の内部バッファ
}

// MarshalBinary はPIDPoolFeatureDataをバイナリにシリアライズします。
func (s *PIDPoolFeatureData) MarshalBinary() ([]byte, error) {
	b := s.b[:0]
	b = append(b, byte(s.ReportID))
	b = binary.LittleEndian.AppendUint16(b, s.RamPoolSize)
	b = append(b, s.MaxSimultaneousEffects)
	b = append(b, s.MemoryManagement)
	return b, nil
}

// EffectParams は力覚計算に必要な物理パラメータを提供します。
// 各エフェクトタイプ（Spring, Damper, Inertia, Friction）はこれらの値を使用して力を計算します。
type EffectParams struct {
	SpringMaxPosition         int32 // ばねエフェクトの最大位置範囲
	SpringPosition            int32 // 現在のばね位置
	DamperMaxVelocity         int32 // ダンパーエフェクトの最大速度
	DamperVelocity            int32 // 現在の速度
	InertiaMaxAcceleration    int32 // 慣性エフェクトの最大加速度
	InertiaAcceleration       int32 // 現在の加速度
	FrictionMaxPositionChange int32 // 摩擦エフェクトの最大位置変化量
	FrictionPositionChange    int32 // 現在の位置変化量
}
