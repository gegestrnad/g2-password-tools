package main

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const (
	minPassphraseWords = 1
	maxPassphraseWords = 24
	minPasswordLength  = 1
	maxPasswordLength  = 128
	defaultWordList    = "apple banana cherry dog elephant frog grape hat igloo juice kiwi lemon mango nut orange pear queen rabbit snake tiger umbrella violin wolf xylophone yak zebra"
	appDataDirName     = "G2 Password Tool"
)

var (
	wordList     []string
	leetMap      map[string][]string
	wordlistFile = "wordlist.txt"
	leetmapFile  = "leetmap.json"
	configFile   = "config.json"
	runtimeDir   string
	cfg          Config
)

type Config struct {
	PassWordCount  int
	PassSeparator  string
	PassRandomCase bool
	PassNumbers    bool
	PassSymbols    bool
	PassLeet       bool
	PassLeetProb   float64

	StringSeparator  string
	StringRandomCase bool
	StringInvertCase bool
	StringLeet       bool
	StringLeetProb   float64

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

type runtimePaths struct {
	wordlist string
	leetmap  string
	config   string
}

func configureRuntimePaths() runtimePaths {
	base := runtimeDir
	if base == "" {
		base = defaultRuntimeDir()
	}
	if err := os.MkdirAll(base, 0755); err != nil {
		log.Printf("could not create app data directory %q, using current directory: %v", base, err)
		base = "."
	}

	paths := runtimePaths{
		wordlist: filepath.Join(base, "wordlist.txt"),
		leetmap:  filepath.Join(base, "leetmap.json"),
		config:   filepath.Join(base, "config.json"),
	}
	migrateRuntimeFile("wordlist.txt", paths.wordlist)
	migrateRuntimeFile("leetmap.json", paths.leetmap)
	migrateRuntimeFile("config.json", paths.config)

	wordlistFile = paths.wordlist
	leetmapFile = paths.leetmap
	configFile = paths.config
	return paths
}

func defaultRuntimeDir() string {
	if dir, err := os.UserConfigDir(); err == nil && dir != "" {
		return filepath.Join(dir, appDataDirName)
	}
	if dir, err := os.UserHomeDir(); err == nil && dir != "" {
		return filepath.Join(dir, "."+strings.ReplaceAll(strings.ToLower(appDataDirName), " ", "-"))
	}
	return "."
}

func migrateRuntimeFile(legacyName, target string) {
	if _, err := os.Stat(target); err == nil {
		return
	}
	if _, err := os.Stat(legacyName); err != nil {
		return
	}
	data, err := os.ReadFile(legacyName)
	if err != nil {
		log.Printf("could not read legacy %s: %v", legacyName, err)
		return
	}
	if err := os.WriteFile(target, data, 0644); err != nil {
		log.Printf("could not migrate %s to %s: %v", legacyName, target, err)
	}
}

func loadOrCreateWordList() {
	if err := ensureWordListFile(); err != nil {
		log.Printf("could not prepare wordlist: %v", err)
		wordList = strings.Fields(defaultWordList)
		return
	}
	data, err := os.ReadFile(wordlistFile)
	if err != nil {
		log.Printf("could not read wordlist: %v", err)
		wordList = strings.Fields(defaultWordList)
		return
	}
	wordList = strings.Fields(string(data))
	if len(wordList) == 0 {
		log.Printf("wordlist was empty, restoring defaults")
		if err := writeDefaultWordList(); err != nil {
			log.Printf("could not restore wordlist: %v", err)
		}
		wordList = strings.Fields(defaultWordList)
	}
}

func ensureWordListFile() error {
	if _, err := os.Stat(wordlistFile); errors.Is(err, os.ErrNotExist) {
		return writeDefaultWordList()
	} else if err != nil {
		return err
	}
	return nil
}

func writeDefaultWordList() error {
	return os.WriteFile(wordlistFile, []byte(defaultWordList), 0644)
}

func loadOrCreateLeetMap() {
	if err := ensureLeetMapFile(); err != nil {
		log.Printf("could not prepare leet map: %v", err)
		leetMap = defaultLeetMap()
		return
	}
	data, err := os.ReadFile(leetmapFile)
	if err != nil {
		log.Printf("could not read leet map: %v", err)
		leetMap = defaultLeetMap()
		return
	}
	if err := json.Unmarshal(data, &leetMap); err != nil || len(leetMap) == 0 {
		if err != nil {
			log.Printf("could not parse leet map, restoring defaults: %v", err)
		} else {
			log.Printf("leet map was empty, restoring defaults")
		}
		leetMap = defaultLeetMap()
		if err := writeDefaultLeetMap(); err != nil {
			log.Printf("could not restore leet map: %v", err)
		}
	}
}

func ensureLeetMapFile() error {
	if _, err := os.Stat(leetmapFile); errors.Is(err, os.ErrNotExist) {
		return writeDefaultLeetMap()
	} else if err != nil {
		return err
	}
	return nil
}

func writeDefaultLeetMap() error {
	b, err := json.MarshalIndent(defaultLeetMap(), "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(leetmapFile, b, 0644)
}

func defaultLeetMap() map[string][]string {
	return map[string][]string{
		"a": {"4", "@"},
		"e": {"3"},
		"i": {"1", "!"},
		"o": {"0"},
		"s": {"5", "$"},
		"t": {"7"},
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
	b, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		log.Printf("could not encode config: %v", err)
		return
	}
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
	c.PassWordCount = clampInt(c.PassWordCount, minPassphraseWords, maxPassphraseWords, 5)
	if c.PassSeparator == "" {
		c.PassSeparator = "_"
	}
	c.PassLeetProb = clampFloat(c.PassLeetProb, 0, 1)

	if c.StringSeparator == "" {
		c.StringSeparator = "_"
	}
	c.StringLeetProb = clampFloat(c.StringLeetProb, 0, 1)

	c.PassGenLength = clampInt(c.PassGenLength, minPasswordLength, maxPasswordLength, 30)
	if !validBoundaryMode(c.PassGenBegin) {
		c.PassGenBegin = "Any"
	}
	if !validBoundaryMode(c.PassGenEnd) {
		c.PassGenEnd = "Any"
	}
	return c
}

func clampInt(v, min, max, def int) int {
	if v <= 0 {
		return def
	}
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

func clampFloat(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}
