package hint_utils

import "math/big"

func SECP_P() big.Int {
	secp_p, _ := new(big.Int).SetString("115792089237316195423570985008687907853269984665640564039457584007908834671663", 10)
	return *secp_p
}

func ALPHA() big.Int {
	alpha := big.NewInt(0)
	return *alpha
}
