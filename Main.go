package main

import (
	"encoding/json"
	    _ "embed"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

var (
	wordList      []string
	leetMap       map[string][]string
	symbols       = []rune{'!', '@', '#', '$', '%', '^', '&', '*', '(', ')', '-', '_', '+', '='}
	wordlistFile  = "wordlist.txt"
	leetmapFile   = "leetmap.json"
	configFile    = "config.json"
	cfg           Config
)
//go:embed icon.ico
var iconData []byte


type Config struct {
// Passphrase generator
PassWordCount int
PassSeparator string
PassRandomCase bool
PassNumbers bool
PassSymbols bool
PassLeet bool
PassLeetProb float64


// String randomizer
StringSeparator string
StringRandomCase bool
StringInvertCase bool
StringLeet bool
StringLeetProb float64


// Password generator (NEW)
PassGenLength int
PassGenLower bool
PassGenUpper bool
PassGenNumber bool
PassGenSymbol bool
PassGenInclude string
PassGenExclude string
PassGenBegin string
PassGenEnd string
PassGenExcludeSimilar bool
PassGenExcludeAmbiguous bool
PassGenNoDuplicate bool
	
}

func main() {
	rand.Seed(time.Now().UnixNano())
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
container.NewTabItem("Password Generator", passwordTab), // NEW TAB ENTRY
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
		ioutil.WriteFile("passphrases.txt", []byte(strings.Join(lines, "\n")), 0644)
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
		words[i] = wordList[rand.Intn(len(wordList))]
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
		if rand.Intn(2) == 0 {
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
	count := minCount + rand.Intn(3)
	runes := []rune(s)
	for i := 0; i < count; i++ {
		pos := rand.Intn(len(runes) + 1)
		num := rune('0' + rand.Intn(10))
		runes = append(runes[:pos], append([]rune{num}, runes[pos:]...)...)
	}
	return string(runes)
}

func insertRandomSymbols(s string, minCount int) string {
	count := minCount + rand.Intn(3)
	runes := []rune(s)
	for i := 0; i < count; i++ {
		pos := rand.Intn(len(runes) + 1)
		sym := symbols[rand.Intn(len(symbols))]
		runes = append(runes[:pos], append([]rune{sym}, runes[pos:]...)...)
	}
	return string(runes)
}

func applyLeet(s string, prob float64) string {
	runes := []rune(s)
	for i, r := range runes {
		strR := string(r)
		if vals, ok := leetMap[strR]; ok {
			if rand.Float64() < prob {
				runes[i] = []rune(vals[rand.Intn(len(vals))])[0]
			}
		}
	}
	return string(runes)
}

// ----------------- Load / Create Resources -----------------
func loadOrCreateWordList() {
	if _, err := os.Stat(wordlistFile); os.IsNotExist(err) {
		defaultWords := "apple banana cherry dog elephant frog grape hat igloo juice kiwi lemon mango nut orange pear queen rabbit snake tiger umbrella violin wolf xylophone yak zebra"
		ioutil.WriteFile(wordlistFile, []byte(defaultWords), 0644)
	}
	data, _ := ioutil.ReadFile(wordlistFile)
	wordList = strings.Fields(string(data))
}

func loadOrCreateLeetMap() {
	if _, err := os.Stat(leetmapFile); os.IsNotExist(err) {
		defaultLeet := map[string][]string{
			"a": {"4", "@"}, "e": {"3"}, "i": {"1", "!"}, "o": {"0"}, "s": {"5", "$"}, "t": {"7"},
		}
		b, _ := json.MarshalIndent(defaultLeet, "", "  ")
		ioutil.WriteFile(leetmapFile, b, 0644)
	}
	data, _ := ioutil.ReadFile(leetmapFile)
	json.Unmarshal(data, &leetMap)
}

func loadOrCreateConfig() {
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		cfg = Config{
			PassWordCount:    5,
			PassSeparator:    "_",
			PassRandomCase:   true,
			PassNumbers:      false,
			PassSymbols:      false,
			PassLeet:         true,
			PassLeetProb:     0.50,
			StringSeparator:  "_",
			StringRandomCase: true,
			StringInvertCase: false,
			StringLeet:       true,
			StringLeetProb:   0.50,
		}
		saveConfig()
	} else {
		data, _ := ioutil.ReadFile(configFile)
		json.Unmarshal(data, &cfg)
	}
}

func saveConfig() {
	b, _ := json.MarshalIndent(cfg, "", "  ")
	ioutil.WriteFile(configFile, b, 0644)
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

    // Character pools
    lowercase := []rune("abcdefghijklmnopqrstuvwxyz")
    uppercase := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ")
    numbers := []rune("0123456789")
    symbols := []rune("!@#$%^&*()-_=+[]{};:,.<>?/\\|")

    // Similar and ambiguous exclusions
    similar := []rune{'o', 'O', '0', 'i', 'I', 'l', '1'}
    ambiguous := []rune{'~', ';', ':', '.', '{', '}', '<', '>', '[', ']', '(', ')', '/', '\\', '\'', '`'}

    // Build pool based on user choices
    var pool []rune
    if cfg.PassGenLower {
        pool = append(pool, lowercase...)
    }
    if cfg.PassGenUpper {
        pool = append(pool, uppercase...)
    }
    if cfg.PassGenNumber {
        pool = append(pool, numbers...)
    }
    if cfg.PassGenSymbol {
        pool = append(pool, symbols...)
    }
    if cfg.PassGenInclude != "" {
        pool = append(pool, []rune(cfg.PassGenInclude)...)
    }

    // Apply exclusions
    excludeSet := make(map[rune]bool)
    for _, r := range cfg.PassGenExclude {
        excludeSet[r] = true
    }
    if cfg.PassGenExcludeSimilar {
        for _, r := range similar {
            excludeSet[r] = true
        }
    }
    if cfg.PassGenExcludeAmbiguous {
        for _, r := range ambiguous {
            excludeSet[r] = true
        }
    }

    // Filter final pool
    filtered := make([]rune, 0, len(pool))
    for _, r := range pool {
        if !excludeSet[r] {
            filtered = append(filtered, r)
        }
    }

    // Handle too-short pool when No Duplicate is enabled
    var warn string
    if cfg.PassGenNoDuplicate && len(filtered) < length {
        cfg.PassGenNoDuplicate = false
        saveConfig()
        warn = "No Duplicate disabled (not enough unique characters)."
    }

    // Build password
    var result []rune
    for len(result) < length {
        ch := filtered[rand.Intn(len(filtered))]
        if cfg.PassGenNoDuplicate && runeInSlice(ch, result) {
            continue
        }
        result = append(result, ch)
    }

    // Apply Begins With / Ends With
    if cfg.PassGenBegin == "Letter" {
        result[0] = randomLetter()
    } else if cfg.PassGenBegin == "Number" {
        result[0] = numbers[rand.Intn(len(numbers))]
    }

    if cfg.PassGenEnd == "Letter" {
        result[len(result)-1] = randomLetter()
    } else if cfg.PassGenEnd == "Number" {
        result[len(result)-1] = numbers[rand.Intn(len(numbers))]
    }

    return string(result), warn
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

func randomLetter() rune {
    if rand.Intn(2) == 0 {
        return rune('a' + rand.Intn(26))
    }
    return rune('A' + rand.Intn(26))
}
