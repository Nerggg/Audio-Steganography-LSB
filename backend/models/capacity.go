package models

type CapacityResult struct {
	// LSB method capacities
	OneLSB   int `json:"1_lsb"`
	TwoLSB   int `json:"2_lsb"`
	ThreeLSB int `json:"3_lsb"`
	FourLSB  int `json:"4_lsb"`
	// Parity coding capacity (1 bit per byte)
	Parity int `json:"parity"`
}
