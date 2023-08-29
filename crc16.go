package chafonrfid

var (
	POLYNOMIAL        uint16 = 0x8408
	INITIAL_CRC_VALUE uint16 = 0xFFFF
)

func calculateCRC16(data []uint8) uint16 {
	var crcValue uint16 = INITIAL_CRC_VALUE
	var i, j int
	for i = 0; i < len(data); i++ {
		crcValue = crcValue ^ uint16(data[i])
		for j = 0; j < 8; j++ {
			if crcValue&0x0001 > 0 {
				crcValue = (crcValue >> 1) ^ POLYNOMIAL
			} else {
				crcValue = (crcValue >> 1)
			}
		}
	}

	return crcValue
}
