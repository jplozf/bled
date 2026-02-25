package main

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func main() {
	app := tview.NewApplication()
	pages := tview.NewPages()

	// --- L'ÉDITEUR ---
	editor := tview.NewTextArea()
	editor.SetPlaceholder("Commencez à écrire ici...")
	editor.SetBorder(true).SetTitle(" Editeur de texte ")

	// --- LE MENU ---
	menuBar := NewAppMenuBar(app, pages)

	menuBar.AddMenu(" Fichier ", []MenuEntry{
		{Label: "Nouveau", Shortcut: tcell.KeyCtrlN, Action: func() {
			editor.SetText("", false)
		}},
		{Label: "Quitter", Shortcut: tcell.KeyCtrlQ, Action: func() {
			app.Stop()
		}},
	}).AddMenu(" Edition ", []MenuEntry{
		{Label: "Tout effacer", Action: func() {
			editor.SetText("", false)
		}},
	})

	// --- LAYOUT ---
	// On empile la barre de menu (hauteur 1) sur l'éditeur (tout le reste)
	layout := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(menuBar, 1, 0, true).
		AddItem(editor, 0, 1, false)

	pages.AddPage("main", layout, true, true)

	// Gestion des raccourcis et focus
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// 1. On vérifie les raccourcis du menu
		if result := menuBar.HandleShortcuts(event); result == nil {
			return nil
		}

		// 2. F10 pour basculer le focus entre menu et éditeur
		if event.Key() == tcell.KeyF10 {
			if menuBar.HasFocus() {
				app.SetFocus(editor)
			} else {
				app.SetFocus(menuBar)
			}
			return nil
		}
		return event
	})

	if err := app.SetRoot(pages, true).Run(); err != nil {
		panic(err)
	}
}
