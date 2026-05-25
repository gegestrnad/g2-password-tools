package main

import (
	_ "embed"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
)

//go:embed icon.ico
var iconData []byte

func main() {
	configureRuntimePaths()
	loadOrCreateWordList()
	loadOrCreateLeetMap()
	loadOrCreateConfig()

	myApp := app.New()
	myApp.SetIcon(fyne.NewStaticResource("icon", iconData))
	myWindow := myApp.NewWindow("G2 Password Tool")

	tabs := container.NewAppTabs(
		container.NewTabItem("Password Generator", buildPasswordGeneratorTab(myWindow)),
		container.NewTabItem("Passphrase Generator", buildPassphraseTab(myWindow)),
		container.NewTabItem("String Randomizer", buildStringRandomizerTab(myWindow)),
	)
	tabs.SetTabLocation(container.TabLocationTop)

	myWindow.SetContent(tabs)
	myWindow.Resize(fyne.NewSize(600, 400))
	myWindow.ShowAndRun()
}
