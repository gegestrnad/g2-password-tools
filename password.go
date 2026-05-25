package main

import "errors"

const (
	lowercaseChars      = "abcdefghijklmnopqrstuvwxyz"
	uppercaseChars      = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	numberChars         = "0123456789"
	passwordSymbolChars = "!@#$%^&*()-_=+[]{};:,.<>?/\\|"
)

var (
	similarChars   = []rune{'o', 'O', '0', 'i', 'I', 'l', '1'}
	ambiguousChars = []rune{'~', ';', ':', '.', '{', '}', '<', '>', '[', ']', '(', ')', '/', '\\', '\'', '`'}
)

type passwordClass struct {
	name  string
	runes []rune
	match func(rune) bool
}

func generatePassword() (string, string) {
	length := clampInt(cfg.PassGenLength, minPasswordLength, maxPasswordLength, 30)
	pool := filteredPasswordPool()
	if len(pool) == 0 {
		return "", "Select at least one character class or add included characters."
	}

	if cfg.PassGenNoDuplicate {
		pool = uniqueRunes(pool)
		if len(pool) < length {
			return "", "No Duplicate: not enough unique characters for the selected length."
		}
	}

	requiredClasses := requiredPasswordClasses()
	if len(requiredClasses) > length {
		return "", "Length is too short to include every selected character class."
	}
	for _, class := range requiredClasses {
		if len(class.runes) == 0 {
			return "", "A selected character class has no available characters after exclusions."
		}
	}

	beginPool, err := boundaryPool(cfg.PassGenBegin, pool)
	if err != nil {
		return "", err.Error()
	}
	endPool, err := boundaryPool(cfg.PassGenEnd, pool)
	if err != nil {
		return "", err.Error()
	}
	if length == 1 && cfg.PassGenBegin != "Any" && cfg.PassGenEnd != "Any" && cfg.PassGenBegin != cfg.PassGenEnd {
		return "", "Length 1 cannot satisfy different begins-with and ends-with rules."
	}

	result := make([]rune, length)
	occupied := make([]bool, length)
	if cfg.PassGenBegin != "Any" {
		if err := setRandomAt(result, occupied, 0, beginPool); err != nil {
			return "", "Unable to satisfy begins-with rule."
		}
	}
	if cfg.PassGenEnd != "Any" {
		if err := setRandomAt(result, occupied, length-1, endPool); err != nil {
			return "", "Unable to satisfy ends-with rule."
		}
	}

	for _, class := range requiredClasses {
		if resultContains(result, class.match) {
			continue
		}
		index, err := randomFreeIndex(occupied)
		if err != nil {
			return "", "Unable to include every selected character class with the selected boundaries."
		}
		if index < 0 {
			return "", "Unable to include every selected character class with the selected boundaries."
		}
		if err := setRandomAt(result, occupied, index, class.runes); err != nil {
			return "", "Unable to include every selected character class."
		}
	}

	for i := range result {
		if occupied[i] {
			continue
		}
		if err := setRandomAt(result, occupied, i, pool); err != nil {
			return "", "Unable to generate password with the selected settings."
		}
	}

	return string(result), ""
}

func setRandomAt(result []rune, occupied []bool, index int, pool []rune) error {
	candidates := pool
	if cfg.PassGenNoDuplicate {
		candidates = removeRunes(pool, result)
	}
	ch, err := pickRune(candidates)
	if err != nil {
		return err
	}
	result[index] = ch
	occupied[index] = true
	return nil
}

func filteredPasswordPool() []rune {
	pool := basePasswordPool()
	excludeSet := passwordExcludeSet()
	filtered := make([]rune, 0, len(pool))
	for _, r := range pool {
		if !excludeSet[r] {
			filtered = append(filtered, r)
		}
	}
	return filtered
}

func basePasswordPool() []rune {
	var pool []rune
	if cfg.PassGenLower {
		pool = append(pool, []rune(lowercaseChars)...)
	}
	if cfg.PassGenUpper {
		pool = append(pool, []rune(uppercaseChars)...)
	}
	if cfg.PassGenNumber {
		pool = append(pool, []rune(numberChars)...)
	}
	if cfg.PassGenSymbol {
		pool = append(pool, []rune(passwordSymbolChars)...)
	}
	pool = append(pool, []rune(cfg.PassGenInclude)...)
	return pool
}

func passwordExcludeSet() map[rune]bool {
	excludeSet := make(map[rune]bool)
	for _, r := range cfg.PassGenExclude {
		excludeSet[r] = true
	}
	if cfg.PassGenExcludeSimilar {
		for _, r := range similarChars {
			excludeSet[r] = true
		}
	}
	if cfg.PassGenExcludeAmbiguous {
		for _, r := range ambiguousChars {
			excludeSet[r] = true
		}
	}
	return excludeSet
}

func requiredPasswordClasses() []passwordClass {
	excludeSet := passwordExcludeSet()
	classes := make([]passwordClass, 0, 4)
	if cfg.PassGenLower {
		classes = append(classes, passwordClass{"lowercase", filterExcluded([]rune(lowercaseChars), excludeSet), isLower})
	}
	if cfg.PassGenUpper {
		classes = append(classes, passwordClass{"uppercase", filterExcluded([]rune(uppercaseChars), excludeSet), isUpper})
	}
	if cfg.PassGenNumber {
		classes = append(classes, passwordClass{"number", filterExcluded([]rune(numberChars), excludeSet), isNumber})
	}
	if cfg.PassGenSymbol {
		classes = append(classes, passwordClass{"symbol", filterExcluded([]rune(passwordSymbolChars), excludeSet), isPasswordSymbol})
	}
	return classes
}

func filterExcluded(pool []rune, excludeSet map[rune]bool) []rune {
	filtered := make([]rune, 0, len(pool))
	for _, r := range pool {
		if !excludeSet[r] {
			filtered = append(filtered, r)
		}
	}
	return filtered
}

func boundaryPool(mode string, pool []rune) ([]rune, error) {
	switch mode {
	case "", "Any":
		return pool, nil
	case "Letter":
		letters := filterRunes(pool, isLetter)
		if len(letters) == 0 {
			return nil, errors.New("Begins/ends-with Letter requires at least one available letter.")
		}
		return letters, nil
	case "Number":
		numbers := filterRunes(pool, isNumber)
		if len(numbers) == 0 {
			return nil, errors.New("Begins/ends-with Number requires at least one available number.")
		}
		return numbers, nil
	default:
		return pool, nil
	}
}

func validBoundaryMode(mode string) bool {
	return mode == "" || mode == "Any" || mode == "Letter" || mode == "Number"
}

func filterRunes(pool []rune, keep func(rune) bool) []rune {
	var filtered []rune
	for _, r := range pool {
		if keep(r) {
			filtered = append(filtered, r)
		}
	}
	return filtered
}

func isLower(r rune) bool {
	return 'a' <= r && r <= 'z'
}

func isUpper(r rune) bool {
	return 'A' <= r && r <= 'Z'
}

func isLetter(r rune) bool {
	return isLower(r) || isUpper(r)
}

func isNumber(r rune) bool {
	return '0' <= r && r <= '9'
}

func isPasswordSymbol(r rune) bool {
	return runeInSlice(r, []rune(passwordSymbolChars))
}

func uniqueRunes(pool []rune) []rune {
	seen := make(map[rune]bool, len(pool))
	unique := make([]rune, 0, len(pool))
	for _, r := range pool {
		if seen[r] {
			continue
		}
		seen[r] = true
		unique = append(unique, r)
	}
	return unique
}

func removeRunes(pool []rune, removes []rune) []rune {
	filtered := make([]rune, 0, len(pool))
	for _, r := range pool {
		if !runeInSlice(r, removes) {
			filtered = append(filtered, r)
		}
	}
	return filtered
}

func randomFreeIndex(occupied []bool) (int, error) {
	var free []int
	for i, used := range occupied {
		if !used {
			free = append(free, i)
		}
	}
	if len(free) == 0 {
		return -1, nil
	}
	index, err := secureRandInt(len(free))
	if err != nil {
		return -1, err
	}
	return free[index], nil
}

func resultContains(result []rune, match func(rune) bool) bool {
	for _, r := range result {
		if r != 0 && match(r) {
			return true
		}
	}
	return false
}

func runeInSlice(r rune, s []rune) bool {
	for _, x := range s {
		if x == r {
			return true
		}
	}
	return false
}
