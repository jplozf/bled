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

### License
This project is licensed under the terms of the [MIT license](https://www.mit.edu/~amini/LICENSE.md).

---

## ⌨️ Keybindings


### Function keys
| Key | Action |
| :--- | :--- |
| **F1** | This help text |
| **F3** | Jump to the **Git Tracking** menu |
| **F4** | Jump to the next match of the current search query |
| **F5** | Jump to the **Macros** menu |
| **F6** | Switch to the previous open file |
| **F7** | Switch to the next open file |
| **F10** | Access to the main menu of Bled |

### Common Functions
* `CTRL + F` : Switch to the **Find & Replace** panel
* `CTRL + K` : Toggle the **File Info** panel
* `CTRL + L` : Access to the main menu of Bled (same as `F10`)
* `CTRL + S` : Saves the current document being edited
* `ALT  + S` : Saves the current document being edited under another name
* `CTRL + N` : Opens a new blank document
* `CTRL + O` : Opens an existing document for editing
* `CTRL + T` : Closes the current document
* `CTRL + G` : Opens the **Goto** panel to jump to a specific part of the document
* `CTRL + Q` : Quit Bled
* `ALT  + Q` : Temporarily return to the shell

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
* **Colors :** `menu_bg_color`, `menu_selected_color`,
* **Interface :** `show_welcome_popup`, `confirm_on_quit`, `show_hidden_files`, `status_bar_separator`.
* **Editor :** `theme`, `max_recent_files`.
* **GitHub :** `github_user`, `github_key`, `github_email`.

### Example `config.json`
```json
{
    "menu_bg_color": "#4682B4",
    "menu_selected_color": "#A0522D",
    "show_welcome_popup": true,
    "confirm_on_quit": true,
    "theme": "monokai",
    "status_bar_separator": "*"
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
* `%GU`: GitHub user from config file
* `%GK`: GitHub key from config file
* `%GE`: GitHub email from config file

Macros syntax is simple, each line in the macros file represents a macro, with the following format :

`<Macro Name> : <Command to execute>`

For example, a macro to open the current file in the default system editor could be defined as follows :

`Open in Explorer : xdg-open %D`

This macro will use the `xdg-open` command to open the current file's directory (represented by the `%D` placeholder) in the default system explorer when executed.

If the command to execute starts with `insert:`, the rest of the command will be inserted into the current document at the current cursor position instead of being executed as a system command.

For example, a macro to insert the current date and time into the document could be defined as follows :

`Insert Timestamp : insert:%T`

This macro will insert the current timestamp (represented by the `%T` placeholder) into the document at the current cursor position when executed.

You can create as many macros as you want, and they will be available in the **Macros** menu for easy access.

## 📁 Templates
* Templates for new files are located into the `~/.bled/templates` directory. 
* You can add or remove templates as you wish. 
* These templates could be structured into sub-folders and rendered as sub-menus in the **Templates** menu. 
* These templates can accept the same placeholders as macros.

## 📦 Git status

When a file is tracked by **Git**, the following informations are displayed in the right top corner of the editor in the following format :

`[current file status in Git]:[current commit hash]:[global Git status]:[current branch name]`

### Exemple :

`M:37d6c0e:Pending:main`

Which means that :
* the current file has been modified locally `M`
* the current commit hash is `37d6c0e` 
* there are `pending` changes in this Git repository
* and the current branch name is `main`.

---

## 📁 Internal Files

Files used internally to manage the behavior of **Bled** are located into the `~/.bled` directory. They look like something like this :


```bash
.
├── bled.log
├── bled.log.bak
├── bled_202603.zip
├── bled_202602.zip
├── bled_202601.zip
├── config.json
├── macros
├── recents
└── templates
    ├── Assembly Source Code 32 bits.asm
    ├── Assembly Source Code 64 bits.asm
    ├── Bash Script
    ├── C
    │   ├── C GTK Source Code.c
    │   ├── C Ncurses Source Code.c
    │   ├── C Source Code.c
    │   └── C++ wxWidgets Source Code.cpp
    ├── Go Source Code.go
    ├── HTML Page.html
    ├── Java Applet.java
    ├── Java Source Code.java
    ├── Java Swing Application.java
    ├── Licenses
    │   ├── apache-v2.0.md
    │   ├── artistic-v2.0.md
    │   ├── bsd-2.md
    │   ├── bsd-3.md
    │   ├── epl-v1.0.md
    │   ├── gnu-agpl-v3.0.md
    │   ├── gnu-fdl-v1.3.md
    │   ├── gnu-gpl-v1.0.md
    │   ├── gnu-gpl-v2.0.md
    │   ├── gnu-gpl-v3.0.md
    │   ├── gnu-lgpl-v2.1.md
    │   ├── gnu-lgpl-v3.0.md
    │   ├── mit.md
    │   ├── mpl-v2.0.md
    │   └── unlicense.md
    ├── Makefile
    ├── Markdown File.md
    ├── Python
    │   ├── Python Script.py
    │   └── Python TkInter.py
    ├── QT MainWindow.ui
    ├── Rust Source Code.rs
    └── XML File.xml
```
Where :

| File | Purpose |
| :--- | :--- |
| `bled.log` | Main log file |
| `bled.log.bak` | Backup log file *(used for archiving)* |
| `bled_YYYYMM.zip` | Archived log file by month |
| `config.json` | Configuration file *(edited from menu **Settings**)* |
| `macros` | Macros file *(edited from menu **Macros**)* |
| `recents` | List of recently opened files |
| `templates` | Folder for templates *(empty by default, use your own templates)* |
