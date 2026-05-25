package main

import (
	"fmt"
	"strings"
)

var symbols = []rune{'!', '@', '#', '$', '%', '^', '&', '*', '(', ')', '-', '_', '+', '='}

func generatePassphrase(count int, sep string, randCase, numbers, symbolsOn, leet bool, leetProb float64) string {
	if count <= 0 || len(wordList) == 0 {
		return ""
	}
	if count > maxPassphraseWords {
		count = maxPassphraseWords
	}
	if sep == "" {
		sep = "_"
	}
	words := make([]string, count)
	for i := 0; i < count; i++ {
		index, err := secureRandInt(len(wordList))
		if err != nil {
			return ""
		}
		words[i] = wordList[index]
	}
	pass := strings.Join(words, sep)
	if randCase {
		pass = applyRandomCase(pass)
	}
	if numbers {
		pass = insertRandomNumbers(pass, 2)
	}
	if symbolsOn {
		pass = insertRandomSymbols(pass, 2)
	}
	if leet {
		pass = applyLeet(pass, clampFloat(leetProb, 0, 1))
	}
	return pass
}

func randomizeString(s, sep string, randCase, invertCase, leet bool, leetProb float64) string {
	if sep == "" {
		sep = "_"
	}
	s = strings.ReplaceAll(s, " ", sep)
	if invertCase {
		s = invertCaseString(s)
	} else if randCase {
		s = applyRandomCase(s)
	}
	if leet {
		s = applyLeet(s, clampFloat(leetProb, 0, 1))
	}
	return s
}

func applyRandomCase(s string) string {
	runes := []rune(s)
	for i := range runes {
		flip, err := secureRandBool()
		if err != nil {
			return s
		}
		if flip {
			r := runes[i]
			if isLower(r) {
				runes[i] = r - 32
			} else if isUpper(r) {
				runes[i] = r + 32
			}
		}
	}
	return string(runes)
}

func invertCaseString(s string) string {
	runes := []rune(s)
	for i, r := range runes {
		if isLower(r) {
			runes[i] = r - 32
		} else if isUpper(r) {
			runes[i] = r + 32
		}
	}
	return string(runes)
}

func insertRandomNumbers(s string, minCount int) string {
	extra, err := secureRandInt(3)
	if err != nil {
		return s
	}
	count := minCount + extra
	runes := []rune(s)
	for i := 0; i < count; i++ {
		pos, err := secureRandInt(len(runes) + 1)
		if err != nil {
			return string(runes)
		}
		digit, err := secureRandInt(10)
		if err != nil {
			return string(runes)
		}
		num := rune('0' + digit)
		runes = append(runes[:pos], append([]rune{num}, runes[pos:]...)...)
	}
	return string(runes)
}

func insertRandomSymbols(s string, minCount int) string {
	extra, err := secureRandInt(3)
	if err != nil {
		return s
	}
	count := minCount + extra
	runes := []rune(s)
	for i := 0; i < count; i++ {
		pos, err := secureRandInt(len(runes) + 1)
		if err != nil {
			return string(runes)
		}
		index, err := secureRandInt(len(symbols))
		if err != nil {
			return string(runes)
		}
		sym := symbols[index]
		runes = append(runes[:pos], append([]rune{sym}, runes[pos:]...)...)
	}
	return string(runes)
}

func applyLeet(s string, prob float64) string {
	prob = clampFloat(prob, 0, 1)
	runes := []rune(s)
	for i, r := range runes {
		strR := string(r)
		if vals, ok := leetMap[strR]; ok && len(vals) > 0 {
			roll, err := secureRandFloat64()
			if err != nil {
				return s
			}
			if roll < prob {
				index, err := secureRandInt(len(vals))
				if err != nil {
					return s
				}
				replacement := []rune(vals[index])
				if len(replacement) > 0 {
					runes[i] = replacement[0]
				}
			}
		}
	}
	return string(runes)
}

func strToInt(s string, def int) int {
	var v int
	_, err := fmt.Sscanf(s, "%d", &v)
	if err != nil {
		return def
	}
	return v
}

func intToStr(v int) string {
	return fmt.Sprintf("%d", v)
}
