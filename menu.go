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
			length += 2
		} // Espace pour "[x] "
		if len(entry.SubEntries) > 0 {
			length += 2
		}
		if length > maxLabelWidth {
			maxLabelWidth = length
		}
	}
	listWidth := maxLabelWidth + 2

	if listWidth < 15 { // Largeur minimale pour que ce ne soit pas trop étroit
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

	// 1. LE FOND GLOBAL DU COMPOSANT
	list.SetBackgroundColor(conf.GetColor(config.MenuBgColor))

	// 2. LE TEXTE PAR DÉFAUT (Non sélectionné)
	// On force le noir sur jaune pour TOUS les items
	list.SetMainTextColor(conf.GetColor(config.MenuTextColor))

	// 3. L'ITEM SÉLECTIONNÉ (Votre barre rouge)
	list.SetSelectedBackgroundColor(conf.GetColor(config.MenuSelectedColor))
	list.SetSelectedTextColor(tcell.ColorWhite)

	// 4. TRÈS IMPORTANT : Désactiver le style "Focus Only"
	// Cela force la liste à utiliser vos couleurs même si le focus bégaye
	list.SetSelectedFocusOnly(false)

	// Si c'est le menu principal (dropdown), on le mémorise
	if pageName == "dropdown" {
		m.activeList = list
	}

	for i := range entries {
		e := &entries[i]
		label := e.Label

		prefix := "  " // Un peu d'espace à gauche
		if e.IsCheckable {
			if e.Checked {
				prefix = "[\u221A] "
			} else {
				prefix = "[ ] "
			}
		}

		// On prépare le label SANS balise de fond (:)
		// On ajoute juste des espaces pour que la barre de sélection (rouge)
		// fasse toute la largeur du menu.
		padding := strings.Repeat(" ", maxLabelWidth+1-len(label)-len(prefix))
		displayLabel := prefix + label + padding

		if e.Disabled {
			displayLabel = fmt.Sprintf("[%s]%s[%s]", conf.GetColor(config.MenuDisabledColor), displayLabel, conf.GetColor(config.MenuTextColor))
		} else if len(e.SubEntries) > 0 {
			// On remplace le dernier espace par la flèche
			displayLabel = displayLabel[:len(displayLabel)-1] + ">"
		}

		list.AddItem(displayLabel, "", 0, func() {
			if e.Disabled {
				return
			}

			// Si c'est coché, on inverse l'état
			if e.IsCheckable {
				e.Checked = !e.Checked
				// Note : On ne ferme pas le menu pour que l'utilisateur
				// voie le changement, ou on le ferme selon votre choix UX.
				m.closeAllMenus()
			} else if len(e.SubEntries) == 0 {
				m.closeAllMenus()
			}

			if e.Action != nil {
				e.Action()
			}
		})
	}

	// --- LA CORRECTION RADICALE ---
	// On définit la zone de dessin de la liste manuellement
	list.SetRect(x, y, listWidth, listHeight)

	list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// --- NAVIGATION DANS LE SOUS-MENU ---
		if pageName == "submenu" {
			if event.Key() == tcell.KeyLeft {
				m.pages.RemovePage("submenu")
				m.app.SetFocus(m.activeList) // Retour au parent direct
				return nil
			}

			// Si on veut passer à l'item SUIVANT du parent avec Droite/Gauche
			if event.Key() == tcell.KeyRight || event.Key() == tcell.KeyLeft {
				m.pages.RemovePage("submenu")

				// Déplacer la sélection dans la liste parente
				curr := m.activeList.GetCurrentItem()
				count := m.activeList.GetItemCount()
				if event.Key() == tcell.KeyRight {
					m.activeList.SetCurrentItem((curr + 1) % count)
				} else {
					m.activeList.SetCurrentItem((curr - 1 + count) % count)
				}

				// Simuler un appui sur Droite pour ouvrir le nouveau sous-menu
				// On peut soit appeler récursivement, soit redonner le focus
				m.app.SetFocus(m.activeList)
				// Optionnel : Vous pouvez ici tester si le nouvel item a un sous-menu et l'ouvrir
				return nil
			}
		}

		// --- NAVIGATION DANS LE DROPDOWN PRINCIPAL ---
		if pageName == "dropdown" {
			// Si on appuie sur Droite sur un item SANS sous-menu
			// ou sur Gauche n'importe quand -> On change de menu (Fichier -> Edition)
			idx := list.GetCurrentItem()
			if (event.Key() == tcell.KeyRight && len(entries[idx].SubEntries) == 0) || event.Key() == tcell.KeyLeft {
				m.closeAllMenus()
				if event.Key() == tcell.KeyRight {
					m.activeBtn = (m.activeBtn + 1) % len(m.buttons)
				} else {
					m.activeBtn = (m.activeBtn - 1 + len(m.buttons)) % len(m.buttons)
				}
				m.OpenCurrentMenu()
				return nil
			}
		}

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

// ****************************************************************************
// closeAllMenus()
// ****************************************************************************
func (m *AppMenuBar) closeAllMenus() {
	m.pages.RemovePage("dropdown")
	m.pages.RemovePage("submenu")
	m.activeList = nil
	m.app.SetFocus(m.Flex)
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
