package motor

import "github.com/nobonobo/q16"

// State モーターのステータス（すべてQ16.16固定小数点）
type State struct {
	Angle       q16.Fixed // 角度 [rad]
	Velocity    q16.Fixed // 角速度 [rad/s]
	Current     q16.Fixed // 電流値 [mA]
	Temperature q16.Fixed // 温度 [°C]
}
