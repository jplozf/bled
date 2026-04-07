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
	APP_NAME          = "Bled"
	APP_STRING        = "Bled © jpl@ozf.fr 2026"
	APP_URL           = "https://github.com/jplozf/bled"
	APP_FOLDER        = ".bled"
	ICON_MODIFIED     = "●"
	APP_ICON          = "⚶"
	NEW_FILE_TEMPLATE = "noname_"
	FILE_CONFIG       = "config.json"
	FILE_LOG          = "bled.log"
	FILE_MRU          = "mru"
)

type Config struct {
	MenuBgColor           string `json:"menu_bg_color"`
	MenuSelectedColor     string `json:"menu_selected_color"`
	MenuTextColor         string `json:"menu_text_color"`
	MenuDisabledColor     string `json:"menu_disabled_color"`
	ShowWelcomePopup      bool   `json:"show_welcome_popup"`
	ConfirmOnQuit         bool   `json:"confirm_on_quit"`
	StatusMessageDuration int    `json:"status_message_duration"`
	ShowHiddenFiles       bool   `json:"show_hidden_files"`
	ColorAccent           string `json:"color_accent"`
	Theme                 string `json:"theme"`
	GithubUser            string `json:"github_user"`
	GithubKey             string `json:"github_key"`
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
// GetLogPath()
// ****************************************************************************
func GetLogPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return FILE_LOG // Fallback
	}

	logDir := filepath.Join(home, APP_FOLDER)

	err = os.MkdirAll(logDir, 0755)
	if err != nil {
		return FILE_LOG // Fallback
	}

	return filepath.Join(logDir, FILE_LOG)
}

// ****************************************************************************
// LoadConfig()
// ****************************************************************************
func LoadConfig() Config {
	path := GetConfigPath()

	// Default configuration
	conf := Config{
		MenuBgColor:           "yellow",
		MenuSelectedColor:     "red",
		MenuTextColor:         "black",
		MenuDisabledColor:     "gray",
		ShowWelcomePopup:      true,
		ConfirmOnQuit:         true,
		StatusMessageDuration: 3,
		ShowHiddenFiles:       false,
		ColorAccent:           "#556B2F",
		Theme:                 "monokai",
		GithubUser:            "github_user",
		GithubKey:             "github_key",
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
