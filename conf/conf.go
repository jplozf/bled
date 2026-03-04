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

	"github.com/gdamore/tcell/v2"
)

const (
	STATUS_MESSAGE_DURATION = 3
	APP_NAME                = "Bled"
	APP_STRING              = "Bled © jpl@ozf.fr 2026"
	APP_URL                 = "https://github.com/jplozf/bled"
	APP_FOLDER              = ".bled"
	ICON_MODIFIED           = "●"
	ICON_DATABASE           = "⛁"
	ICON_EXPLORER           = "🗁"
	NEW_FILE_TEMPLATE       = "noname_"
	NEW_DATABASE_TEMPLATE   = "db_"
	FILE_LOG                = "bled.log"
	FILE_INI                = "bled.ini"
	FILE_FIND_HISTORY       = "find"
	FILE_SHELL_HISTORY      = "history"
	FILE_SQL_HISTORY        = "sql"
	FILE_SHELL_OUTPUT       = "output"
	FILE_MACROS             = "macros"
	FILE_MRU                = "mru"
	FKEY_LABELS             = "F1=Help F2=Panel F3=Git F4=Command F6=Previous F7=Next F8=Settings F9=Context F10=Menu F12=Exit"
	CKEY_LABELS             = "Ctrl+E=Explorer Ctrl+F=Find… Ctrl+S=Save Alt+S=Save as… Ctrl+N=New Ctrl+O=Open… Ctrl+T=Close Alt+M=Macros"
	DEFAULT_COLOR_ACCENT    = "#556B2F"
	FILE_MAX_PREVIEW        = 1024
	HASH_THRESHOLD_SIZE     = 1_073_741_824.0
	COLOR_FOLDER            = tcell.ColorLightGreen
	COLOR_FILE              = tcell.ColorYellow
	COLOR_EXECUTABLE        = tcell.ColorLightYellow
	COLOR_SELECTED          = tcell.ColorRed
	LABEL_PARENT_FOLDER     = "<UP>"
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
	MenuBgColor       string `json:"menu_bg_color"`       // Jaune
	MenuSelectedColor string `json:"menu_selected_color"` // Rouge
	MenuTextColor     string `json:"menu_text_color"`     // Noir
	MenuDisabledColor string `json:"menu_disabled_color"` // Gris
}

func LoadConfig(path string) Config {
	// Valeurs par défaut si le fichier n'existe pas
	conf := Config{
		MenuBgColor:       "yellow",
		MenuSelectedColor: "red",
		MenuTextColor:     "black",
		MenuDisabledColor: "gray",
	}

	data, err := os.ReadFile(path)
	if err == nil {
		json.Unmarshal(data, &conf)
	}
	return conf
}

// Fonction utilitaire pour convertir le string JSON en tcell.Color
func GetColor(name string) tcell.Color {
	return tcell.GetColor(name)
}
