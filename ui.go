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
	fileMenu    []MenuEntry
	viewEntries []MenuEntry
	editor      *femto.View
	config      conf.Config
)

// ****************************************************************************
// setUI()
// ****************************************************************************
func setUI() {
	config = conf.LoadConfig("config.json")
	tview.Styles.PrimitiveBackgroundColor = conf.GetColor(config.MenuBgColor)
	tview.Styles.PrimaryTextColor = conf.GetColor(config.MenuTextColor)

	app = tview.NewApplication()
	pages = tview.NewPages()
	status = tview.NewTextView().
		SetTextAlign(tview.AlignLeft).
		SetDynamicColors(true)

	status.SetBackgroundColor(conf.GetColor(config.MenuBgColor))
	status.SetTextColor(conf.GetColor(config.MenuTextColor))

	// --- L'ÉDITEUR ---
	buffer := femto.NewBufferFromString(string(""), "./dummy")
	editor = femto.NewView(buffer)
	editor.Buf.Settings["keepautoindent"] = true
	editor.Buf.Settings["softwrap"] = true
	editor.Buf.Settings["scrollbar"] = true
	editor.Buf.Settings["statusline"] = false

	// --- LE MENU ---
	menuBar = NewAppMenuBar(app, pages)

	fileMenu = []MenuEntry{
		{Label: "New", Action: func() { /*...*/ }},
		{Label: "Open", Action: func() { /*...*/ }},
		{Label: "Save", Disabled: true, Action: func() { /*...*/ }},    // Désactivé par défaut
		{Label: "Save as", Disabled: true, Action: func() { /*...*/ }}, // Désactivé par défaut
		{Label: "Quit", Shortcut: tcell.KeyCtrlQ, Action: func() { app.Stop() }},
	}

	// Dans la logique de votre éditeur (TextArea)
	/*
		editor.SetChangedFunc(func() {
			if len(editor.GetText()) > 0 {
				fichierEntries[1].Disabled = false // On active "Sauvegarder" dès qu'il y a du texte
			} else {
				fichierEntries[1].Disabled = true
			}
		})

		editor.SetMovedFunc(func() {
			_, _, row, col := editor.GetCursor()
			// row+1 et col+1 car les index commencent à 0
			msg := fmt.Sprintf("Ligne: %d, Col: %d | F10: Menu | Ctrl+Q: Quitter", row+1, col+1)
			updateStatus(msg)
		})
	*/

	// Et on ajoute le menu normalement
	menuBar.AddMenu(" File ", fileMenu)

	viewEntries = []MenuEntry{
		{
			Label:       "Numéros de ligne",
			IsCheckable: true,
			Checked:     true, // Activé par défaut
			Action: func() {
				// On récupère l'état via une variable ou en parcourant le menu
				// Ici, on bascule l'affichage dans le TextArea
				// editor.SetShowLineNumbers(!editor.GetShowLineNumbers())
			},
		},
		{
			Label:       "Mode Sombre",
			IsCheckable: true,
			Checked:     false,
			Action: func() {
				if editor.GetBackgroundColor() == tcell.ColorBlack {
					editor.SetBackgroundColor(tcell.ColorWhite)
					// editor.SetTextStyle(tcell.StyleDefault.Foreground(tcell.ColorBlack))
				} else {
					editor.SetBackgroundColor(tcell.ColorBlack)
					// editor.SetTextStyle(tcell.StyleDefault.Foreground(tcell.ColorWhite))
				}
			},
		},
	}

	menuBar.AddMenu(" Affichage ", viewEntries)

	// --- LAYOUT ---
	// On empile la barre de menu (hauteur 1) sur l'éditeur (tout le reste)
	layout := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(menuBar, 1, 0, false).
		AddItem(editor, 0, 1, true). // Éditeur (milieu - flexible)
		AddItem(status, 1, 0, false)

	pages.AddPage("main", layout, true, true)
	SetStatus(fmt.Sprintf(" %s v%s", conf.APP_NAME, getFullVersion()))

	go func() {
		for {
			time.Sleep(500 * time.Millisecond) // Prevent too much CPU usage
			app.QueueUpdateDraw(func() {
				refreshStatus()
			})
		}
	}()
}

// ****************************************************************************
// SetStatus()
// ****************************************************************************
func SetStatus(txt string) {
	status.SetText(txt)
	DurationOfTime := time.Duration(conf.STATUS_MESSAGE_DURATION) * time.Second
	f := func() {
		status.SetText("")
	}
	time.AfterFunc(DurationOfTime, f)
	/*
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
	*/
}

// ****************************************************************************
// refreshStatus()
// ****************************************************************************
func refreshStatus() {
	cursor := CurrentFile.FemtoView.Cursor
	// Femto utilise des coordonnées 0-indexed, on ajoute +1 pour l'humain
	// statusLeft.SetText(fmt.Sprintf(" [Fichier: %s]", editor.Buf.GetName()))
	// statusRight.SetText(fmt.Sprintf("Ln %d, Col %d ", cursor.Y+1, cursor.X+1))
	SetStatus(fmt.Sprintf(" [File: %s] Line: %d, Column: %d", filepath.Base(CurrentFile.FName), cursor.Y+1, cursor.X+1))
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
