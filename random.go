package main

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

func pickRune(pool []rune) (rune, error) {
	index, err := secureRandInt(len(pool))
	if err != nil {
		return 0, err
	}
	return pool[index], nil
}

func secureRandBool() (bool, error) {
	n, err := secureRandInt(2)
	return n == 0, err
}

func secureRandFloat64() (float64, error) {
	n, err := secureRandInt(1_000_000)
	if err != nil {
		return 0, err
	}
	return float64(n) / 1_000_000, nil
}

func secureRandInt(max int) (int, error) {
	if max <= 0 {
		return 0, fmt.Errorf("invalid random max: %d", max)
	}
	n, err := rand.Int(rand.Reader, big.NewInt(int64(max)))
	if err != nil {
		return 0, err
	}
	return int(n.Int64()), nil
}
