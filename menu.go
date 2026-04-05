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
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// ****************************************************************************
// TYPES
// ****************************************************************************
type MenuEntry struct {
	Label       string
	Shortcut    tcell.Key
	Action      func()
	SubEntries  []MenuEntry
	Disabled    bool
	IsCheckable bool
	Checked     bool
}

type AppMenuBar struct {
	*tview.Flex
	app           *tview.Application
	pages         *tview.Pages
	buttons       []*tview.Button
	shortcuts     map[tcell.Key]func()
	activeBtn     int
	currentPos    int
	spacer        *tview.Box
	menuTitles    []string
	menuData      map[string][]MenuEntry
	activeList    *tview.List
	mainPrimitive tview.Primitive
	versionView   *tview.TextView
	menuStack     []string
}

// ****************************************************************************
// NewAppMenuBar()
// ****************************************************************************
func NewAppMenuBar(app *tview.Application, pages *tview.Pages) *AppMenuBar {
	m := &AppMenuBar{
		Flex:          tview.NewFlex().SetDirection(tview.FlexColumn),
		app:           app,
		pages:         pages,
		mainPrimitive: editor,
		shortcuts:     make(map[tcell.Key]func()),
	}

	// Horizontale navigation between menu buttons with Left/Right keys
	m.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyRight:
			m.activeBtn = (m.activeBtn + 1) % len(m.buttons)
			m.app.SetFocus(m.buttons[m.activeBtn])
			return nil
		case tcell.KeyLeft:
			m.activeBtn = (m.activeBtn - 1 + len(m.buttons)) % len(m.buttons)
			m.app.SetFocus(m.buttons[m.activeBtn])
			return nil
		}
		return event
	})

	// Background color for the menu bar
	m.SetBackgroundColor(conf.GetColor(conf.LoadConfig().MenuBgColor))

	// Initialisation de la version dès le départ
	versionStr := fmt.Sprintf("%s %s %s", conf.APP_ICON, conf.APP_NAME, getFullVersion())
	m.versionView = tview.NewTextView().
		SetTextAlign(tview.AlignRight).
		SetDynamicColors(true).
		SetText(versionStr + " ")

	m.versionView.SetBackgroundColor(conf.GetColor(conf.LoadConfig().MenuBgColor))
	m.versionView.SetTextColor(conf.GetColor(conf.LoadConfig().MenuTextColor))

	return m
}

// ****************************************************************************
// UpdateMenu()
// ****************************************************************************
func (m *AppMenuBar) UpdateMenu(title string, entries []MenuEntry) {
	if m.menuData == nil {
		m.menuData = make(map[string][]MenuEntry)
	}
	// On met à jour uniquement les données
	m.menuData[title] = entries
	m.registerShortcuts(entries)
}

// ****************************************************************************
// AddMenu()
// ****************************************************************************
func (m *AppMenuBar) AddMenu(title string, entries []MenuEntry) *AppMenuBar {
	if m.menuData == nil {
		m.menuData = make(map[string][]MenuEntry)
	}

	// --- Sinon, c'est la création initiale du bouton ---
	// m.menuTitles = append(m.menuTitles, title)
	m.menuData[title] = entries
	m.registerShortcuts(entries)
	buttonExists := false
	for _, btn := range m.buttons {
		if btn.GetLabel() == title {
			buttonExists = true
			break
		}
	}
	// 3. Si le bouton n'existe pas, on le crée et on l'ajoute
	if !buttonExists {
		btn := tview.NewButton(title)

		// --- Ton code de style (Yellow/Orange) ici ---
		// On définit explicitement les styles pour CHAQUE état
		styleNormal := tcell.StyleDefault.Background(conf.GetColor(conf.LoadConfig().MenuBgColor)).Foreground(conf.GetColor(conf.LoadConfig().MenuTextColor))
		styleFocus := tcell.StyleDefault.Background(conf.GetColor(conf.LoadConfig().MenuSelectedColor)).Foreground(conf.GetColor(conf.LoadConfig().MenuTextColor))

		// 1. Style au repos
		btn.SetStyle(styleNormal)
		btn.SetLabelColor(conf.GetColor(conf.LoadConfig().MenuTextColor))
		btn.SetBackgroundColor(conf.GetColor(conf.LoadConfig().MenuBgColor))

		// 2. Style quand le bouton est sélectionné (le fameux bleu par défaut)
		// On force l'Orange (ou le Jaune) pour écraser le bleu du thème
		btn.SetActivatedStyle(styleFocus)
		btn.SetBackgroundColorActivated(conf.GetColor(conf.LoadConfig().MenuSelectedColor))
		btn.SetLabelColorActivated(conf.GetColor(conf.LoadConfig().MenuTextColor))

		// Style au repos (Jaune comme la barre)
		btn.SetBackgroundColor(conf.GetColor(conf.LoadConfig().MenuBgColor))
		btn.SetLabelColor(conf.GetColor(conf.LoadConfig().MenuTextColor))

		// Style quand on navigue dessus (Orange pour voir où on est)
		btn.SetBackgroundColorActivated(conf.GetColor(conf.LoadConfig().MenuSelectedColor))
		btn.SetLabelColorActivated(conf.GetColor(conf.LoadConfig().MenuTextColor))

		btn.SetSelectedFunc(func() {
			bx, by, _, _ := btn.GetRect()
			m.showDropdown(entries, bx, by+1, "dropdown")
		})

		btn.SetSelectedFunc(func() {
			bx, by, _, _ := btn.GetRect()
			// On récupère les données dynamiques au clic
			m.showDropdown(m.menuData[title], bx, by+1, "dropdown")
		})

		m.buttons = append(m.buttons, btn)
		m.rebuildBar() // On redessine pour inclure le nouveau bouton
	}

	return m
}

// ****************************************************************************
// rebuildBar()
// ****************************************************************************
func (m *AppMenuBar) rebuildBar() {
	// 1. On vide complètement le Flex
	m.Clear()

	// 2. On ajoute tous les boutons de menu (à gauche)
	for _, btn := range m.buttons {
		// On calcule la largeur selon le texte du bouton + padding
		width := len(btn.GetLabel()) + 2
		m.AddItem(btn, width, 0, false)
	}

	// 3. On ajoute le spacer (il prend tout l'espace restant : poids 1)
	if m.spacer == nil {
		m.spacer = tview.NewBox().SetBackgroundColor(conf.GetColor(config.MenuBgColor))
	}
	m.AddItem(m.spacer, 0, 1, false)

	// 4. On ajoute la version (à droite)
	if m.versionView != nil {
		// On récupère le texte brut pour calculer la largeur sans les tags [color]
		versionText := m.versionView.GetText(true)
		m.AddItem(m.versionView, len(versionText)+1, 0, false)
	}
}

/*
// ****************************************************************************
// AddMenu_old()
// ****************************************************************************
func (m *AppMenuBar) AddMenu_old(title string, entries []MenuEntry) *AppMenuBar {
	m.menuData = append(m.menuData, entries)
	m.registerShortcuts(entries)

	btn := tview.NewButton(title)

	// On définit explicitement les styles pour CHAQUE état
	styleNormal := tcell.StyleDefault.Background(conf.GetColor(conf.LoadConfig().MenuBgColor)).Foreground(conf.GetColor(conf.LoadConfig().MenuTextColor))
	styleFocus := tcell.StyleDefault.Background(conf.GetColor(conf.LoadConfig().MenuSelectedColor)).Foreground(conf.GetColor(conf.LoadConfig().MenuTextColor))

	// 1. Style au repos
	btn.SetStyle(styleNormal)
	btn.SetLabelColor(conf.GetColor(conf.LoadConfig().MenuTextColor))
	btn.SetBackgroundColor(conf.GetColor(conf.LoadConfig().MenuBgColor))

	// 2. Style quand le bouton est sélectionné (le fameux bleu par défaut)
	// On force l'Orange (ou le Jaune) pour écraser le bleu du thème
	btn.SetActivatedStyle(styleFocus)
	btn.SetBackgroundColorActivated(conf.GetColor(conf.LoadConfig().MenuSelectedColor))
	btn.SetLabelColorActivated(conf.GetColor(conf.LoadConfig().MenuTextColor))

	// Style au repos (Jaune comme la barre)
	btn.SetBackgroundColor(conf.GetColor(conf.LoadConfig().MenuBgColor))
	btn.SetLabelColor(conf.GetColor(conf.LoadConfig().MenuTextColor))

	// Style quand on navigue dessus (Orange pour voir où on est)
	btn.SetBackgroundColorActivated(conf.GetColor(conf.LoadConfig().MenuSelectedColor))
	btn.SetLabelColorActivated(conf.GetColor(conf.LoadConfig().MenuTextColor))

	btn.SetSelectedFunc(func() {
		bx, by, _, _ := btn.GetRect()
		m.showDropdown(entries, bx, by+1, "dropdown")
	})

	m.buttons = append(m.buttons, btn)

	// 1. On retire l'ancien spacer et l'ancienne version s'ils existent
	// Pour éviter de les accumuler à chaque fois qu'on ajoute un menu
	if m.spacer != nil {
		m.RemoveItem(m.spacer)
	}
	// Si vous avez stocké m.versionView dans votre struct AppMenuBar :
	if m.versionView != nil {
		m.RemoveItem(m.versionView)
	}

	// 2. On ajoute le bouton actuel
	m.AddItem(btn, len(title)+2, 0, false)

	// 3. On recrée le spacer (il occupe tout l'espace vide au milieu)
	m.spacer = tview.NewBox().SetBackgroundColor(conf.GetColor(conf.LoadConfig().MenuBgColor))
	m.AddItem(m.spacer, 0, 1, false)

	// 4. On ajoute la version COMPLÈTEMENT À DROITE
	versionStr := conf.APP_ICON + " " + conf.APP_NAME + " " + getFullVersion()
	m.versionView = tview.NewTextView().
		SetTextAlign(tview.AlignRight).
		SetDynamicColors(true).
		SetText(versionStr + " ")

	m.versionView.SetBackgroundColor(conf.GetColor(conf.LoadConfig().MenuBgColor))
	m.versionView.SetTextColor(conf.GetColor(conf.LoadConfig().MenuTextColor))

	// On ajoute la version avec une largeur fixe (longueur du texte + 1)
	m.AddItem(m.versionView, len(versionStr)+1, 0, false)

	return m
}
*/

// ****************************************************************************
// registerShortcuts()
// ****************************************************************************
func (m *AppMenuBar) registerShortcuts(entries []MenuEntry) {
	for _, e := range entries {
		if e.Shortcut != tcell.KeyRune && e.Action != nil {
			m.shortcuts[e.Shortcut] = e.Action
		}
		if len(e.SubEntries) > 0 {
			m.registerShortcuts(e.SubEntries)
		}
	}
}

// ****************************************************************************
// showDropdown()
// ****************************************************************************
func (m *AppMenuBar) showDropdown(entries []MenuEntry, x, y int, pageName string) {
	// 1. Calcul de la largeur dynamique
	maxLabelWidth := 0
	for _, entry := range entries {
		length := len(entry.Label) + 3
		if entry.IsCheckable {
			length += 4
		}
		if len(entry.SubEntries) > 0 {
			length += 3
		}
		if length > maxLabelWidth {
			maxLabelWidth = length
		}
	}
	listWidth := maxLabelWidth + 2
	if listWidth < 15 {
		listWidth = 15
	}

	listHeight := len(entries)

	// Détection de bord d'écran
	_, _, sw, sh := m.pages.GetInnerRect()
	if sw > 0 && x+listWidth > sw {
		x = sw - listWidth
	}
	if sh > 0 && y+listHeight > sh {
		y = sh - listHeight
	}

	list := tview.NewList().ShowSecondaryText(false)
	list.SetBorder(false)
	list.SetBackgroundColor(conf.GetColor(config.MenuBgColor))
	list.SetMainTextColor(conf.GetColor(config.MenuTextColor))
	list.SetSelectedBackgroundColor(conf.GetColor(config.MenuSelectedColor))
	list.SetSelectedTextColor(tcell.ColorWhite)
	list.SetSelectedFocusOnly(false)

	// Si c'est le premier menu (dropdown principal), on vide la pile
	if pageName == "dropdown" {
		m.menuStack = []string{pageName}
		m.activeList = list
	} else {
		// Sinon on ajoute ce sous-menu à la pile
		m.menuStack = append(m.menuStack, pageName)
	}

	for i := range entries {
		e := &entries[i]
		idx := i // Capture de l'index pour le closure
		label := e.Label

		prefix := "  "
		if e.IsCheckable {
			if e.Checked {
				prefix = "[\u221A] "
			} else {
				prefix = "[ ] "
			}
		}

		padding := strings.Repeat(" ", maxLabelWidth-len(label)-len(prefix))
		displayLabel := prefix + label + padding

		if e.Disabled {
			displayLabel = fmt.Sprintf("[%s]%s[%s]", conf.GetColor(config.MenuDisabledColor), displayLabel, conf.GetColor(config.MenuTextColor))
		} else if len(e.SubEntries) > 0 {
			displayLabel += " >"
		}

		list.AddItem(displayLabel, "", 0, func() {
			if e.Disabled {
				return
			}

			// --- GESTION SOUS-MENU AU CLIC ---
			if len(e.SubEntries) > 0 {
				subPageName := fmt.Sprintf("submenu_%d", len(m.menuStack))
				m.showDropdown(e.SubEntries, x+listWidth, y+idx, subPageName)
				return
			}

			if e.IsCheckable {
				e.Checked = !e.Checked
				m.closeAllMenus()
			} else {
				m.closeAllMenus()
			}

			if e.Action != nil {
				e.Action()
			}
		})
	}

	list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		currIdx := list.GetCurrentItem()
		entry := entries[currIdx]

		switch event.Key() {
		case tcell.KeyRight:
			// Ouvrir le sous-menu si présent
			if len(entry.SubEntries) > 0 {
				subPageName := fmt.Sprintf("submenu_%d", len(m.menuStack))
				m.showDropdown(entry.SubEntries, x+listWidth, y+currIdx, subPageName)
				return nil
			}
			// Si on est au niveau 0 et pas de sous-menu, on change de menu principal
			if pageName == "dropdown" {
				m.closeAllMenus()
				m.activeBtn = (m.activeBtn + 1) % len(m.buttons)
				m.OpenCurrentMenu()
				return nil
			}

		case tcell.KeyLeft:
			// Si c'est un sous-menu, on revient en arrière
			if len(m.menuStack) > 1 {
				lastPage := m.menuStack[len(m.menuStack)-1]
				m.menuStack = m.menuStack[:len(m.menuStack)-1]
				m.pages.RemovePage(lastPage)
				// Le focus revient automatiquement à la page précédente visible
				return nil
			}
			// Si on est au niveau 0, on change de menu principal
			if pageName == "dropdown" {
				m.closeAllMenus()
				m.activeBtn = (m.activeBtn - 1 + len(m.buttons)) % len(m.buttons)
				m.OpenCurrentMenu()
				return nil
			}

		case tcell.KeyEsc:
			m.closeAllMenus()
			return nil
		}
		return event
	})

	list.SetRect(x, y, listWidth, listHeight)
	m.pages.AddPage(pageName, list, false, true)
	m.app.SetFocus(list)
}

// ****************************************************************************
// closeAllMenus()
// ****************************************************************************
func (m *AppMenuBar) closeAllMenus() {
	// Supprime toutes les pages enregistrées dans la pile
	for _, pageName := range m.menuStack {
		m.pages.RemovePage(pageName)
	}
	m.menuStack = nil
	m.activeList = nil
	m.app.SetFocus(m.mainPrimitive)
}

// ****************************************************************************
// HandleShortcuts()
// ****************************************************************************
func (m *AppMenuBar) HandleShortcuts(event *tcell.EventKey) *tcell.EventKey {
	if action, ok := m.shortcuts[event.Key()]; ok {
		action()
		return nil
	}
	return event
}

// ****************************************************************************
// OpenCurrentMenu()
// ****************************************************************************
func (m *AppMenuBar) OpenCurrentMenu() {
	if m.activeBtn >= 0 && m.activeBtn < len(m.buttons) {
		btn := m.buttons[m.activeBtn]
		title := btn.GetLabel()
		bx, by, _, _ := btn.GetRect()
		if entries, ok := m.menuData[title]; ok {
			m.showDropdown(entries, bx, by+1, "dropdown")
		}
	}
}

// ****************************************************************************
// OpenMenu()
// ****************************************************************************
func (m *AppMenuBar) OpenMenu() {
	btn := m.buttons[0]
	title := btn.GetLabel()
	bx, by, _, _ := btn.GetRect()
	if entries, ok := m.menuData[title]; ok {
		m.showDropdown(entries, bx, by+1, "dropdown")
	}
}

// ****************************************************************************
// ShowMenuPopup() - Affiche un menu au centre de l'écran avec un titre
// ****************************************************************************
func (m *AppMenuBar) ShowMenuPopup(title string, entries []MenuEntry) {
	// 1. Calcul des dimensions
	maxLabelWidth := len(title) + 4
	for _, entry := range entries {
		length := len(entry.Label) + 6
		if length > maxLabelWidth {
			maxLabelWidth = length
		}
	}

	listWidth := maxLabelWidth + 4
	listHeight := len(entries) + 2 // +2 pour la bordure haut/bas

	// 2. Calcul du centre de l'écran
	_, _, screenWidth, screenHeight := m.pages.GetInnerRect()
	x := (screenWidth - listWidth) / 2
	y := (screenHeight - listHeight) / 2

	// 3. Création de la liste
	list := tview.NewList().ShowSecondaryText(false)
	list.SetBorder(true)
	list.SetTitle(" " + title + " ")
	list.SetTitleAlign(tview.AlignCenter)

	// Couleurs (on réutilise votre thème)
	list.SetBackgroundColor(conf.GetColor(config.MenuBgColor))
	list.SetMainTextColor(conf.GetColor(config.MenuTextColor))
	list.SetSelectedBackgroundColor(conf.GetColor(config.MenuSelectedColor))
	list.SetSelectedTextColor(tcell.ColorWhite)
	list.SetBorderColor(conf.GetColor(config.MenuSelectedColor))
	list.SetSelectedFocusOnly(false)

	// Gestion de la pile pour pouvoir fermer avec ESC
	pageName := "popup_" + title
	m.menuStack = append(m.menuStack, pageName)

	for i := range entries {
		e := &entries[i]
		idx := i

		displayLabel := "  " + e.Label
		if len(e.SubEntries) > 0 {
			displayLabel += " >"
		}

		list.AddItem(displayLabel, "", 0, func() {
			if e.Disabled {
				return
			}

			if len(e.SubEntries) > 0 {
				m.showDropdown(e.SubEntries, x+listWidth, y+idx+1, "submenu")
				return
			}

			m.closeAllMenus()
			if e.Action != nil {
				e.Action()
			}
		})
	}

	// Capture clavier pour quitter ou naviguer
	list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			m.closeAllMenus()
			return nil
		}
		return event
	})

	list.SetRect(x, y, listWidth, listHeight)
	m.pages.AddPage(pageName, list, false, true)
	m.app.SetFocus(list)
}
