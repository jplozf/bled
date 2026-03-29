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
	"bled/utils"
	"fmt"
	"path/filepath"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/pgavlin/femto"
	"github.com/pgavlin/femto/runtime"
	"github.com/rivo/tview"
)

// ****************************************************************************
// VARS
// ****************************************************************************
var (
	fileMenu                                                             []MenuEntry
	editEntries                                                          []MenuEntry
	helpEntries                                                          []MenuEntry
	editor                                                               *femto.View
	config                                                               conf.Config
	MsgBox                                                               *Dialog
	statusFilePos, statusSize, statusModified, statusTime, statusMessage *tview.TextView
	layout                                                               *tview.Flex
	searchPanel                                                          *SearchPanel
)

type SearchPanel struct {
	*tview.Flex
	input  *tview.InputField
	label  *tview.TextView
	active bool
}

// ****************************************************************************
// NewSearchPanel()
// ****************************************************************************
func NewSearchPanel() *SearchPanel {
	s := &SearchPanel{
		Flex:  tview.NewFlex().SetDirection(tview.FlexColumn),
		input: tview.NewInputField().SetLabel(" Find: ").SetFieldWidth(30),
		label: tview.NewTextView().SetDynamicColors(true),
	}

	// Style du panneau
	s.input.SetFieldBackgroundColor(tcell.ColorBlack).SetLabelColor(tcell.ColorBlack).SetFieldTextColor(tcell.ColorYellow)
	s.Flex.AddItem(s.input, 0, 1, true).
		AddItem(s.label, 20, 0, false)

	return s
}

// ****************************************************************************
// setUI()
// ****************************************************************************
func setUI() {
	config = conf.LoadConfig()
	tview.Styles.PrimitiveBackgroundColor = conf.GetColor(config.MenuBgColor)
	tview.Styles.PrimaryTextColor = conf.GetColor(config.MenuTextColor)

	app = tview.NewApplication()
	pages = tview.NewPages()
	status = tview.NewTextView().
		SetTextAlign(tview.AlignLeft).
		SetDynamicColors(true)

	status.SetBackgroundColor(conf.GetColor(config.MenuBgColor))
	status.SetTextColor(conf.GetColor(config.MenuTextColor))

	// EDITOR
	buffer := femto.NewBufferFromString(string(""), "./dummy")
	editor = femto.NewView(buffer)
	editor.Buf.Settings["keepautoindent"] = true
	editor.Buf.Settings["softwrap"] = true
	editor.Buf.Settings["scrollbar"] = true
	editor.Buf.Settings["statusline"] = false

	// MENUS
	menuBar = NewAppMenuBar(app, pages)

	fileMenu = []MenuEntry{
		{Label: "New", Action: func() { newFile() }, Shortcut: tcell.KeyCtrlN},
		{Label: "Open", Action: func() { InputFileOpen() }, Shortcut: tcell.KeyCtrlO},
		{Label: "Save", Disabled: true, Action: func() { saveFile() }, Shortcut: tcell.KeyCtrlS},
		{Label: "Save as", Action: func() { SaveFileAs() }},
		{Label: "Close", Action: func() { closeCurrentFile() }, Shortcut: tcell.KeyCtrlT},
		{Label: "Quit", Shortcut: tcell.KeyCtrlQ, Action: func() { safeQuit() }},
	}
	menuBar.AddMenu(" File ", fileMenu)

	editEntries = []MenuEntry{
		{Label: "Goto", Action: func() { InputGotoLine() }, Shortcut: tcell.KeyCtrlG},
		{Label: "Find", Action: func() {
			layout.ResizeItem(searchPanel, 1, 0)
			app.SetFocus(searchPanel.input)
			searchPanel.active = true
		}, Shortcut: tcell.KeyCtrlF},
		{Label: "Replace", Action: func() { /*...*/ }},
	}
	menuBar.AddMenu(" Edit ", editEntries)

	helpEntries = []MenuEntry{
		{Label: "Manual", Action: func() { ShowManual() }, Shortcut: tcell.KeyF1},
		{Label: "About", Action: func() { ShowWelcomePopup() }},
		{Label: "Settings", Action: func() { openSettings() }},
	}
	menuBar.AddMenu(" Help ", helpEntries)

	// STATUS BAR COMPONENTS
	statusFilePos = tview.NewTextView().SetDynamicColors(true)
	statusMessage = tview.NewTextView().SetDynamicColors(true)
	statusSize = tview.NewTextView().SetDynamicColors(true).SetTextAlign(tview.AlignCenter)
	statusModified = tview.NewTextView().SetDynamicColors(true).SetTextAlign(tview.AlignCenter)
	statusTime = tview.NewTextView().SetDynamicColors(true).SetTextAlign(tview.AlignRight)

	// Horizontal layout for the status bar
	statusBar := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(statusFilePos, 0, 1, false).
		AddItem(statusMessage, 0, 1, false).
		AddItem(statusSize, 12, 0, false).
		AddItem(statusModified, 10, 0, false).
		AddItem(statusTime, 10, 0, false)

	searchPanel = NewSearchPanel()
	searchPanel.input.SetChangedFunc(func(text string) {
		if text == "" {
			searchPanel.label.SetText("")
			return
		}
		totalMatches = performSearch(text)
		searchPanel.label.SetText(fmt.Sprintf(" [black]%d occurence(s)[-]", totalMatches))
	})
	searchPanel.input.SetDoneFunc(func(key tcell.Key) {
		switch key {
		case tcell.KeyEnter:
			query := searchPanel.input.GetText()
			if query != "" {
				lastSearchQuery = query
				// searchPanel.active = false
				app.SetFocus(editor)
				jumpToNextMatch()
			}
		case tcell.KeyEsc:
			layout.ResizeItem(searchPanel, 0, 0)
			searchPanel.active = false
			app.SetFocus(editor)
		}
	})

	// --- LAYOUT ---
	layout = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(menuBar, 1, 0, false).
		AddItem(editor, 0, 1, true).
		AddItem(searchPanel, 0, 0, false).
		AddItem(statusBar, 1, 0, false)

	pages.AddPage("main", layout, true, true)

	// Refresh status bar every 500ms
	go func() {
		for {
			time.Sleep(500 * time.Millisecond) // Prevent too much CPU usage
			app.QueueUpdateDraw(func() {
				refreshStatus()
			})
		}
	}()

	startMessageWorker()

	// Show welcome message on startup
	if config.ShowWelcomePopup {
		ShowWelcomePopup()
	}
}

// ****************************************************************************
// SetStatus()
// ****************************************************************************
func SetStatus(txt string) {
	select {
	case messageQueue <- txt:
	default:
		// If the channel is full, we can choose to drop the message or block until there's space.
		// Here, we choose to drop the message to avoid blocking the application.
	}
}

// ****************************************************************************
// refreshStatus()
// ****************************************************************************
func refreshStatus() {
	name := filepath.Base(CurrentFile.FName)
	if name == "" {
		name = "[New File]"
	}

	cursorX, cursorY := editor.Cursor.X, editor.Cursor.Y
	statusFilePos.SetText(fmt.Sprintf("[%s] %s  Ln %d, Col %d", getFilePagination(), name, cursorY+1, cursorX+1))

	size := editor.Buf.Len()
	statusSize.SetText(utils.HumanFileSize(float64(size)))

	modifiedText := ""
	if CurrentFile.ReadOnly {
		modifiedText = "[yellow]READ-ONLY[-]"
		fileMenu[2].Disabled = true // Disable "Save" if the file is read-only
	} else if CurrentFile.FemtoBuffer != nil && CurrentFile.FemtoBuffer.Modified() {
		modifiedText = "[red]modified[-]"
		fileMenu[2].Disabled = false // Enable "Save"
	} else {
		modifiedText = ""
	}
	statusModified.SetText(modifiedText)

	statusTime.SetText(time.Now().Format("15:04:05 "))
}

// ****************************************************************************
// SetTheme()
// ****************************************************************************
func SetTheme(theme string) {
	if runtimeTheme := runtime.Files.FindFile(femto.RTColorscheme, theme); runtimeTheme != nil {
		if data, err := runtimeTheme.Data(); err == nil {
			var colorscheme femto.Colorscheme
			colorscheme = femto.ParseColorscheme(string(data))
			editor.SetRuntimeFiles(runtime.Files)
			editor.SetColorscheme(colorscheme)
		}
	}
}

// ****************************************************************************
// startMessageWorker()
// ****************************************************************************
func startMessageWorker() {
	go func() {
		duration := time.Duration(config.StatusMessageDuration) * time.Second
		for msg := range messageQueue {
			app.QueueUpdateDraw(func() {
				statusMessage.SetText(msg)
			})

			time.Sleep(duration)

			app.QueueUpdateDraw(func() {
				statusMessage.SetText("")
			})

			time.Sleep(200 * time.Millisecond) // more delay to prevent message overlap if many messages are sent in a short time
		}
	}()
}

// ****************************************************************************
// refreshFileMenu()
// ****************************************************************************
func refreshFileMenu() {
	fileMenu = []MenuEntry{
		{Label: "New", Action: func() { newFile() }, Shortcut: tcell.KeyCtrlN},
		{Label: "Open", Action: func() { InputFileOpen() }, Shortcut: tcell.KeyCtrlO},
		{Label: "Save", Disabled: true, Action: func() { saveFile() }, Shortcut: tcell.KeyCtrlS},
		{Label: "Save as", Action: func() { SaveFileAs() }},
		{Label: "Close", Action: func() { closeCurrentFile() }, Shortcut: tcell.KeyCtrlT},
		{Label: "Quit", Shortcut: tcell.KeyCtrlQ, Action: func() { safeQuit() }},
	}

	for i, f := range efiles {
		displayName := filepath.Base(f.FName)
		if displayName == "." || displayName == "" {
			displayName = "[noname]"
		}

		prefix := "  "
		if CurrentFile != nil && f == CurrentFile {
			prefix = "✓ "
		}

		targetIdx := i
		fileMenu = append(fileMenu, MenuEntry{
			Label:  prefix + displayName,
			Action: func() { switchDocument(targetIdx) },
		})
	}
	menuBar.UpdateMenu(" File ", fileMenu)
}
