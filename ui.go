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
	"os"
	"path/filepath"
	"strings"
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
	gitEntries                                                           []MenuEntry
	macroEntries                                                         []MenuEntry
	editor                                                               *femto.View
	config                                                               conf.Config
	MsgBox                                                               *Dialog
	statusFilePos, statusSize, statusModified, statusTime, statusMessage *tview.TextView
	statusTabs                                                           *tview.TextView
	layout                                                               *tview.Flex
	searchPanel                                                          *SearchPanel
	gotoPanel                                                            *GotoPanel
	statusBar                                                            *tview.Flex
	HighlightedColor                                                     tcell.Color
	TextColor                                                            tcell.Color
	DisabledTextColor                                                    tcell.Color
	SelectedTextColor                                                    tcell.Color
)

// ****************************************************************************
// TYPES
// ****************************************************************************
type SearchPanel struct {
	*tview.Flex
	searchInput  *tview.InputField
	replaceInput *tview.InputField
	label        *tview.TextView
	caseCheck    *tview.Checkbox
	active       bool
}

type GotoPanel struct {
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
		Flex:         tview.NewFlex().SetDirection(tview.FlexRow),
		searchInput:  tview.NewInputField().SetLabel(" Find    : ").SetFieldWidth(30),
		replaceInput: tview.NewInputField().SetLabel(" Replace : ").SetFieldWidth(30),
		label:        tview.NewTextView().SetDynamicColors(true),
		caseCheck:    tview.NewCheckbox().SetLabel(" Case sensitive : "),
	}

	// Action on case sensitivity change
	s.caseCheck.SetChangedFunc(func(checked bool) {
		// Relaunch the search with the new case sensitivity setting
		performSearch(s.searchInput.GetText())
	})

	s.searchInput.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyTab {
			app.SetFocus(s.caseCheck)
			return nil
		}
		return event
	})

	s.caseCheck.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyTab {
			app.SetFocus(s.replaceInput)
			return nil
		}
		return event
	})

	s.replaceInput.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyTab {
			app.SetFocus(s.searchInput)
			return nil
		}

		evkReplaceOne := tcell.NewEventKey(tcell.KeyRune, 'r', tcell.ModAlt)
		if event.Key() == evkReplaceOne.Key() && event.Rune() == evkReplaceOne.Rune() && event.Modifiers() == evkReplaceOne.Modifiers() {
			replaceCurrent()
			return nil
		}

		evkReplaceAll := tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModAlt)
		if event.Key() == evkReplaceAll.Key() && event.Rune() == evkReplaceAll.Rune() && event.Modifiers() == evkReplaceAll.Modifiers() {
			replaceAll()
			return nil
		}

		return event
	})

	// Panel's style
	s.searchInput.SetFieldBackgroundColor(TextColor).SetLabelColor(TextColor).SetFieldTextColor(conf.GetColor(config.MenuBgColor))
	s.replaceInput.SetFieldBackgroundColor(TextColor).SetLabelColor(TextColor).SetFieldTextColor(conf.GetColor(config.MenuBgColor))
	s.caseCheck.SetFieldBackgroundColor(TextColor).SetLabelColor(TextColor).SetFieldTextColor(conf.GetColor(config.MenuBgColor))
	s.label.SetBackgroundColor(conf.GetColor(config.MenuBgColor))

	row1 := tview.NewFlex().
		AddItem(s.searchInput, 0, 1, true).
		AddItem(s.caseCheck, 20, 0, false).
		AddItem(s.label, 20, 0, false)

	row2 := tview.NewFlex().
		AddItem(s.replaceInput, 0, 1, false).
		AddItem(tview.NewTextView().SetText(" [Alt+R] Replace [Alt+A] All"), 30, 0, false)

	s.AddItem(row1, 1, 0, true).
		AddItem(row2, 1, 0, false)

	return s
}

// ****************************************************************************
// NewGotoPanel()
// ****************************************************************************
func NewGotoPanel() *GotoPanel {
	g := &GotoPanel{
		Flex:  tview.NewFlex().SetDirection(tview.FlexColumn),
		input: tview.NewInputField().SetLabel(" Goto : ").SetFieldWidth(30),
		label: tview.NewTextView().SetDynamicColors(true).SetText(" (Ex: 50, 80%, -10, top, end)"),
	}
	g.input.SetFieldBackgroundColor(TextColor).SetLabelColor(TextColor).SetFieldTextColor(conf.GetColor(config.MenuBgColor))
	g.label.SetBackgroundColor(conf.GetColor(config.MenuBgColor))
	g.AddItem(g.input, 0, 1, true).
		AddItem(g.label, 30, 0, false)

	return g
}

// ****************************************************************************
// setUI()
// ****************************************************************************
func setUI() {
	config = conf.LoadConfig()
	TextColor = tcell.GetColor(utils.GetContrastedTextColor(config.MenuBgColor))
	SelectedTextColor = tcell.GetColor(utils.GetContrastedTextColor(config.MenuSelectedColor))
	tview.Styles.PrimitiveBackgroundColor = conf.GetColor(config.MenuBgColor)
	tview.Styles.PrimaryTextColor = TextColor
	_, targetHue := utils.GetClosestTargetHue(utils.HexToRGB(config.MenuSelectedColor))
	r, g, b := utils.HexToRGB(config.MenuBgColor)
	HighlightedColor = tcell.GetColor(utils.GetStyledTextColor(r, g, b, targetHue))

	app = tview.NewApplication()
	pages = tview.NewPages()
	status = tview.NewTextView().
		SetTextAlign(tview.AlignLeft).
		SetDynamicColors(true)

	status.SetBackgroundColor(conf.GetColor(config.MenuBgColor))
	status.SetTextColor(TextColor)

	// EDITOR
	buffer := femto.NewBufferFromString(string(""), "./dummy")
	editor = femto.NewView(buffer)
	editor.Buf.Settings["keepautoindent"] = true
	editor.Buf.Settings["softwrap"] = true
	editor.Buf.Settings["scrollbar"] = true
	editor.Buf.Settings["statusline"] = false

	// MENUS
	menuBar = NewAppMenuBar(app, pages)
	setMenuUI()

	// STATUS BAR COMPONENTS
	statusFilePos = tview.NewTextView().SetDynamicColors(true)
	statusMessage = tview.NewTextView().SetDynamicColors(true)
	statusSize = tview.NewTextView().SetDynamicColors(true).SetTextAlign(tview.AlignCenter)
	statusTabs = tview.NewTextView().SetDynamicColors(true).SetTextAlign(tview.AlignCenter)
	statusModified = tview.NewTextView().SetDynamicColors(true).SetTextAlign(tview.AlignCenter)
	statusTime = tview.NewTextView().SetDynamicColors(true).SetTextAlign(tview.AlignRight)

	// Horizontal layout for the status bar
	statusBar = tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(statusFilePos, 0, 1, false).
		AddItem(tview.NewTextView().SetText("⯈"), 1, 1, false).
		AddItem(statusMessage, 0, 1, false).
		AddItem(statusSize, 20, 0, false).
		AddItem(statusTabs, 10, 0, false).
		AddItem(statusModified, 10, 0, false).
		AddItem(statusTime, 10, 0, false)

	gotoPanel = NewGotoPanel()
	gotoPanel.input.SetDoneFunc(func(key tcell.Key) {
		switch key {
		case tcell.KeyEnter:
			text := gotoPanel.input.GetText()
			if text != "" {
				gotoLocation(text)
				gotoPanel.input.SetText("")
			}
			toggleGotoPanel(false)
			gotoPanel.active = false
			app.SetFocus(editor)
		case tcell.KeyEsc:
			toggleGotoPanel(false)
			gotoPanel.active = false
			app.SetFocus(editor)
		}
	})

	searchPanel = NewSearchPanel()
	searchPanel.searchInput.SetChangedFunc(func(text string) {
		if text == "" {
			searchPanel.label.SetText("")
			return
		}
		totalMatches = performSearch(text)
		searchPanel.label.SetText(fmt.Sprintf(" [%s]%d occurence(s)[-]", TextColor, totalMatches))
	})

	searchPanel.searchInput.SetDoneFunc(func(key tcell.Key) {
		switch key {
		case tcell.KeyEnter:
			query := searchPanel.searchInput.GetText()
			if query != "" {
				lastSearchQuery = query
				app.SetFocus(editor)
				jumpToNextMatch()
			}
		case tcell.KeyEsc:
			layout.ResizeItem(searchPanel, 0, 0)
			searchPanel.active = false
			app.SetFocus(editor)
		}
	})

	initInfoPanel()

	// --- LAYOUT ---
	layout = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(menuBar, 1, 0, false).
		AddItem(editor, 0, 1, true).
		AddItem(gotoPanel, 0, 0, false).
		AddItem(searchPanel, 0, 0, false).
		AddItem(infoPanel, 0, 0, false).
		AddItem(statusBar, 1, 0, false)

	pages.AddPage("main", layout, true, true)

	// Refresh status bar every 500ms
	go func() {
		for {
			time.Sleep(100 * time.Millisecond) // Prevent too much CPU usage
			app.QueueUpdateDraw(func() {
				refreshStatus()
			})
		}
	}()

	// startMessageWorker()
	StartStatusConsumer()

	// Show welcome message on startup
	if config.ShowWelcomePopup {
		ShowWelcomePopup()
	}
}

// ****************************************************************************
// SetStatus()
// ****************************************************************************
func SetStatus(txt string) {
	splittedText := strings.Split(txt, "\n")
	if len(splittedText) <= 1 {
		current := time.Now()
		conf.LogFile.WriteString(fmt.Sprintf("%s [%s] : %s\n", current.Format("20060102-150405"), SessionID, txt))
	} else {
		for _, s := range splittedText {
			current := time.Now()
			conf.LogFile.WriteString(fmt.Sprintf("%s [%s] : %s\n", current.Format("20060102-150405"), SessionID, s))
		}
	}

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
	percent := getScrollPercentage()
	fileStatus := fmt.Sprintf("[%s] [%s]%s[-]  Ln %d, Col %d", getFilePagination(), HighlightedColor, name, cursorY+1, cursorX+1)
	statusFilePos.SetText(fileStatus)
	statusBar.ResizeItem(statusFilePos, NetLength(fileStatus)+1, 1)

	size := editor.Buf.Len()
	statusSize.SetText(utils.HumanFileSize(float64(size)) + fmt.Sprintf(" (%d%%)", percent))

	modifiedText := ""
	if CurrentFile.ReadOnly {
		// modifiedText = fmt.Sprintf("[%s]READ-ONLY[-]", conf.GetColor(config.MenuSelectedColor))
		modifiedText = fmt.Sprintf("[%s]READ-ONLY[-]", HighlightedColor)
		fileMenu[5].Disabled = true // Disable "Save" if the file is read-only
	} else if CurrentFile.FemtoBuffer != nil && CurrentFile.FemtoBuffer.Modified() {
		// modifiedText = fmt.Sprintf("[%s]MODIFIED[-]", conf.GetColor(config.MenuSelectedColor))
		modifiedText = fmt.Sprintf("[%s]MODIFIED[-]", HighlightedColor)
		fileMenu[5].Disabled = false // Enable "Save"
	} else if CurrentFile.FollowMode {
		// modifiedText = fmt.Sprintf("[%s]FOLLOWED[-]", conf.GetColor(config.MenuSelectedColor))
		modifiedText = fmt.Sprintf("[%s]FOLLOWED[-]", HighlightedColor)
	} else {
		modifiedText = ""
	}
	statusModified.SetText(modifiedText)

	isSoft, width := getTabConfig()
	mode := "TAB"
	if isSoft {
		mode = "SPC"
	}
	statusTabs.SetText(fmt.Sprintf("%s:%d", mode, width))

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
// StartStatusConsumer()
// ****************************************************************************
func StartStatusConsumer() {
	// Listener goroutine for the message channel
	go func() {
		for msg := range messageQueue {
			ribbonMutex.Lock()
			statusRibbon += " | " + msg
			ribbonMutex.Unlock()
		}
	}()

	// Scrolling goroutine for the status ribbon
	go func() {
		ticker := time.NewTicker(120 * time.Millisecond)
		defer ticker.Stop()

		for range ticker.C {
			ribbonMutex.Lock()
			if len(statusRibbon) == 0 {
				ribbonMutex.Unlock()
				continue
			}

			runes := []rune(statusRibbon)
			statusRibbon = string(runes[1:])
			display := statusRibbon
			ribbonMutex.Unlock()

			app.QueueUpdateDraw(func() {
				_, _, width, _ := statusMessage.GetInnerRect()
				if len(display) > width && width > 0 {
					statusMessage.SetText(display[:width])
				} else {
					statusMessage.SetText(display)
				}
			})
			// Clean up the ribbon if it's empty after scrolling
			if len(strings.TrimSpace(statusRibbon)) == 0 {
				statusRibbon = ""
				app.QueueUpdateDraw(func() { statusMessage.SetText("") })
			}
		}
	}()
}

// ****************************************************************************
// refreshFileMenu()
// ****************************************************************************
func refreshFileMenu() {
	// Rebuild the "Recent" submenu with the updated list of recent files
	recentEntries := []MenuEntry{}

	for _, path := range RecentFiles {
		filePath := path
		recentEntries = append(recentEntries, MenuEntry{
			Label:  filepath.Base(filePath),
			Action: func() { openFile(filePath, false) },
		})
	}

	if len(recentEntries) == 0 {
		recentEntries = append(recentEntries, MenuEntry{Label: "Aucun fichier récent", Disabled: true})
	}

	// Rebuild the "New from Template" submenu
	templateDir := filepath.Join(conf.GetConfigDir(), "templates")
	os.MkdirAll(templateDir, 0755) // Sécurité
	templateSubMenu := buildTemplateMenu(templateDir)

	// Rebuild the "File" menu
	fileMenu = []MenuEntry{
		{Label: "New", Action: func() { newFile() }, Shortcut: tcell.KeyCtrlN},
		{
			Label:      "New from Template",
			SubEntries: templateSubMenu,
		},
		{Label: "Open", Action: func() { InputFileOpen() }, Shortcut: tcell.KeyCtrlO},
		{
			Label:      "Recent",
			SubEntries: recentEntries,
		},
		{IsSeparator: true},
		{Label: "Save", Disabled: true, Action: func() { saveFile() }, Shortcut: tcell.KeyCtrlS},
		{Label: "Save as", Action: func() { SaveFileAs() }},
		{Label: "Close", Action: func() { closeCurrentFile() }, Shortcut: tcell.KeyCtrlT},
		{Label: "Quit", Shortcut: tcell.KeyCtrlQ, Action: func() { safeQuit() }},
		{Label: "Shell", Action: func() { shellEscape() }},
		{IsSeparator: true},
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

// ****************************************************************************
// setMenuUI()
// ****************************************************************************
func setMenuUI() {
	fileMenu = []MenuEntry{
		{Label: "New", Action: func() { newFile() }, Shortcut: tcell.KeyCtrlN},
		{Label: "Open", Action: func() { InputFileOpen() }, Shortcut: tcell.KeyCtrlO},
		{Label: "Save", Disabled: true, Action: func() { saveFile() }, Shortcut: tcell.KeyCtrlS},
		{Label: "Save as", Action: func() { SaveFileAs() }},
		{Label: "Close", Action: func() { closeCurrentFile() }, Shortcut: tcell.KeyCtrlT},
		{Label: "Quit", Shortcut: tcell.KeyCtrlQ, Action: func() { safeQuit() }},
		{IsSeparator: true},
	}
	menuBar.AddMenu(" File ", fileMenu)

	gitEntries = []MenuEntry{
		{Label: "Status", Action: func() { DoGitStatus() }},
		{Label: "Log", Action: func() { DoGitLog() }},
		{Label: "Add all (.)", Action: func() { DoGitAddAll() }},
		{Label: "Commit", Action: func() { DoGitCommit() }},
		{Label: "Push", Action: func() { DoGitPush() }},
		{Label: "Commit & Push", Action: func() { DoGitCommitPush() }},
		{Label: "Fetch", Action: func() { DoGitFetch() }},
		{Label: "Pull (Fetch & Merge)", Action: func() { DoGitPull() }},
		{Label: "Initialize", Action: func() { DoGitInit() }},
		{Label: "Initialize & Push", Action: func() { DoGitBang() }},
		{Label: "Clone", Action: func() { DoGitClone() }},
	}

	refreshMacrosMenu()

	editEntries = []MenuEntry{
		{Label: "Goto", Action: func() { toggleGotoPanel(true) }, Shortcut: tcell.KeyCtrlG},
		{Label: "Find & Replace", Action: func() {
			toggleSearchPanel(true)
		}, Shortcut: tcell.KeyCtrlF},
		{Label: "Git Tracking", SubEntries: gitEntries, Action: func() { menuBar.ShowMenuPopup("Git Tracking", gitEntries) }, Shortcut: tcell.KeyF3},
		{Label: "Macros", SubEntries: macroEntries, Action: func() { menuBar.ShowMenuPopup("Macros", macroEntries) }, Shortcut: tcell.KeyF5},
		{Label: "Info", Action: func() { toggleInfoPanel() }, Shortcut: tcell.KeyCtrlK},
	}
	menuBar.AddMenu(" Edit ", editEntries)

	helpEntries = []MenuEntry{
		{Label: "Manual", Action: func() { ShowManual() }, Shortcut: tcell.KeyF1},
		{Label: "About", Action: func() { ShowWelcomePopup() }},
		{Label: "Settings", Action: func() { openSettings() }},
	}
	menuBar.AddMenu(" Help ", helpEntries)
}

// ****************************************************************************
// buildTemplateMenu()
// ****************************************************************************
func buildTemplateMenu(path string) []MenuEntry {
	var entries []MenuEntry

	files, err := os.ReadDir(path)
	if err != nil {
		return nil
	}

	for _, f := range files {
		fullPath := filepath.Join(path, f.Name())

		if f.IsDir() {
			subEntries := buildTemplateMenu(fullPath)
			if len(subEntries) > 0 {
				entries = append(entries, MenuEntry{
					Label:      f.Name(),
					SubEntries: subEntries,
				})
			}
		} else {
			tName := f.Name()
			tPath := fullPath

			entries = append(entries, MenuEntry{
				Label: tName,
				Action: func() {
					CreateFromTemplate(tPath)
				},
			})
		}
	}
	return entries
}
