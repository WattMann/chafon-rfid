package chafonrfid

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"sync"

	"github.com/tarm/serial"
)

var (
	DEFAULT_BAUD_RATE_BPS = 56700
	REQUEST_TIMEOUT_MS    = 15
)

var (
	MODE_INVENTORY Mode = 0x1
	MODE_COMMAND   Mode = 0x2
)

type Mode uint8

type Handle struct {
	mutex      sync.Mutex
	mode       Mode
	port       *serial.Port
	shouldStop bool
}

func Inititialize(com string) (*Handle, error) {
	port_cfg := &serial.Config{Name: com, Baud: DEFAULT_BAUD_RATE_BPS}
	port, err := serial.OpenPort(port_cfg)
	if err != nil {
		return nil, nil
	}

	h := Handle{
		mutex: sync.Mutex{},
		port:  port,
	}

	return &h, nil
}

func (h *Handle) Terminate() {
	h.shouldStop = true
}

func (handle *Handle) StartInventoryMode() {
	if handle.mode == MODE_INVENTORY {
		return
	}

	handle.mutex.Lock()
	handle.mode = MODE_INVENTORY
	go handle.reciever()
}

func (handle *Handle) StopInventoryMode() {
	handle.mode = MODE_COMMAND
	handle.mutex.Unlock()
}

func (handle *Handle) reciever() {
	for handle.mode == MODE_INVENTORY {

	}
}

func (handle *Handle) sendcmd(frame CommandFrame) (ResponseFrame, error) {
	if handle.mode != MODE_COMMAND {
		return ResponseFrame{}, errors.New("reader is in invalid mode")
	}

	handle.mutex.Lock()
	defer func() {
		handle.mutex.Unlock()
	}()

	err := handle.writeFrame(frame)
	if err != nil {
		return ResponseFrame{}, fmt.Errorf("failed to write frame: %s", err.Error())
	}

	response, err := handle.readFrame()
	if err != nil {
		return ResponseFrame{}, fmt.Errorf("failed to read frame: %s", err.Error())
	}

	return response, nil

}

func (handle *Handle) readFrame() (ResponseFrame, error) {
	buffer := make([]uint8, 255)

	var payload bytes.Buffer
	decoder := gob.NewDecoder(&payload)

	n, err := handle.port.Read(buffer)
	buffer = buffer[:n]
	decoder.Decode(&buffer)
	if err != nil {
		return ResponseFrame{}, fmt.Errorf("failed to read response: %s", err.Error())
	}

	length, err := payload.ReadByte()
	if err != nil {
		return ResponseFrame{}, fmt.Errorf("failed to read response (len): %s", err.Error())
	}

	adr, err := payload.ReadByte()
	if err != nil {
		return ResponseFrame{}, fmt.Errorf("failed to read response (adr): %s", err.Error())
	}

	cmd, err := payload.ReadByte()
	if err != nil {
		return ResponseFrame{}, fmt.Errorf("failed to read response (cmd): %s", err.Error())
	}

	status, err := payload.ReadByte()
	if err != nil {
		return ResponseFrame{}, fmt.Errorf("failed to read response (status): %s", err.Error())
	}

	data := make([]byte, length-4)
	n, err = payload.Read(data)
	if err != nil || n != len(data) {
		return ResponseFrame{}, fmt.Errorf("failed to read response (data): %s", err.Error())
	}

	lsb, err := payload.ReadByte()
	if err != nil {
		return ResponseFrame{}, fmt.Errorf("failed to read response (lsb): %s", err.Error())
	}

	msb, err := payload.ReadByte()
	if err != nil {
		return ResponseFrame{}, fmt.Errorf("failed to read response (msb): %s", err.Error())
	}

	uidata := make([]uint8, len(data))
	for i, v := range data {
		uidata[i] = uint8(v)
	}

	response := ResponseFrame{
		Len:    uint8(length),
		Adr:    uint8(adr),
		Cmd:    uint8(cmd),
		Status: uint8(status),
		Data:   uidata,
		LSB:    uint8(lsb),
		MSB:    uint8(msb),
	}

	return response, nil
}

func (handle *Handle) writeFrame(frame CommandFrame) error {
	var payload bytes.Buffer
	encoder := gob.NewEncoder(&payload)

	serialized := frame.Serialize()
	if len(serialized) > 255 {
		return errors.New("frame too large (>251)")
	}

	err := encoder.Encode(serialized)
	if err != nil {
		return fmt.Errorf("failed to encode frame: %s", err.Error())
	}
	_, err = handle.port.Write(payload.Bytes())
	if err != nil {
		return fmt.Errorf(" ailed to write data to serial: %s", err.Error())
	}
	err = handle.port.Flush()
	if err != nil {
		return fmt.Errorf("failed to flush data to serial: %s", err.Error())
	}

	return nil
}
