# ⚶ B L E D

**Text User Interface (TUI) Editor** *Copyright © JPL 2026* — [GitHub Repository](https://github.com/jplozf/bled)

---

## 📝 Presentation
**Bled** is a Text User Interface (TUI) Editor.
* Bled is written in **Go** and has been tested on Linux sytem.
* Built from source, it should run on Windows or MacOS systems as well.
* Multiple files can be opened at the same time, and you can switch between them easily.
* Common **Git** commands are implemented.

### Credits
Pay honour to whom honour is due, packages used in this project are as follows :

* `rivo/tview` : Package tview implements rich widgets for terminal based user interfaces.
* `gdamore/tcell` : Tcell is an alternate terminal package, similar in some ways to termbox, but better in others. 
* `pgavlin/femto` : An editor component for tview. Derived from the *micro* editor. 

---

## ⌨️ Keybindings


### Function keys
| Key | Action |
| :--- | :--- |
| **F1** | This help text |
| **F3** | Jump to the "Git Tracking" menu |
| **F4** | Jump to the next match of the current search query |
| **F5** | Jump to the "Macros" menu |
| **F6** | Switch to the previous open file |
| **F7** | Switch to the next open file |
| **F10** | Access to the main menu of Bled |

### Common Functions
* `CTRL + F` : Switch to the "Find & Replace" panel
* `CTRL + S` : Saves the current document being edited
* `ALT  + S` : Saves the current document being edited under another name
* `CTRL + N` : Opens a new blank document
* `CTRL + O` : Opens an existing document for editing
* `CTRL + T` : Closes the current document
* `CTRL + G` : Opens the "Goto" panel to jump to a specific part of the document
* `CTRL + Q` : Quit Bled

### Text Editing
* `CTRL + C` / `X` / `V` : Copy / Cut / Paste
* `CTRL + Z` : Cancels the previous entry (Undo)
* `CTRL + Y` : Redo the previous cancelled operation

---

## 🚀 "Goto" Panel (`CTRL + G`)
Navigating easily in the document :
* `<number>` : Jump to a specific line number in the current document
* `<n%>` : Jump to a specific percentage of the document (e.g. `50%` will jump to the middle of the document)
* `+n` / `-n` : Jump to a specific line number relative to the current line (e.g. `+10` will jump 10 lines down, `-5` will jump 5 lines up) 
* `top` / `start` : Jump to the beginning of the document
* `bottom` / `end` : Jump to the end of the document
* `follow` / `$` : Enable **Follow Mode** same as `tail -f`, which will keep the current document continuously scrolled to the end as new lines are added (useful for log files)

---

## ⚙️ Settings
Settings are stored in a configuration file, as a **JSON** file located in the user's home directory : `~/.bled/config.json`.

### Available Settings
* **Colors :** `menu_bg_color`, `menu_selected_color`, `menu_text_color`, `menu_disabled_color`.
* **Interface :** `show_welcome_popup`, `confirm_on_quit`, `status_message_duration`, `show_hidden_files`, `color_accent`.
* **Editor :** `theme`, `max_recent_files`.
* **GitHub :** `github_user`, `github_key`, `github_email`.

### Example `config.json`
```json
{
    "menu_bg_color": "#4682B4",
    "menu_selected_color": "#A0522D",
    "menu_text_color": "black",
    "menu_disabled_color": "gray",
    "show_welcome_popup": true,
    "confirm_on_quit": true,
    "theme": "monokai"
}
```

## 🎨 Availables Themes
These themes are taken from the `pgavlin\femto` package.
Theme's names are case sensitive.
  
* `atom-dark-tc`
* `bubblegum`
* `cmc-16`
* `cmc-paper`
* `cmc-tc`
* `darcula`
* `default`
* `geany`
* `github-tc`
* `gruvbox-tc`
* `gruvbox`
* `material-tc`
* `monokai`
* `railscast`
* `simple`
* `solarized-tc`
* `solarized`
* `twilight`
* `zenburn`

## 🛠️ Macros
Macros are also supported in Bled, and can be used to access and launch common pre-defined tasks :
* Macros are accessible through the **Macros** menu, reachable also directly via `F5` key.
* Macros are stored in a separate file `~/.bled/macros`, as a simple text file located in the user's home directory.
* The macros file is generated at first launch if not already present, and contains some placeholders that can be used to create powerful macros.

Here are the available placeholders that can be used as variables in macros :
* `%D` : Full directory of current file
* `%P` : Parent directory of current file
* `%F` : Full file name with directory and extension of current file
* `%f` : File name without path and with extension of current file
* `%e` : File name without path nor extension of current file
* `%L` : Line number of current file in editor
* `%T` : Current timestamp
* `%H` : Home directory of current user
* `%U` : Current user name
* `%s` : OS path separator

Macros syntax is simple, each line in the macros file represents a macro, with the following format :

`<Macro Name> : <Command to execute>`

For example, a macro to open the current file in the default system editor could be defined as follows :

`Open in Explorer : xdg-open %D`

This macro will use the `xdg-open` command to open the current file's directory (represented by the `%D` placeholder) in the default system explorer when executed.

You can create as many macros as you want, and they will be available in the **Macros** menu for easy access.
