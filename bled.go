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
package main

// ****************************************************************************
// IMPORTS
// ****************************************************************************
import (
	"bled/conf"
	"fmt"
	"os"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// ****************************************************************************
// VARS
// ****************************************************************************
var (
	app          *tview.Application
	menuBar      *AppMenuBar
	pages        *tview.Pages
	status       *tview.TextView
	messageQueue = make(chan string, 100) // queue for status messages, with a buffer of 100 messages
	DlgQuit      *Dialog
)

// MAJOR Version number, injected at build time
const MAJOR = "0"

// GitVersion is the number of git commits and the git hash, injected at build time
var GitVersion = "dev"

// ****************************************************************************
// main()
// ****************************************************************************
func main() {
	// Setup the UI components
	setUI()

	// Global input capture for shortcuts and focus management
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// Manage global shortcuts
		if result := menuBar.HandleShortcuts(event); result == nil {
			return nil
		}

		switch event.Key() {
		case tcell.KeyF10:
			if menuBar.HasFocus() {
				menuBar.closeAllMenus()
				app.SetFocus(editor)
			} else {
				app.SetFocus(menuBar)
				menuBar.OpenMenu()
			}
			return nil
		case tcell.KeyF6:
			prevFile()
			return nil
		case tcell.KeyF7:
			nextFile()
			return nil
		}
		return event
	})

	// Editor keyboard's events manager
	editor.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// On traite la touche
		switch event.Key() {
		case tcell.KeyBackspace, tcell.KeyBackspace2:
			editor.Backspace()
		}

		return event
	})

	if len(os.Args) > 1 {
		filename := os.Args[1]
		openFile(filename)
	}

	// Start the application
	if err := app.SetRoot(pages, true).Run(); err != nil {
		panic(err)
	}
}

// ****************************************************************************
// getFullVersion()
// ****************************************************************************
func getFullVersion() string {
	return fmt.Sprintf("%s.%s", MAJOR, GitVersion)
}

// ****************************************************************************
// ShowWelcomePopup()
// ****************************************************************************
func ShowWelcomePopup() {
	msg := fmt.Sprintf("Welcome to %s v%s\n\nVisit the GitHub repository :\n%s", conf.APP_NAME, getFullVersion(), conf.APP_URL)
	MsgBox = MsgBox.OK("Welcome", msg, nil, 0, "main", editor)
	pages.AddPage("msgNewVersion", MsgBox.Popup(), true, false)
	pages.ShowPage("msgNewVersion")
}

// ****************************************************************************
// safeQuit()
// ****************************************************************************
func safeQuit() {
	if CurrentFile.FemtoBuffer != nil && CurrentFile.FemtoBuffer.Modified() {
		DlgQuit = DlgQuit.YesNo("Quit", // Title
			"You have unsaved changes.\nDo you want to quit without saving ?", // Message
			func(rc DlgButton, idx int) {
				if rc == BUTTON_YES {
					app.Stop()
				} else {
					SetStatus("Canceling quit")
				}
			},
			0,
			"main", editor)
		pages.AddPage("dlgQuit", DlgQuit.Popup(), true, false)
		pages.ShowPage("dlgQuit")
	} else {
		app.Stop()
		fmt.Printf("%s %s v%s - %s\n", conf.APP_ICON, conf.APP_NAME, getFullVersion(), conf.APP_URL)
	}
}

// ****************************************************************************
// openSettings()
// ****************************************************************************
func openSettings() {
	path := conf.GetConfigPath()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		SetStatus("[red]Error : The configuration file does not exist yet.[-]")
		return
	}
	openFile(path)
	SetStatus("Configuration loaded : " + path)
	app.SetFocus(editor)
}
