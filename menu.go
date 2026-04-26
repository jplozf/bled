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
	IsSeparator bool
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

	// Horizontal navigation between menu buttons with Left/Right keys
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

	// Version initialization from the start
	// versionStr := fmt.Sprintf("%s %s %s", conf.APP_ICON, conf.APP_NAME, getFullVersion())
	m.versionView = tview.NewTextView().
		SetTextAlign(tview.AlignRight).
		SetDynamicColors(true).
		SetText(GITInfos + " ")

	m.versionView.SetBackgroundColor(conf.GetColor(conf.LoadConfig().MenuBgColor))
	m.versionView.SetTextColor(TextColor)

	return m
}

// ****************************************************************************
// UpdateMenu()
// ****************************************************************************
func (m *AppMenuBar) UpdateMenu(title string, entries []MenuEntry) {
	if m.menuData == nil {
		m.menuData = make(map[string][]MenuEntry)
	}
	// We only update the data
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

	// --- Otherwise, it's the initial button creation ---
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
	// If the button doesn't exist, we create it with the appropriate styles and behavior
	if !buttonExists {
		btn := tview.NewButton(title)

		// We explicitly define styles for EACH state
		styleNormal := tcell.StyleDefault.Background(conf.GetColor(conf.LoadConfig().MenuBgColor)).Foreground(TextColor)
		styleFocus := tcell.StyleDefault.Background(conf.GetColor(conf.LoadConfig().MenuSelectedColor)).Foreground(SelectedTextColor)

		// Idle style
		btn.SetStyle(styleNormal)
		btn.SetLabelColor(TextColor)
		btn.SetBackgroundColor(conf.GetColor(conf.LoadConfig().MenuBgColor))

		// Style when the button is selected
		btn.SetActivatedStyle(styleFocus)
		btn.SetBackgroundColorActivated(conf.GetColor(conf.LoadConfig().MenuSelectedColor))
		btn.SetLabelColorActivated(SelectedTextColor)

		// Idle style
		btn.SetBackgroundColor(conf.GetColor(conf.LoadConfig().MenuBgColor))
		btn.SetLabelColor(TextColor)

		// Style when navigating over it
		btn.SetBackgroundColorActivated(conf.GetColor(conf.LoadConfig().MenuSelectedColor))
		btn.SetLabelColorActivated(SelectedTextColor)

		btn.SetSelectedFunc(func() {
			bx, by, _, _ := btn.GetRect()
			m.showDropdown(entries, bx, by+1, "dropdown")
		})

		btn.SetSelectedFunc(func() {
			bx, by, _, _ := btn.GetRect()
			// We retrieve dynamic data on click
			m.showDropdown(m.menuData[title], bx, by+1, "dropdown")
		})

		m.buttons = append(m.buttons, btn)
		m.rebuildBar() // We redraw to include the new button
	}

	return m
}

// ****************************************************************************
// rebuildBar()
// ****************************************************************************
func (m *AppMenuBar) rebuildBar() {
	m.Clear()

	for _, btn := range m.buttons {
		width := len(btn.GetLabel()) + 2
		m.AddItem(btn, width, 0, false)
	}

	if m.spacer == nil {
		m.spacer = tview.NewBox().SetBackgroundColor(conf.GetColor(config.MenuBgColor))
	}
	m.AddItem(m.spacer, 0, 1, false)

	if m.versionView != nil {
		m.versionView.SetText(GITInfos + " ")
		versionText := m.versionView.GetText(false)
		m.AddItem(m.versionView, len(versionText)+1, 0, false)
	}
}

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
	// Dynamic width calculation
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

	// Screen edge detection
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
	list.SetMainTextColor(TextColor)
	list.SetSelectedBackgroundColor(conf.GetColor(config.MenuSelectedColor))
	list.SetSelectedTextColor(tcell.ColorWhite)
	list.SetSelectedFocusOnly(false)

	// If it's the first menu (main dropdown), we clear the stack
	if pageName == "dropdown" {
		m.menuStack = []string{pageName}
		m.activeList = list
	} else {
		// Otherwise, we add this sub-menu to the stack
		m.menuStack = append(m.menuStack, pageName)
	}

	for i := range entries {
		e := &entries[i]
		idx := i // Capture the index for the closure

		if e.IsSeparator {
			sepLine := " " + strings.Repeat("-", listWidth-2) + " "
			list.AddItem(sepLine, "", 0, nil)
			continue
		}

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
			displayLabel = fmt.Sprintf("[%s]%s[%s]", conf.GetColor("gray"), displayLabel, TextColor)
		} else if len(e.SubEntries) > 0 {
			displayLabel += " ⯈"
		}

		list.AddItem(displayLabel, "", 0, func() {
			if e.Disabled {
				return
			}

			// --- CLICK SUB-MENU MANAGEMENT ---
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
		case tcell.KeyDown:
			nextIdx := currIdx + 1
			if nextIdx < list.GetItemCount() && entries[nextIdx].IsSeparator {
				if nextIdx+1 < list.GetItemCount() {
					list.SetCurrentItem(nextIdx + 1)
					return nil
				}
			}

		case tcell.KeyUp:
			prevIdx := currIdx - 1
			if prevIdx >= 0 && entries[prevIdx].IsSeparator {
				if prevIdx-1 >= 0 {
					list.SetCurrentItem(prevIdx - 1)
					return nil
				}
			}

		case tcell.KeyRight:
			// Open the sub-menu if present
			if len(entry.SubEntries) > 0 {
				subPageName := fmt.Sprintf("submenu_%d", len(m.menuStack))
				m.showDropdown(entry.SubEntries, x+listWidth, y+currIdx, subPageName)
				return nil
			}
			// If we are at level 0 and no sub-menu, we change the main menu
			if pageName == "dropdown" {
				m.closeAllMenus()
				m.activeBtn = (m.activeBtn + 1) % len(m.buttons)
				m.OpenCurrentMenu()
				return nil
			}

		case tcell.KeyLeft:
			// If it's a sub-menu, we go back
			if len(m.menuStack) > 1 {
				lastPage := m.menuStack[len(m.menuStack)-1]
				m.menuStack = m.menuStack[:len(m.menuStack)-1]
				m.pages.RemovePage(lastPage)
				// Focus automatically returns to the previous visible page
				return nil
			}
			// If we are at level 0, we change the main menu
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
	// Removes all pages saved in the stack
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
	// Dimensions calculation
	maxLabelWidth := len(title) + 4
	for _, entry := range entries {
		length := len(entry.Label) + 6
		if length > maxLabelWidth {
			maxLabelWidth = length
		}
	}

	listWidth := maxLabelWidth + 4
	listHeight := len(entries) + 2 // +2 for top/bottom border

	// Screen center calculation
	_, _, screenWidth, screenHeight := m.pages.GetInnerRect()
	x := (screenWidth - listWidth) / 2
	y := (screenHeight - listHeight) / 2

	// List creation
	list := tview.NewList().ShowSecondaryText(false)
	list.SetBorder(true)
	list.SetTitle(" " + title + " ")
	list.SetTitleAlign(tview.AlignCenter)

	// Colors (we reuse your theme)
	list.SetBackgroundColor(conf.GetColor(config.MenuBgColor))
	list.SetMainTextColor(TextColor)
	list.SetSelectedBackgroundColor(conf.GetColor(config.MenuSelectedColor))
	list.SetSelectedTextColor(tcell.ColorWhite)
	list.SetBorderColor(conf.GetColor(config.MenuSelectedColor))
	list.SetSelectedFocusOnly(false)

	// Stack management to be able to close with ESC
	pageName := "popup_" + title
	m.menuStack = append(m.menuStack, pageName)

	for i := range entries {
		e := &entries[i]
		idx := i

		if e.IsSeparator {
			sepLine := " " + strings.Repeat("-", listWidth-2) + " "
			list.AddItem(sepLine, "", 0, nil)
			continue
		}

		displayLabel := "  " + e.Label + strings.Repeat(" ", maxLabelWidth-len(e.Label))
		if len(e.SubEntries) > 0 {
			displayLabel += " ⯈"
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

	// Keyboard capture to quit or navigate
	list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		currIdx := list.GetCurrentItem()
		switch event.Key() {
		case tcell.KeyDown:
			nextIdx := currIdx + 1
			if nextIdx < list.GetItemCount() && entries[nextIdx].IsSeparator {
				if nextIdx+1 < list.GetItemCount() {
					list.SetCurrentItem(nextIdx + 1)
					return nil
				}
			}

		case tcell.KeyUp:
			prevIdx := currIdx - 1
			if prevIdx >= 0 && entries[prevIdx].IsSeparator {
				if prevIdx-1 >= 0 {
					list.SetCurrentItem(prevIdx - 1)
					return nil
				}
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
