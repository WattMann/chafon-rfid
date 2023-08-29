package chafonrfid

import (
	"errors"
)

type Frame interface {
	CRC16() uint16
}

type CommandFrame struct {
	Len  uint8
	Adr  uint8
	Cmd  uint8
	Data []uint8
	LSB  uint8
	MSB  uint8
}

type ResponseFrame struct {
	Len    uint8
	Adr    uint8
	Cmd    uint8
	Status uint8
	Data   []uint8
	LSB    uint8
	MSB    uint8
}

func CreateCommand(adr, cmd uint8, data []uint8) (CommandFrame, error) {
	if len(data) > 251 {
		return CommandFrame{}, errors.New("data too big (>251)")
	}

	length := uint8(len(data)) + 4

	var lsb, msb uint8
	payload := make([]uint8, len(data)+3)
	payload = append(payload, length, adr, cmd)
	payload = append(payload, data...)
	crc := calculateCRC16(payload)

	lsb = uint8(crc)
	msb = uint8(crc >> 8)
	frame := CommandFrame{
		Len:  length,
		Adr:  adr,
		Cmd:  cmd,
		Data: data,
		LSB:  lsb,
		MSB:  msb,
	}

	return frame, nil
}

func (c *CommandFrame) Serialize() []byte {
	data := make([]byte, len(c.Data)+5)
	data = append(data, c.Len, c.Adr, c.Cmd)
	data = append(data, c.Data...)
	data = append(data, c.LSB, c.MSB)

	return data
}
