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
F4  : Jump to the next match of the current search query
F5  : Jump to the "Macros" menu
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

⯈ The "Goto" panel, reachable via CTRL + G, supports the following functions :

<line>             : Jump to a specific line number in the current document
<percentage%>      : Jump to a specific percentage of the document (e.g. 50% will jump to the middle of the document)
<+line> or <-line> : Jump to a specific line number relative to the current line (e.g. +10 will jump 10 lines down, -5 will jump 5 lines up) 
top                : Jump to the top of the document
start              : Jump to the start of the document (same as top)
bottom             : Jump to the bottom of the document
end                : Jump to the end of the document (same as bottom)
follow             : Enable "Follow Mode", which will keep the current document continuously scrolled to the end as new lines are added (useful for log files)
$                  : Same as "follow"

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
github_email            : Email used for Github commits

Here is the default configuration file content, which is generated at first launch if not already present in the user's home directory :	
{                                                                                                                                
    "menu_bg_color": "#4682B4",                                                                                                  
    "menu_selected_color": "#A0522D",                                                                                            
    "menu_text_color": "black",                                                                                                  
    "menu_disabled_color": "gray",                                                                                               
    "show_welcome_popup": true,                                                                                                 
    "confirm_on_quit": true,                                                                                                     
    "theme": "monokai"                                                                               
}                                                                                                                                
Don't forget to edit the "github_user", "github_key" and "github_email" fields if you want to use the GitHub integration features of Bled.


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

⯈ Macros are also supported in Bled, and can be used to access and launch common pre-defined tasks :

Macros are accessible through the "Macros" menu, reachable also directly via F5 key.
Macros are stored in a separate file, as a simple text file located in the user's home directory.
The macros file is generated at first launch if not already present, and contains some placeholders that can be used to create powerful macros.
Here are the available placeholders that can be used in macros :

%D : Full directory of current file
%P : Parent directory of current file
%F : Full file name with directory and extension of current file
%f : File name without path and with extension of current file
%e : File name without path nor extension of current file
%L : Line number of current file in editor
%T : Current timestamp
%H : Home directory of current user
%s : OS path separator

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
