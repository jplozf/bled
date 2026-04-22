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
	FILE_MACROS       = "macros"
	FILE_RECENTS      = "recents"
	FOLDER_TEMPLATES  = "templates"
	FILE_NEW_TEMPLATE = "noname"
)

type Config struct {
	MenuBgColor       string `json:"menu_bg_color"`
	MenuSelectedColor string `json:"menu_selected_color"`
	MenuTextColor     string `json:"menu_text_color"`
	MenuDisabledColor string `json:"menu_disabled_color"`
	ShowWelcomePopup  bool   `json:"show_welcome_popup"`
	ConfirmOnQuit     bool   `json:"confirm_on_quit"`
	ShowHiddenFiles   bool   `json:"show_hidden_files"`
	Theme             string `json:"theme"`
	GithubUser        string `json:"github_user"`
	GithubKey         string `json:"github_key"`
	GithubEmail       string `json:"github_email"`
	MaxRecentFiles    int    `json:"max_recent_files"`
}

var LogFile *os.File

// ****************************************************************************
// GetConfigDir()
// ****************************************************************************
func GetConfigDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return APP_FOLDER
	}

	configDir := filepath.Join(home, APP_FOLDER)
	err = os.MkdirAll(configDir, 0755)
	if err != nil {
		return APP_FOLDER // Fallback
	}

	return configDir
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
// GetMacrosPath()
// ****************************************************************************
func GetMacrosPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return FILE_MACROS // Fallback
	}

	macrosDir := filepath.Join(home, APP_FOLDER)

	err = os.MkdirAll(macrosDir, 0755)
	if err != nil {
		return FILE_MACROS // Fallback
	}

	return filepath.Join(macrosDir, FILE_MACROS)
}

// ****************************************************************************
// GetRecentPath()
// ****************************************************************************
func GetRecentPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return FILE_RECENTS // Fallback
	}

	recentDir := filepath.Join(home, APP_FOLDER)

	err = os.MkdirAll(recentDir, 0755)
	if err != nil {
		return FILE_RECENTS // Fallback
	}

	return filepath.Join(recentDir, FILE_RECENTS)
}

// ****************************************************************************
// LoadConfig()
// ****************************************************************************
func LoadConfig() Config {
	path := GetConfigPath()

	// Default configuration
	conf := Config{
		MenuBgColor:       "#4682B4",
		MenuSelectedColor: "#A0522D",
		MenuTextColor:     "black",
		MenuDisabledColor: "gray",
		ShowWelcomePopup:  true,
		ConfirmOnQuit:     true,
		ShowHiddenFiles:   false,
		Theme:             "monokai",
		GithubUser:        "github_user",
		GithubKey:         "github_key",
		GithubEmail:       "github_email",
		MaxRecentFiles:    10,
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
