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
	"github.com/pgavlin/femto"
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
	lastSearchQuery  string
	totalMatches     int
	currentMatchIdx  int
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

		// ALT+S
		evkSaveAs := tcell.NewEventKey(tcell.KeyRune, 's', tcell.ModAlt)
		if event.Key() == evkSaveAs.Key() && event.Rune() == evkSaveAs.Rune() && event.Modifiers() == evkSaveAs.Modifiers() {
			SaveFileAs()
			return nil
		}

		switch event.Key() {
		case tcell.KeyF4:
			if CurrentFile != nil {
				jumpToNextMatch()
			}
			return nil

		case tcell.KeyEsc:
			if searchPanel.active {
				layout.ResizeItem(searchPanel, 0, 0)
				app.SetFocus(editor)
				searchPanel.active = false
				return nil
			}

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
		default:
			return event
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
	for i, f := range efiles {
		if f.FName == "Bled Manual" {
			switchDocument(i)
			SetStatus("Switching to help manual")
			return
		}
	}

	helpBuf := femto.NewBufferFromString(helpText, "Help.txt")

	helpFile := &efile{
		FName:       "Bled Manual",
		FemtoBuffer: helpBuf,
		FemtoView:   femto.NewView(helpBuf),
		ReadOnly:    true,
	}

	efiles = append(efiles, helpFile)
	switchDocument(len(efiles) - 1)

	SetStatus("Opening help manual")
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

var helpText = `⚶ B L E D   -   Copyright © JPL 2026

Bled is a TUI (Text User Interface) Editor.
Bled is written in Go and has been tested on Linux sytem.
Built from source, it should run on Windows or MacOS systems as well.
Several files can be opened at the same time, and you can switch between them easily.

Pay honour to whom honour is due, packages used in this project are as follows :
rivo/tview    : Package tview implements rich widgets for terminal based user interfaces.
gdamore/tcell : Tcell is an alternate terminal package, similar in some ways to termbox, but better in others. 
pgavlin/femto : An editor component for tview. Derived from the micro editor. 

⯈ The main functions are reachable through function keys :

F1  : This help text
F4  : Jump to the next match of the current search query
F6  : Switch to the previous open file
F7  : Switch to the next open file
F10 : Access to the main menu of Bled

⯈ Alternate common functions are also reachable through CTRL and ALT keys :

CTRL + F : Switch to the Find & Replace panel
CTRL + S : Saves the current document being edited
ALT  + S : Saves the current document being edited under another name
CTRL + N : Opens a new blank document
CTRL + O : Opens an existing document for editing
CTRL + T : Closes the current document
CTRL + G : Opens the "Goto line" dialog to jump to a specific line number
CTRL + Q : Quit Bled

⯈ When editing a text, common editing functions are of course supported :

CTRL + C : Copy the selection
CTRL + X : Cut the selection
CTRL + V : Paste the selection
CTRL + Z : Cancels the previous entry 
CTRL + Y : Redo the previous cancelled operation

⯈ Settings are stored in a configuration file, as a JSON file located in the user's home directory :

menu_bg_color           : background color of the menu bar
menu_selected_color     : background color of the selected menu item
menu_text_color         : color of the text in the menu
menu_disabled_color     : color of disabled menu items
show_welcome_popup      : whether to show a welcome popup at startup
confirm_on_quit         : whether to ask for confirmation when quitting with unsaved changes
status_message_duration : duration in seconds for which status messages are displayed
show_hidden_files       : whether to show hidden files in the file browser dialog
color_accent            : accent color used in the UI (e.g. for highlights)
theme                   : theme name (see below for available themes)

⯈ Available themes are as follows (theme names are case-sensitive) :

atom-dark-tc
bubblegum
cmc-16
cmc-paper
cmc-tc
darcula
default
geany
github-tc
gruvbox-tc
gruvbox
material-tc
monokai
railscast
simple
solarized-tc
solarized
twilight
zenburn

`

// ----------------------------------------------------------------------------
//    Some common useful icons
// ----------------------------------------------------------------------------
//   ⚶
//   ●
//   ⯈
//   ⎇
//   ⟟
//   🗨
//   ⚠
//   ©
// ----------------------------------------------------------------------------
