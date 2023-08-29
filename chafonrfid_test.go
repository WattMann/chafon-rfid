package chafonrfid

import (
	"testing"
)

func TestChafonRFID(t *testing.T) {
	_, err := Inititialize("COM1")
	if err != nil {
		t.Error(err)
		return
	}

}
