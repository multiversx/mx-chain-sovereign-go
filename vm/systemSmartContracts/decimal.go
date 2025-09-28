package systemSmartContracts

import "github.com/nikolaydubina/fpdecimal"

var (
	DecimalZero, _     = fpdecimal.Parse([]byte("0"))
	DecimalOne, _      = fpdecimal.Parse([]byte("1"))
	DecimalTwo, _      = fpdecimal.Parse([]byte("2"))
	DecimalTen, _      = fpdecimal.Parse([]byte("10"))
	DecimalThousand, _ = fpdecimal.Parse([]byte("1000"))
)