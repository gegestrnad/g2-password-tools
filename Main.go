package main

import (
	"crypto/rand"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/big"
	"os"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

var (
	wordList     []string
	leetMap      map[string][]string
	symbols      = []rune{'!', '@', '#', '$', '%', '^', '&', '*', '(', ')', '-', '_', '+', '='}
	wordlistFile = "wordlist.txt"
	leetmapFile  = "leetmap.json"
	configFile   = "config.json"
	cfg          Config
)

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

//go:embed icon.ico
var iconData []byte

type Config struct {
	// Passphrase generator
	PassWordCount  int
	PassSeparator  string
	PassRandomCase bool
	PassNumbers    bool
	PassSymbols    bool
	PassLeet       bool
	PassLeetProb   float64

	// String randomizer
	StringSeparator  string
	StringRandomCase bool
	StringInvertCase bool
	StringLeet       bool
	StringLeetProb   float64

	// Password generator
	PassGenLength           int
	PassGenLower            bool
	PassGenUpper            bool
	PassGenNumber           bool
	PassGenSymbol           bool
	PassGenInclude          string
	PassGenExclude          string
	PassGenBegin            string
	PassGenEnd              string
	PassGenExcludeSimilar   bool
	PassGenExcludeAmbiguous bool
	PassGenNoDuplicate      bool
}

func main() {
	loadOrCreateWordList()
	loadOrCreateLeetMap()
	loadOrCreateConfig()

	myApp := app.New()
	myApp.SetIcon(fyne.NewStaticResource("icon", iconData))
	myWindow := myApp.NewWindow("G2 Password Tool")

	passwordTab := buildPasswordGeneratorTab(myWindow) // NEW TAB
	passphraseTab := buildPassphraseTab(myWindow)
	stringTab := buildStringRandomizerTab(myWindow)

	tabs := container.NewAppTabs(
		container.NewTabItem("Password Generator", passwordTab),
		container.NewTabItem("Passphrase Generator", passphraseTab),
		container.NewTabItem("String Randomizer", stringTab),
	)
	tabs.SetTabLocation(container.TabLocationTop)

	myWindow.SetContent(tabs)
	myWindow.Resize(fyne.NewSize(600, 400))
	myWindow.ShowAndRun()
}

// ----------------- Passphrase Generator -----------------
func buildPassphraseTab(myWindow fyne.Window) fyne.CanvasObject {
	wordCountEntry := widget.NewEntry()
	wordCountEntry.SetText(intToStr(cfg.PassWordCount))

	separatorEntry := widget.NewEntry()
	separatorEntry.SetText(cfg.PassSeparator)

	leetSlider := widget.NewSlider(0, 4)
	leetSlider.Value = cfg.PassLeetProb * 4

	randomCaseCheck := widget.NewCheck("Random Case", func(bool) {})
	randomCaseCheck.SetChecked(cfg.PassRandomCase)

	numbersCheck := widget.NewCheck("Numbers", func(bool) {})
	numbersCheck.SetChecked(cfg.PassNumbers)

	symbolsCheck := widget.NewCheck("Symbols", func(bool) {})
	symbolsCheck.SetChecked(cfg.PassSymbols)

	leetCheck := widget.NewCheck("Leetspeak", func(bool) {})
	leetCheck.SetChecked(cfg.PassLeet)

	passphraseOutput := widget.NewMultiLineEntry()
	passphraseOutput.Disable()

	generateBtn := widget.NewButton("Generate", func() {
		count := strToInt(wordCountEntry.Text, 4)
		sep := separatorEntry.Text
		if sep == "" {
			sep = "_"
		}
		prob := leetSlider.Value * 0.25
		pass := generatePassphrase(count, sep, randomCaseCheck.Checked, numbersCheck.Checked, symbolsCheck.Checked, leetCheck.Checked, prob)
		passphraseOutput.SetText(pass)
	})

	copyBtn := widget.NewButton("Copy Text", func() {
		myApp := fyne.CurrentApp()
		myApp.Clipboard().SetContent(passphraseOutput.Text)
		dialog.ShowInformation("Copied!", "Passphrase copied to clipboard", myWindow)
	})

	exportBtn := widget.NewButton("Export 20 Passphrases", func() {
		var lines []string
		prob := leetSlider.Value * 0.25
		for i := 0; i < 20; i++ {
			lines = append(lines, generatePassphrase(strToInt(wordCountEntry.Text, 4), separatorEntry.Text, randomCaseCheck.Checked, numbersCheck.Checked, symbolsCheck.Checked, leetCheck.Checked, prob))
		}
		if err := os.WriteFile("passphrases.txt", []byte(strings.Join(lines, "\n")), 0644); err != nil {
			dialog.ShowError(err, myWindow)
			return
		}
		dialog.ShowInformation("Exported", "20 passphrases saved to passphrases.txt", myWindow)
	})

	form := container.NewVBox(
		widget.NewLabel("Number of Words:"),
		wordCountEntry,
		widget.NewLabel("Separator:"),
		separatorEntry,
		randomCaseCheck,
		numbersCheck,
		symbolsCheck,
		leetCheck,
		widget.NewLabel("Leet Probability:"),
		leetSlider,
		container.NewHBox(generateBtn, copyBtn, exportBtn),
		passphraseOutput,
	)

	return form
}

// ----------------- String Randomizer -----------------
func buildStringRandomizerTab(myWindow fyne.Window) fyne.CanvasObject {
	inputEntry := widget.NewMultiLineEntry()
	inputEntry.SetPlaceHolder("Enter text here...")

	separatorEntry := widget.NewEntry()
	separatorEntry.SetText(cfg.StringSeparator)

	randomCaseCheck := widget.NewCheck("Random Case", func(bool) {})
	randomCaseCheck.SetChecked(cfg.StringRandomCase)

	invertCaseCheck := widget.NewCheck("Invert Case", func(bool) {})
	invertCaseCheck.SetChecked(cfg.StringInvertCase)

	leetCheck := widget.NewCheck("Leetspeak", func(bool) {})
	leetCheck.SetChecked(cfg.StringLeet)

	leetSlider := widget.NewSlider(0, 4)
	leetSlider.Value = cfg.StringLeetProb * 4

	stringOutput := widget.NewMultiLineEntry()
	stringOutput.Disable()

	randomizeBtn := widget.NewButton("Randomize", func() {
		prob := leetSlider.Value * 0.25
		out := randomizeString(inputEntry.Text, separatorEntry.Text, randomCaseCheck.Checked, invertCaseCheck.Checked, leetCheck.Checked, prob)
		stringOutput.SetText(out)
	})

	copyBtn := widget.NewButton("Copy Text", func() {
		myApp := fyne.CurrentApp()
		myApp.Clipboard().SetContent(stringOutput.Text)
		dialog.ShowInformation("Copied!", "Randomized text copied to clipboard", myWindow)
	})

	form := container.NewVBox(
		widget.NewLabel("Input Text:"),
		inputEntry,
		widget.NewLabel("Replace spaces with:"),
		separatorEntry,
		randomCaseCheck,
		invertCaseCheck,
		leetCheck,
		widget.NewLabel("Leet Probability:"),
		leetSlider,
		container.NewHBox(randomizeBtn, copyBtn),
		stringOutput,
	)

	return form
}

// ----------------- Utilities -----------------
func generatePassphrase(count int, sep string, randCase, numbers, symbolsOn, leet bool, leetProb float64) string {
	if count <= 0 || len(wordList) == 0 {
		return ""
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
		pass = applyLeet(pass, leetProb)
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
		s = applyLeet(s, leetProb)
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
			if 'a' <= r && r <= 'z' {
				runes[i] = r - 32
			} else if 'A' <= r && r <= 'Z' {
				runes[i] = r + 32
			}
		}
	}
	return string(runes)
}

func invertCaseString(s string) string {
	runes := []rune(s)
	for i, r := range runes {
		if 'a' <= r && r <= 'z' {
			runes[i] = r - 32
		} else if 'A' <= r && r <= 'Z' {
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
	runes := []rune(s)
	for i, r := range runes {
		strR := string(r)
		if vals, ok := leetMap[strR]; ok {
			roll, err := secureRandFloat64()
			if err != nil {
				return s
			}
			if roll < prob {
				index, err := secureRandInt(len(vals))
				if err != nil {
					return s
				}
				runes[i] = []rune(vals[index])[0]
			}
		}
	}
	return string(runes)
}

// ----------------- Load / Create Resources -----------------
func loadOrCreateWordList() {
	if _, err := os.Stat(wordlistFile); os.IsNotExist(err) {
		defaultWords := "apple banana cherry dog elephant frog grape hat igloo juice kiwi lemon mango nut orange pear queen rabbit snake tiger umbrella violin wolf xylophone yak zebra"
		if err := os.WriteFile(wordlistFile, []byte(defaultWords), 0644); err != nil {
			log.Printf("could not create wordlist: %v", err)
		}
	}
	data, err := os.ReadFile(wordlistFile)
	if err != nil {
		log.Printf("could not read wordlist: %v", err)
		return
	}
	wordList = strings.Fields(string(data))
}

func loadOrCreateLeetMap() {
	if _, err := os.Stat(leetmapFile); os.IsNotExist(err) {
		defaultLeet := map[string][]string{
			"a": {"4", "@"}, "e": {"3"}, "i": {"1", "!"}, "o": {"0"}, "s": {"5", "$"}, "t": {"7"},
		}
		b, _ := json.MarshalIndent(defaultLeet, "", "  ")
		if err := os.WriteFile(leetmapFile, b, 0644); err != nil {
			log.Printf("could not create leet map: %v", err)
		}
	}
	data, err := os.ReadFile(leetmapFile)
	if err != nil {
		log.Printf("could not read leet map: %v", err)
		leetMap = map[string][]string{}
		return
	}
	if err := json.Unmarshal(data, &leetMap); err != nil {
		log.Printf("could not parse leet map: %v", err)
		leetMap = map[string][]string{}
	}
}

func loadOrCreateConfig() {
	cfg = defaultConfig()

	data, err := os.ReadFile(configFile)
	if errors.Is(err, os.ErrNotExist) {
		saveConfig()
		return
	}
	if err != nil {
		log.Printf("could not read config, using defaults: %v", err)
		return
	}

	next := defaultConfig()
	if err := json.Unmarshal(data, &next); err != nil {
		log.Printf("could not parse config, restoring defaults: %v", err)
		cfg = defaultConfig()
		saveConfig()
		return
	}
	cfg = normalizeConfig(next)
	saveConfig()
}

func saveConfig() {
	b, _ := json.MarshalIndent(cfg, "", "  ")
	if err := os.WriteFile(configFile, b, 0644); err != nil {
		log.Printf("could not save config: %v", err)
	}
}

func defaultConfig() Config {
	return Config{
		PassWordCount:  5,
		PassSeparator:  "_",
		PassRandomCase: true,
		PassNumbers:    false,
		PassSymbols:    false,
		PassLeet:       true,
		PassLeetProb:   0.50,

		StringSeparator:  "_",
		StringRandomCase: true,
		StringInvertCase: false,
		StringLeet:       true,
		StringLeetProb:   0.50,

		PassGenLength:           30,
		PassGenLower:            true,
		PassGenUpper:            true,
		PassGenNumber:           true,
		PassGenSymbol:           true,
		PassGenInclude:          "",
		PassGenExclude:          "",
		PassGenBegin:            "Any",
		PassGenEnd:              "Any",
		PassGenExcludeSimilar:   false,
		PassGenExcludeAmbiguous: false,
		PassGenNoDuplicate:      true,
	}
}

func normalizeConfig(c Config) Config {
	if c.PassWordCount <= 0 {
		c.PassWordCount = 5
	}
	if c.PassSeparator == "" {
		c.PassSeparator = "_"
	}
	if c.StringSeparator == "" {
		c.StringSeparator = "_"
	}
	if c.PassGenLength <= 0 {
		c.PassGenLength = 30
	}
	if !validBoundaryMode(c.PassGenBegin) {
		c.PassGenBegin = "Any"
	}
	if !validBoundaryMode(c.PassGenEnd) {
		c.PassGenEnd = "Any"
	}
	return c
}

// ----------------- Helpers -----------------
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

// ----------------- Password Generator -----------------
// This function builds the Password Generator tab UI.
// It matches the visual style of the other tabs and automatically saves settings when changed.

func buildPasswordGeneratorTab(myWindow fyne.Window) fyne.CanvasObject {
	// Create UI entries and checkboxes with defaults loaded from cfg
	lengthEntry := widget.NewEntry()
	lengthEntry.SetText(intToStr(cfg.PassGenLength))
	lengthEntry.OnChanged = func(s string) {
		cfg.PassGenLength = strToInt(s, 30)
		saveConfig()
	}

	lowercaseCheck := widget.NewCheck("Lowercase (a-z)", func(b bool) {
		cfg.PassGenLower = b
		saveConfig()
	})
	lowercaseCheck.SetChecked(cfg.PassGenLower)

	uppercaseCheck := widget.NewCheck("Uppercase (A-Z)", func(b bool) {
		cfg.PassGenUpper = b
		saveConfig()
	})
	uppercaseCheck.SetChecked(cfg.PassGenUpper)

	numbersCheck := widget.NewCheck("Numbers (0-9)", func(b bool) {
		cfg.PassGenNumber = b
		saveConfig()
	})
	numbersCheck.SetChecked(cfg.PassGenNumber)

	symbolsCheck := widget.NewCheck("Symbols (!@#...)", func(b bool) {
		cfg.PassGenSymbol = b
		saveConfig()
	})
	symbolsCheck.SetChecked(cfg.PassGenSymbol)

	includeEntry := widget.NewEntry()
	includeEntry.SetText(cfg.PassGenInclude)
	includeEntry.OnChanged = func(s string) {
		cfg.PassGenInclude = s
		saveConfig()
	}

	excludeEntry := widget.NewEntry()
	excludeEntry.SetText(cfg.PassGenExclude)
	excludeEntry.OnChanged = func(s string) {
		cfg.PassGenExclude = s
		saveConfig()
	}

	beginsSelect := widget.NewSelect([]string{"Any", "Letter", "Number"}, func(s string) {
		cfg.PassGenBegin = s
		saveConfig()
	})
	beginsSelect.SetSelected(cfg.PassGenBegin)

	endsSelect := widget.NewSelect([]string{"Any", "Letter", "Number"}, func(s string) {
		cfg.PassGenEnd = s
		saveConfig()
	})
	endsSelect.SetSelected(cfg.PassGenEnd)

	excludeSimilarCheck := widget.NewCheck("Exclude Similar (o,0,i,l,1)", func(b bool) {
		cfg.PassGenExcludeSimilar = b
		saveConfig()
	})
	excludeSimilarCheck.SetChecked(cfg.PassGenExcludeSimilar)

	excludeAmbiguousCheck := widget.NewCheck("Exclude Ambiguous (~,;:.{}<>)", func(b bool) {
		cfg.PassGenExcludeAmbiguous = b
		saveConfig()
	})
	excludeAmbiguousCheck.SetChecked(cfg.PassGenExcludeAmbiguous)

	noDuplicateCheck := widget.NewCheck("No Duplicate Characters", func(b bool) {
		cfg.PassGenNoDuplicate = b
		saveConfig()
	})
	noDuplicateCheck.SetChecked(cfg.PassGenNoDuplicate)

	warningLabel := widget.NewLabel("")
	warningLabel.Hide()

	outputEntry := widget.NewMultiLineEntry()
	outputEntry.Disable()

	// Generate button logic
	generateBtn := widget.NewButton("Generate", func() {
		pass, warn := generatePassword()
		outputEntry.SetText(pass)
		if warn != "" {
			warningLabel.SetText(warn)
			warningLabel.Show()
		} else {
			warningLabel.Hide()
		}
	})

	// Copy button logic
	copyBtn := widget.NewButton("Copy Text", func() {
		fyne.CurrentApp().Clipboard().SetContent(outputEntry.Text)
		dialog.ShowInformation("Copied!", "Password copied to clipboard", myWindow)
	})

	form := container.NewVBox(
		widget.NewLabel("Password Length:"),
		lengthEntry,
		lowercaseCheck,
		uppercaseCheck,
		numbersCheck,
		symbolsCheck,
		widget.NewLabel("Characters to Include:"),
		includeEntry,
		widget.NewLabel("Characters to Exclude:"),
		excludeEntry,
		widget.NewLabel("Begins With:"),
		beginsSelect,
		widget.NewLabel("Ends With:"),
		endsSelect,
		excludeSimilarCheck,
		excludeAmbiguousCheck,
		noDuplicateCheck,
		warningLabel,
		container.NewHBox(generateBtn, copyBtn),
		outputEntry,
	)

	return form
}

// ----------------- Password Generator Logic -----------------
// Builds the password string according to all user settings in cfg.
func generatePassword() (string, string) {
	length := cfg.PassGenLength
	if length <= 0 {
		length = 30
	}

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

	result := make([]rune, 0, length)
	available := append([]rune(nil), pool...)
	for len(result) < length {
		ch, err := pickRune(available)
		if err != nil {
			return "", "Unable to generate password with the selected settings."
		}
		result = append(result, ch)
		if cfg.PassGenNoDuplicate {
			available = removeRune(available, ch)
		}
	}

	if cfg.PassGenBegin != "Any" {
		ch, err := pickRune(beginPool)
		if err != nil {
			return "", "Unable to satisfy begins-with rule."
		}
		if cfg.PassGenNoDuplicate {
			result, err = replaceUniqueAt(result, 0, ch, beginPool)
			if err != nil {
				return "", err.Error()
			}
		} else {
			result[0] = ch
		}
	}

	if cfg.PassGenEnd != "Any" {
		ch, err := pickRune(endPool)
		if err != nil {
			return "", "Unable to satisfy ends-with rule."
		}
		index := len(result) - 1
		if cfg.PassGenNoDuplicate {
			next, err := replaceUniqueAt(result, index, ch, endPool)
			if err != nil {
				return "", err.Error()
			}
			result = next
		} else {
			result[index] = ch
		}
	}

	return string(result), ""
}

func filteredPasswordPool() []rune {
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

func replaceUniqueAt(result []rune, index int, preferred rune, pool []rune) ([]rune, error) {
	if result[index] == preferred || !runeInSlice(preferred, result) {
		result[index] = preferred
		return result, nil
	}

	candidates := uniqueRunes(pool)
	for _, candidate := range candidates {
		if !runeInSlice(candidate, result) {
			result[index] = candidate
			return result, nil
		}
	}
	return nil, errors.New("No Duplicate cannot satisfy the selected begins/ends-with rule.")
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

func isLetter(r rune) bool {
	return ('a' <= r && r <= 'z') || ('A' <= r && r <= 'Z')
}

func isNumber(r rune) bool {
	return '0' <= r && r <= '9'
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

func removeRune(pool []rune, remove rune) []rune {
	for i, r := range pool {
		if r == remove {
			return append(pool[:i], pool[i+1:]...)
		}
	}
	return pool
}

// ----------------- Helper Functions -----------------
func runeInSlice(r rune, s []rune) bool {
	for _, x := range s {
		if x == r {
			return true
		}
	}
	return false
}

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
