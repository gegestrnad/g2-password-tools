package main

import (
	"io"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func buildPassphraseTab(myWindow fyne.Window) fyne.CanvasObject {
	wordCountEntry := widget.NewEntry()
	wordCountEntry.SetText(intToStr(cfg.PassWordCount))
	wordCountEntry.OnChanged = func(s string) {
		cfg.PassWordCount = clampInt(strToInt(s, 5), minPassphraseWords, maxPassphraseWords, 5)
		saveConfig()
	}

	separatorEntry := widget.NewEntry()
	separatorEntry.SetText(cfg.PassSeparator)
	separatorEntry.OnChanged = func(s string) {
		cfg.PassSeparator = s
		saveConfig()
	}

	leetSlider := widget.NewSlider(0, 4)
	leetSlider.Step = 1
	leetSlider.Value = cfg.PassLeetProb * 4
	leetSlider.OnChanged = func(v float64) {
		cfg.PassLeetProb = clampFloat(v*0.25, 0, 1)
		saveConfig()
	}

	randomCaseCheck := widget.NewCheck("Random Case", func(b bool) {
		cfg.PassRandomCase = b
		saveConfig()
	})
	randomCaseCheck.SetChecked(cfg.PassRandomCase)

	numbersCheck := widget.NewCheck("Numbers", func(b bool) {
		cfg.PassNumbers = b
		saveConfig()
	})
	numbersCheck.SetChecked(cfg.PassNumbers)

	symbolsCheck := widget.NewCheck("Symbols", func(b bool) {
		cfg.PassSymbols = b
		saveConfig()
	})
	symbolsCheck.SetChecked(cfg.PassSymbols)

	leetCheck := widget.NewCheck("Leetspeak", func(b bool) {
		cfg.PassLeet = b
		saveConfig()
	})
	leetCheck.SetChecked(cfg.PassLeet)

	passphraseOutput := widget.NewMultiLineEntry()
	passphraseOutput.Wrapping = fyne.TextWrapWord

	warningLabel := widget.NewLabel("")
	warningLabel.Hide()

	generateBtn := widget.NewButton("Generate", func() {
		count := clampInt(strToInt(wordCountEntry.Text, 5), minPassphraseWords, maxPassphraseWords, 5)
		sep := separatorEntry.Text
		if sep == "" {
			sep = "_"
		}
		prob := clampFloat(leetSlider.Value*0.25, 0, 1)
		pass := generatePassphrase(count, sep, randomCaseCheck.Checked, numbersCheck.Checked, symbolsCheck.Checked, leetCheck.Checked, prob)
		passphraseOutput.SetText(pass)
		if count != strToInt(wordCountEntry.Text, 5) {
			warningLabel.SetText("Word count was limited to the supported range.")
			warningLabel.Show()
		} else {
			warningLabel.Hide()
		}
	})

	copyBtn := widget.NewButton("Copy Text", func() {
		fyne.CurrentApp().Clipboard().SetContent(passphraseOutput.Text)
		dialog.ShowInformation("Copied!", "Passphrase copied to clipboard", myWindow)
	})

	exportBtn := widget.NewButton("Export 20 Passphrases", func() {
		lines := make([]string, 0, 20)
		prob := clampFloat(leetSlider.Value*0.25, 0, 1)
		count := clampInt(strToInt(wordCountEntry.Text, 5), minPassphraseWords, maxPassphraseWords, 5)
		sep := separatorEntry.Text
		if sep == "" {
			sep = "_"
		}
		for i := 0; i < 20; i++ {
			lines = append(lines, generatePassphrase(count, sep, randomCaseCheck.Checked, numbersCheck.Checked, symbolsCheck.Checked, leetCheck.Checked, prob))
		}
		saveDialog := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
			if err != nil {
				dialog.ShowError(err, myWindow)
				return
			}
			if writer == nil {
				return
			}
			defer writer.Close()
			if _, err := io.WriteString(writer, strings.Join(lines, "\n")); err != nil {
				dialog.ShowError(err, myWindow)
				return
			}
			dialog.ShowInformation("Exported", "20 passphrases saved.", myWindow)
		}, myWindow)
		saveDialog.SetFileName("passphrases.txt")
		saveDialog.Show()
	})

	return container.NewVBox(
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
		warningLabel,
		container.NewHBox(generateBtn, copyBtn, exportBtn),
		passphraseOutput,
	)
}

func buildStringRandomizerTab(myWindow fyne.Window) fyne.CanvasObject {
	inputEntry := widget.NewMultiLineEntry()
	inputEntry.SetPlaceHolder("Enter text here...")

	separatorEntry := widget.NewEntry()
	separatorEntry.SetText(cfg.StringSeparator)
	separatorEntry.OnChanged = func(s string) {
		cfg.StringSeparator = s
		saveConfig()
	}

	randomCaseCheck := widget.NewCheck("Random Case", func(b bool) {
		cfg.StringRandomCase = b
		saveConfig()
	})
	randomCaseCheck.SetChecked(cfg.StringRandomCase)

	invertCaseCheck := widget.NewCheck("Invert Case", func(b bool) {
		cfg.StringInvertCase = b
		saveConfig()
	})
	invertCaseCheck.SetChecked(cfg.StringInvertCase)

	leetCheck := widget.NewCheck("Leetspeak", func(b bool) {
		cfg.StringLeet = b
		saveConfig()
	})
	leetCheck.SetChecked(cfg.StringLeet)

	leetSlider := widget.NewSlider(0, 4)
	leetSlider.Step = 1
	leetSlider.Value = cfg.StringLeetProb * 4
	leetSlider.OnChanged = func(v float64) {
		cfg.StringLeetProb = clampFloat(v*0.25, 0, 1)
		saveConfig()
	}

	stringOutput := widget.NewMultiLineEntry()
	stringOutput.Wrapping = fyne.TextWrapWord

	randomizeBtn := widget.NewButton("Randomize", func() {
		prob := clampFloat(leetSlider.Value*0.25, 0, 1)
		out := randomizeString(inputEntry.Text, separatorEntry.Text, randomCaseCheck.Checked, invertCaseCheck.Checked, leetCheck.Checked, prob)
		stringOutput.SetText(out)
	})

	copyBtn := widget.NewButton("Copy Text", func() {
		fyne.CurrentApp().Clipboard().SetContent(stringOutput.Text)
		dialog.ShowInformation("Copied!", "Randomized text copied to clipboard", myWindow)
	})

	return container.NewVBox(
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
}

func buildPasswordGeneratorTab(myWindow fyne.Window) fyne.CanvasObject {
	lengthEntry := widget.NewEntry()
	lengthEntry.SetText(intToStr(cfg.PassGenLength))
	lengthEntry.OnChanged = func(s string) {
		cfg.PassGenLength = clampInt(strToInt(s, 30), minPasswordLength, maxPasswordLength, 30)
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
	outputEntry.Wrapping = fyne.TextWrapWord

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

	copyBtn := widget.NewButton("Copy Text", func() {
		fyne.CurrentApp().Clipboard().SetContent(outputEntry.Text)
		dialog.ShowInformation("Copied!", "Password copied to clipboard", myWindow)
	})

	return container.NewVBox(
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
}
