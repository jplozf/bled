package main

// ****************************************************************************
// H E L P   T E X T
// ****************************************************************************
var helpText = `--------------------------------------------------------------------------
⚶ B L E D   -   Copyright © JPL 2026   -   https://github.com/jplozf/bled
--------------------------------------------------------------------------

Bled is a Text User Interface (TUI) Editor.
Bled is written in Go and has been tested on Linux sytem.
Built from source, it should run on Windows or MacOS systems as well.
Multiple files can be opened at the same time, and you can switch between them easily.
Common Git commands are implemented.

Pay honour to whom honour is due, packages used in this project are as follows :
rivo/tview    : Package tview implements rich widgets for terminal based user interfaces.
gdamore/tcell : Tcell is an alternate terminal package, similar in some ways to termbox, but better in others. 
pgavlin/femto : An editor component for tview. Derived from the micro editor. 

⯈ The main functions are reachable through function keys :

F1  : This help text
F3  : Jump to the "Git Tracking" menu
F5  : Jump to the "Find & Replace" panel
F4  : Jump to the next match of the current search query
F6  : Switch to the previous open file
F7  : Switch to the next open file
F10 : Access to the main menu of Bled

⯈ Alternate common functions are also reachable through CTRL and ALT keys :

CTRL + F : Switch to the "Find & Replace" panel
CTRL + S : Saves the current document being edited
ALT  + S : Saves the current document being edited under another name
CTRL + N : Opens a new blank document
CTRL + O : Opens an existing document for editing
CTRL + T : Closes the current document
CTRL + G : Opens the "Goto" panel to jump to a specific part of the document
CTRL + Q : Quit Bled

⯈ When editing a text, common editing functions are of course supported :

CTRL + C : Copy the selection
CTRL + X : Cut the selection
CTRL + V : Paste the selection
CTRL + Z : Cancels the previous entry 
CTRL + Y : Redo the previous cancelled operation

⯈ Settings are stored in a configuration file, as a JSON file located in the user's home directory :

menu_bg_color           : Background color of the menu bar
menu_selected_color     : Background color of the selected menu item
menu_text_color         : Color of the text in the menu
menu_disabled_color     : Color of disabled menu items
show_welcome_popup      : Whether to show a welcome popup at startup
confirm_on_quit         : Whether to ask for confirmation when quitting with unsaved changes
status_message_duration : Duration in seconds for which status messages are displayed
show_hidden_files       : Whether to show hidden files in the file browser dialog
color_accent            : Accent color used in the UI (e.g. for highlights)
theme                   : Theme name (see below for available themes)
github_user             : User name used for Github commits
github_key              : Key used for Github operations

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
