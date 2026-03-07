// ****************************************************************************
//
//	 _     _          _
//	| |__ | | ___  __| |
//	| '_ \| |/ _ \/ _` |
//	| |_) | |  __/ (_| |
//	|_.__/|_|\___|\__,_|
//
// ****************************************************************************
// B L E D   -   Copyright © JPL 2026
// ****************************************************************************
package conf

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/gdamore/tcell/v2"
)

const (
	STATUS_MESSAGE_DURATION = 3
	APP_NAME                = "Bled"
	APP_STRING              = "Bled © jpl@ozf.fr 2026"
	APP_URL                 = "https://github.com/jplozf/bled"
	APP_FOLDER              = ".bled"
	ICON_MODIFIED           = "●"
	NEW_FILE_TEMPLATE       = "noname_"
	FILE_CONFIG             = "config.json"
	FILE_MRU                = "mru"
	DEFAULT_COLOR_ACCENT    = "#556B2F"
	COLOR_FOLDER            = tcell.ColorLightGreen
	COLOR_FILE              = tcell.ColorYellow
	COLOR_EXECUTABLE        = tcell.ColorLightYellow
	COLOR_SELECTED          = tcell.ColorRed
)

var Version string

var LogFile *os.File

type SConfigGeneral struct {
	Theme            string
	GitUser          string
	GitKey           string
	Workspace        string
	ShowHidden       bool
	ConfirmExit      bool
	CleanUpOnExit    bool
	FormatTime       string
	FormatDate       string
	ColorAccent      string
	InteractiveShell bool
	OutErrPrefix     bool
}

type SConfigPrivate struct {
	UISleepUpdate        int
	UIGITUpdate          int
	UIStatusTimeout      int
	ClearNullFilesOnExit bool
}

var ConfigGeneral SConfigGeneral

type Config struct {
	MenuBgColor       string `json:"menu_bg_color"`
	MenuSelectedColor string `json:"menu_selected_color"`
	MenuTextColor     string `json:"menu_text_color"`
	MenuDisabledColor string `json:"menu_disabled_color"`
}

// ****************************************************************************
// GetConfigPath()
// ****************************************************************************
func GetConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return FILE_CONFIG // Fallback
	}

	configDir := filepath.Join(home, APP_FOLDER)

	err = os.MkdirAll(configDir, 0755)
	if err != nil {
		return FILE_CONFIG // Fallback
	}

	return filepath.Join(configDir, FILE_CONFIG)
}

// ****************************************************************************
// LoadConfig()
// ****************************************************************************
func LoadConfig() Config {
	path := GetConfigPath()

	// Valeurs par défaut
	conf := Config{
		MenuBgColor:       "yellow",
		MenuSelectedColor: "red",
		MenuTextColor:     "black",
		MenuDisabledColor: "gray",
	}

	data, err := os.ReadFile(path)
	if err != nil {
		SaveConfig(path, conf) // Default config saved if not exists
		return conf
	}

	json.Unmarshal(data, &conf)
	return conf
}

// ****************************************************************************
// SaveConfig()
// ****************************************************************************
func SaveConfig(path string, conf Config) {
	data, _ := json.MarshalIndent(conf, "", "    ")
	os.WriteFile(path, data, 0644)
}

// ****************************************************************************
// GetColor()
// ****************************************************************************
func GetColor(name string) tcell.Color {
	return tcell.GetColor(name)
}
