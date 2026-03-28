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
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// ****************************************************************************
// VARS
// ****************************************************************************
var (
	app              *tview.Application
	menuBar          *AppMenuBar
	pages            *tview.Pages
	status           *tview.TextView
	messageQueue     = make(chan string, 100) // queue for status messages, with a buffer of 100 messages
	DlgQuit          *Dialog
	DlgInputFileOpen *Dialog
	DlgSaveFileAs    *Dialog
	DlgInputGotoLine *Dialog
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
		if CurrentFile != nil && CurrentFile.ReadOnly {
			// Only allow navigation keys and some basic actions in read-only mode
			switch event.Key() {
			case tcell.KeyUp, tcell.KeyDown, tcell.KeyLeft, tcell.KeyRight,
				tcell.KeyPgUp, tcell.KeyPgDn, tcell.KeyHome, tcell.KeyEnd:
				return event
			default:
				// If a key produces a character or a modification action, we cancel it (return nil)
				if event.Rune() != 0 || event.Key() == tcell.KeyEnter || event.Key() == tcell.KeyTab {
					return nil
				}
				return nil
			}
		}

		switch event.Key() {
		case tcell.KeyBackspace, tcell.KeyBackspace2:
			editor.Backspace()
		}
		return event
	})

	if len(os.Args) > 1 {
		// Open file(s) if provided as command-line argument(s)
		for _, arg := range os.Args[1:] {
			absPath, err := filepath.Abs(arg)
			if err != nil {
				SetStatus("Error resolving path: " + err.Error())
				continue
			}
			openFile(absPath, false)
		}
	} else {
		// Open a new file if no filename is provided
		newFile()
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
	openFile(path, false)
	SetStatus("Configuration loaded : " + path)
	app.SetFocus(editor)
}

// ****************************************************************************
// InputFileOpen()
// ****************************************************************************
func InputFileOpen() {
	path, err := os.Getwd()
	if err != nil {
		SetStatus("Error getting current working directory : " + err.Error())
		return
	}
	DlgInputFileOpen = DlgInputFileOpen.FileBrowser("Open File", // Title
		path,
		doOpenFile,
		0,
		"main", editor, false)
	pages.AddPage("dlgInputFileOpen", DlgInputFileOpen.Popup(), true, false)
	pages.ShowPage("dlgInputFileOpen")
	app.SetFocus(DlgInputFileOpen)
}

// ****************************************************************************
// doOpenFile()
// ****************************************************************************
func doOpenFile(rc DlgButton, idx int) {
	if rc == BUTTON_OK {
		fn := DlgInputFileOpen.Value
		SetStatus("Opening " + fn)
		openFile(fn, false)
	}
}

// ****************************************************************************
// SaveFileAs()
// ****************************************************************************
func SaveFileAs() {
	SetStatus("Save as...")
	DlgSaveFileAs = DlgSaveFileAs.Input("Save File as...", // Title
		"New name for this file :", // Message
		CurrentFile.FName,
		confirmSaveAs,
		0,
		"main", editor) // Focus return
	pages.AddPage("dlgSaveFileAs", DlgSaveFileAs.Popup(), true, false)
	pages.ShowPage("dlgSaveFileAs")
}

// ****************************************************************************
// confirmSaveAs()
// ****************************************************************************
func confirmSaveAs(rc DlgButton, idx int) {
	if rc == BUTTON_OK {
		newName := DlgSaveFileAs.Value
		err := ioutil.WriteFile(newName, []byte(CurrentFile.FemtoBuffer.String()), 0600)
		if err == nil {
			SetStatus(fmt.Sprintf("File %s successfully saved", newName))
			CurrentFile.FName = newName
			CurrentFile.FemtoBuffer.IsModified = false
			closeCurrentFile()
			openFile(newName, false)
		} else {
			SetStatus(err.Error())
		}
	}
}

// ****************************************************************************
// InputGotoLine()
// ****************************************************************************
func InputGotoLine() {
	SetStatus("Goto line...")
	DlgInputGotoLine = DlgInputGotoLine.Input("Goto Line...", // Title
		"Go to line :",                       // Message
		fmt.Sprintf("%d", editor.Cursor.Y+1), // Pre-fill with current line number (1-based)
		confirmGotoLine,
		0,
		"main", editor) // Focus return
	pages.AddPage("dlgInputGotoLine", DlgInputGotoLine.Popup(), true, false)
	pages.ShowPage("dlgInputGotoLine")
}

// ****************************************************************************
// confirmGotoLine()
// ****************************************************************************
func confirmGotoLine(rc DlgButton, idx int) {
	if rc == BUTTON_OK {
		line := DlgInputGotoLine.Value
		lineNum, err := strconv.Atoi(line)
		if err != nil {
			SetStatus("Error: invalid line number")
			return
		}
		GoLine(lineNum)
	}
}

// ****************************************************************************
// ShowManual()
// ****************************************************************************
func ShowManual() {
	msg := "F1     : Show this manual\nF6     : Previous file\nF7     : Next file\nF10    : Focus/Unfocus menu bar\nCtrl+N : New file\nCtrl+O : Open file\nCtrl+S : Save file\nCtrl+T : Close file\nCtrl+G : Goto line\nCtrl+Q : Quit"
	MsgBox = MsgBox.OK("Manual", msg, nil, 0, "main", editor)
	pages.AddPage("msgManual", MsgBox.Popup(), true, false)
	pages.ShowPage("msgManual")
}

// ****************************************************************************
// closeCurrentFile()
// ****************************************************************************
func closeCurrentFile() {
	if CurrentFile == nil {
		SetStatus("[red]Error : No file to close[-]")
		return
	}
	targetIdx := -1
	for i, f := range efiles {
		if f == CurrentFile {
			targetIdx = i
			break
		}
	}
	if CurrentFile.FemtoBuffer != nil && CurrentFile.FemtoBuffer.Modified() {
		DlgQuit = DlgQuit.YesNo("Close File", // Title
			fmt.Sprintf("File %s has unsaved changes.\nDo you want to close without saving ?", CurrentFile.FName), // Message
			func(rc DlgButton, idx int) {
				if rc == BUTTON_YES {
					closeFile(targetIdx)
				} else {
					SetStatus("Canceling close")
				}
			},
			0,
			"main", editor)
		pages.AddPage("dlgCloseFile", DlgQuit.Popup(), true, false)
		pages.ShowPage("dlgCloseFile")
	} else {
		closeFile(targetIdx)
	}
}
