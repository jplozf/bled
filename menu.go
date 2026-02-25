package main

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// MenuEntry représente une option dans un menu déroulant
type MenuEntry struct {
	Label    string
	Shortcut tcell.Key // ex: tcell.KeyCtrlQ
	Action   func()
}

// AppMenuBar est notre composant réutilisable
type AppMenuBar struct {
	*tview.Flex
	app        *tview.Application
	pages      *tview.Pages
	buttons    []*tview.Button
	shortcuts  map[tcell.Key]func() // Table de hachage pour accès rapide
	activeBtn  int
	currentPos int
}

func NewAppMenuBar(app *tview.Application, pages *tview.Pages) *AppMenuBar {
	m := &AppMenuBar{
		Flex:      tview.NewFlex().SetDirection(tview.FlexColumn),
		app:       app, // <--- On stocke l'instance
		pages:     pages,
		shortcuts: make(map[tcell.Key]func()),
	}

	// Navigation horizontale au clavier
	m.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyRight {
			m.activeBtn = (m.activeBtn + 1) % len(m.buttons)
			m.app.SetFocus(m.buttons[m.activeBtn])
			return nil
		} else if event.Key() == tcell.KeyLeft {
			m.activeBtn = (m.activeBtn - 1 + len(m.buttons)) % len(m.buttons)
			m.app.SetFocus(m.buttons[m.activeBtn])
			return nil
		}
		return event
	})

	return m
}

// AddMenu ajoute un bouton et son menu déroulant associé
func (m *AppMenuBar) AddMenu(title string, entries []MenuEntry) *AppMenuBar {
	xOffset := m.currentPos
	width := len(title) + 4

	list := tview.NewList().ShowSecondaryText(false)
	list.SetBorder(true)

	for _, entry := range entries {
		action := entry.Action
		list.AddItem(entry.Label, "", 0, func() {
			m.pages.RemovePage("dropdown")
			if action != nil {
				action()
			}
		})

		// Enregistrement du raccourci global s'il existe
		if entry.Shortcut != tcell.KeyRune {
			m.shortcuts[entry.Shortcut] = action
		}
	}

	btn := tview.NewButton(title).SetSelectedFunc(func() {
		dropdown := tview.NewGrid().
			SetColumns(xOffset, 22, 0).
			SetRows(1, list.GetItemCount()+2, 0).
			AddItem(list, 1, 1, 1, 1, 0, 0, true)
		m.pages.AddPage("dropdown", dropdown, true, true)
		m.app.SetFocus(list)
	})

	m.buttons = append(m.buttons, btn)
	m.AddItem(btn, width, 0, false)
	m.currentPos += width
	return m
}

// HandleShortcuts doit être appelé dans le SetInputCapture principal de l'app
func (m *AppMenuBar) HandleShortcuts(event *tcell.EventKey) *tcell.EventKey {
	if action, exists := m.shortcuts[event.Key()]; exists {
		if action != nil {
			action()
		}
		return nil // On consomme l'événement
	}
	return event
}
