package board

import (
	"machine"
)

const (
	FLASH_TARGET_OFFSET = 0 // 書き込み開始アドレス
)

// Flashに1ページ書き込み
func WriteFlashBlock(data []byte) error {
	err := machine.Flash.EraseBlocks(0, 1)
	if err != nil {
		return err
	}
	if _, err := machine.Flash.WriteAt(data, FLASH_TARGET_OFFSET); err != nil {
		return err
	}
	return nil
}

// Flashから読み出し
func ReadFlashBlock() ([]byte, error) {
	buff := make([]byte, machine.Flash.WriteBlockSize())
	n, err := machine.Flash.ReadAt(buff, FLASH_TARGET_OFFSET)
	if err != nil {
		return nil, err
	}
	return buff[:n], nil
}
