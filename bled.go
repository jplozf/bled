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
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

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
	ACmd             []string
	ICmd             int
	currentDir       string
	GITInfos         string
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
	GITInfos = fmt.Sprintf("%s %s %s", conf.APP_ICON, conf.APP_NAME, getFullVersion())
	Macros = make(map[string]string)
	setUI()
	go checkUpdate()

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
			pages.SwitchToPage("main")
			app.SetFocus(editor)
			if searchPanel.active {
				layout.ResizeItem(searchPanel, 0, 0)
				app.SetFocus(editor)
				searchPanel.active = false
			}
			if gotoPanel.active {
				layout.ResizeItem(gotoPanel, 0, 0)
				app.SetFocus(editor)
				gotoPanel.active = false
			}
			return nil

		case tcell.KeyF10:
			if menuBar.HasFocus() {
				menuBar.closeAllMenus()
				app.SetFocus(editor)
			} else {
				toggleGotoPanel(false)
				toggleSearchPanel(false)
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
	})

	// Editor keyboard's events manager
	editor.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if CurrentFile != nil && CurrentFile.FollowMode {
			CurrentFile.FollowMode = false
			SetStatus("Follow Mode: OFF")
			return nil
		}

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

		if event.Key() == tcell.KeyTab {
			isSoft, width := getTabConfig()

			if isSoft {
				// SOFT TABS : insert spaces instead of a tab character
				for i := 0; i < width; i++ {
					CurrentFile.FemtoView.InsertSpace()
				}
				return nil
			}

			// HARD TABS : Insert a tab character
			return event
		}

		switch event.Key() {
		case tcell.KeyBackspace, tcell.KeyBackspace2:
			editor.Backspace()
		}
		return event
	})

	ReadRecent()
	refreshFileMenu()

	if len(os.Args) > 1 {
		// Open file(s) if provided as command-line argument(s)
		for _, arg := range os.Args[1:] {
			absPath, err := filepath.Abs(arg)
			if err != nil {
				SetStatus("Error resolving path: " + err.Error())
				continue
			}
			openFile(absPath, false)
			currentDir = filepath.Dir(absPath)
		}
	} else {
		// Open a new file if no filename is provided
		newFile()
		currentDir, _ = os.Getwd()
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
	msg := fmt.Sprintf("Welcome to %s v%s\n\nVisit the GitHub repository :\n%s\n\nF1  : Help\nF10 : Menu\nEsc : Close this panel", conf.APP_NAME, getFullVersion(), conf.APP_URL)
	MsgBox = MsgBox.Info("Welcome", msg, nil, 0, "main", editor)
	pages.AddPage("msgNewVersion", MsgBox.Popup(), true, false)
	pages.ShowPage("msgNewVersion")
}

// ****************************************************************************
// safeQuit()
// ****************************************************************************
func safeQuit() {
	var fileToProcess *efile
	i := 0
	for _, f := range efiles {
		if f.FemtoBuffer != nil && f.FemtoBuffer.Modified() {
			fileToProcess = f
			break
		}
		i++
	}

	if fileToProcess == nil {
		if config.ConfirmOnQuit {
			confirmQuitOnly()
			return
		} else {
			terminateApp()
			return
		}
	}

	switchDocument(i)

	DlgQuit = DlgQuit.YesNoCancel("Quit", // Title
		fmt.Sprintf("File '%s'\nhas been modified.\n\nDo you want to save it before quitting ?", fileToProcess.FName), // Message
		func(rc DlgButton, idx int) {
			switch rc {
			case BUTTON_YES:
				saveFile()
				safeQuit()
			case BUTTON_NO:
				fileToProcess.FemtoBuffer.IsModified = false
				safeQuit()
			case BUTTON_CANCEL:
				SetStatus("Canceling quit")
			}
		},
		0, "main", editor)

	pages.AddPage("dlgQuit", DlgQuit.Popup(), true, false)
	pages.ShowPage("dlgQuit")
}

// ****************************************************************************
// terminateApp()
// ****************************************************************************
func terminateApp() {
	SaveMacros()
	SaveRecent()
	app.Stop()
	fmt.Printf("%s %s v%s - %s\n", conf.APP_ICON, conf.APP_NAME, getFullVersion(), conf.APP_URL)
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
	DlgInputFileOpen = DlgInputFileOpen.FileBrowser("Open File", // Title
		currentDir,
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
		SetStatus("Error : No file to close")
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

// ****************************************************************************
// getTabConfig()
// ****************************************************************************
func getTabConfig() (isSoft bool, width int) {
	if CurrentFile == nil {
		return false, 8 // Par défaut standard
	}

	// Si c'est du Python, on impose les Soft Tabs (4 espaces)
	if strings.HasSuffix(strings.ToLower(CurrentFile.FName), ".py") {
		return true, 4
	}

	// Pour le Go ou les autres, on utilise souvent des Hard Tabs (largeur 8 ou 4)
	return false, 4
}

// ****************************************************************************
// convertTabsToSpaces()
// ****************************************************************************
func convertTabsToSpaces(buf *femto.Buffer, width int) int {
	count := 0
	spaces := strings.Repeat(" ", width)

	// On parcourt chaque ligne du buffer
	for i := 0; i < buf.NumLines; i++ {
		line := string(buf.Line(i))
		if strings.Contains(line, "\t") {
			// Remplacement de tous les \t par les espaces
			newLine := strings.ReplaceAll(line, "\t", spaces)

			// On remplace la ligne dans le buffer
			// Note: On utilise souvent Replace() sur le buffer pour garder l'Undo
			buf.Replace(femto.Loc{X: 0, Y: i}, femto.Loc{X: len(line), Y: i}, newLine)
			count++
		}
	}
	return count
}

// ****************************************************************************
// smartDetab()
// ****************************************************************************
func smartDetab(line string, width int) string {
	spaces := strings.Repeat(" ", width)
	for strings.HasPrefix(line, "\t") {
		line = strings.Replace(line, "\t", spaces, 1)
	}
	return line
}

// ****************************************************************************
// isTextFile()
// ****************************************************************************
func isTextFile(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()

	// Read a chunk of the file to analyze its content
	buffer := make([]byte, 512)
	n, err := f.Read(buffer)
	if err != nil && err != io.EOF {
		return false, err
	}

	// Mime type detection based on the content
	contentType := http.DetectContentType(buffer[:n])

	// if the content type starts with "text/" or is "application/octet-stream", we consider it as a text file
	if strings.HasPrefix(contentType, "text/") || contentType == "application/octet-stream" {
		// Look for null bytes in the content, which are common in binary files but rare in text files.
		for i := 0; i < n; i++ {
			if buffer[i] == 0 {
				return false, nil // Null byte found, likely a binary file
			}
		}
		return true, nil
	}

	return false, nil
}

// ****************************************************************************
// getScrollPercentage()
// ****************************************************************************
func getScrollPercentage() int {
	if CurrentFile == nil || CurrentFile.FemtoBuffer == nil {
		return 0
	}

	totalLines := CurrentFile.FemtoBuffer.NumLines
	if totalLines <= 1 {
		return 100
	}

	currentLine := CurrentFile.FemtoView.Cursor.Y
	percent := (currentLine * 100) / (totalLines - 1)

	return percent
}

// ****************************************************************************
// toggleGotoPanel()
// ****************************************************************************
func toggleGotoPanel(show bool) {
	if show {
		// On cache le panneau de recherche s'il est ouvert pour éviter l'empilement
		toggleSearchPanel(false)

		layout.ResizeItem(gotoPanel, 1, 0) // On lui donne 1 ligne de hauteur
		gotoPanel.active = true
		app.SetFocus(gotoPanel.input)
	} else {
		layout.ResizeItem(gotoPanel, 0, 0) // On le cache
		gotoPanel.active = false
		if CurrentFile != nil {
			app.SetFocus(CurrentFile.FemtoView)
		}
	}
}

// ****************************************************************************
// toggleSearchPanel()
// ****************************************************************************
func toggleSearchPanel(show bool) {
	if show {
		// On cache le panneau de recherche s'il est ouvert pour éviter l'empilement
		toggleGotoPanel(false)

		layout.ResizeItem(searchPanel, 2, 0) // On lui donne 2 lignes de hauteur
		searchPanel.active = true
		app.SetFocus(searchPanel.searchInput)
	} else {
		layout.ResizeItem(searchPanel, 0, 0) // On le cache
		searchPanel.active = false
		if CurrentFile != nil {
			app.SetFocus(CurrentFile.FemtoView)
		}
	}
}

// ****************************************************************************
// confirmQuitOnly()
// ****************************************************************************
func confirmQuitOnly() {
	DlgQuit = DlgQuit.YesNo("Quit", "Are you sure you want to quit ?",
		func(rc DlgButton, idx int) {
			if rc == BUTTON_YES {
				terminateApp()
			}
		},
		0, "main", editor)

	pages.AddPage("dlgConfirm", DlgQuit.Popup(), true, false)
	pages.ShowPage("dlgConfirm")
}

// ****************************************************************************
// getRemoteHash()
// ****************************************************************************
func getRemoteHash() (string, error) {
	client := &http.Client{Timeout: 5 * time.Second}
	url := "https://api.github.com/repos/jplozf/Bled/commits/main"

	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		Sha string `json:"sha"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if len(result.Sha) > 7 {
		return result.Sha[:7], nil
	}
	return result.Sha, nil
}

// ****************************************************************************
// checkUpdate()
// ****************************************************************************
func checkUpdate() {
	remoteHash, err := getRemoteHash()
	if err != nil {
		return
	}

	if GitVersion != "dev" {
		localHash := GitVersion[len(GitVersion)-7:]
		if !strings.EqualFold(remoteHash, localHash) {
			msg := fmt.Sprintf("An update is available : %s (local)\n                         %s (remote)\n\nVisit the GitHub repository :\n%s", localHash, remoteHash, conf.APP_URL)
			MsgBox = MsgBox.Info("Update available", msg, nil, 0, "main", editor)
			pages.AddPage("msgNewVersion", MsgBox.Popup(), true, false)
			pages.ShowPage("msgNewVersion")
		}
	}
}
