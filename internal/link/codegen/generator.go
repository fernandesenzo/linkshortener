package codegen

import (
	"crypto/rand"
	"errors"
	"math/big"
)

const charset = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"

type RandomGenerator struct {
}

func New() *RandomGenerator {
	return &RandomGenerator{}
}

func (g *RandomGenerator) Generate(length int) (string, error) {
	if length < 1 {
		return "", errors.New("cannot generate a code with length shorter than one")
	}
	finalCode := make([]byte, length)
	limit := big.NewInt(36)
	for i := 0; i < length; i++ {
		num, _ := rand.Int(rand.Reader, limit)
		finalCode[i] = charset[num.Int64()]
	}
	return string(finalCode), nil
}
