package main

import (
	"bled/conf"
	"bled/utils"
	"bufio"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/pgavlin/femto"
)

// ****************************************************************************
// VARS
// ****************************************************************************
var (
	Macros map[string]string
)

// ****************************************************************************
// SaveMacros()
// ****************************************************************************
func SaveMacros() {
	SetStatus("Saving macros file")
	fMac, err := os.Create(conf.GetMacrosPath())
	if err == nil {
		defer fMac.Close()
		wMac := bufio.NewWriter(fMac)
		fmt.Fprintln(wMac, "################################################################################")
		fmt.Fprintln(wMac, "# Macros file generated on "+time.Now().Format("20060201-150405"))
		fmt.Fprintln(wMac, "################################################################################")
		fmt.Fprintln(wMac, "# These placeholders can be used into macros :")
		fmt.Fprintln(wMac, "# %D  : Full directory of current file")
		fmt.Fprintln(wMac, "# %P  : Parent directory of current file")
		fmt.Fprintln(wMac, "# %F  : Full file name with directory and extension of current file")
		fmt.Fprintln(wMac, "# %f  : File name without path and with extension of current file")
		fmt.Fprintln(wMac, "# %e  : File name without path nor extension of current file")
		fmt.Fprintln(wMac, "# %L  : Line number of current file in editor")
		fmt.Fprintln(wMac, "# %T  : Current timestamp")
		fmt.Fprintln(wMac, "# %H  : Home directory of current user")
		fmt.Fprintln(wMac, "# %U  : Current user name")
		fmt.Fprintln(wMac, "# %s  : OS path separator")
		fmt.Fprintln(wMac, "# %GU : GitHub user from config file")
		fmt.Fprintln(wMac, "# %GK : GitHub key from config file")
		fmt.Fprintln(wMac, "# %GE : GitHub email from config file")
		fmt.Fprintln(wMac, "################################################################################")
		fmt.Fprintln(wMac, "")
		for k, v := range Macros {
			fmt.Fprintln(wMac, k+" : "+v)
		}
		wMac.Flush()
	}
}

// ****************************************************************************
// ReadMacros()
// ****************************************************************************
func ReadMacros() {
	SetStatus("Reading macros file")
	fMac, err := os.Open(conf.GetMacrosPath())
	if err == nil {
		for k := range Macros {
			delete(Macros, k)
		}
		defer fMac.Close()
		sMac := bufio.NewScanner(fMac)
		for sMac.Scan() {
			if strings.TrimSpace(string(sMac.Text())) != "" {
				if strings.TrimSpace(string(sMac.Text()[0])) != "#" {
					m := strings.Split(sMac.Text(), ":")
					Macros[strings.TrimSpace(m[0])] = strings.TrimSpace(strings.Join(m[1:], ":"))
				}
			}
		}
	}
}

// ****************************************************************************
// XeqMacro()
// ****************************************************************************
func XeqMacro(k any) {
	macroName := k.(string)
	SetStatus("Executing macro : [" + macroName + "]")

	macroContent := Macros[macroName]
	if strings.HasPrefix(macroContent, "insert:") {
		if CurrentFile.ReadOnly {
			SetStatus("Cannot insert snippet : file is read-only")
			return
		}
		snippet := replaceVariablesInMacro(macroName)
		finalSnippet := strings.TrimPrefix(snippet, "insert:")
		finalSnippet = strings.ReplaceAll(finalSnippet, "\\n", "\n")
		finalSnippet = strings.ReplaceAll(finalSnippet, "\\t", "\t")
		finalSnippet = strings.ReplaceAll(finalSnippet, "\\r", "\r")

		startPos := editor.Buf.Cursor.Loc
		cursorOffset := strings.Index(finalSnippet, "$0")
		if cursorOffset != -1 {
			finalSnippet = strings.Replace(finalSnippet, "$0", "", 1)
		}

		pos := editor.Buf.Cursor.Loc
		editor.Buf.Insert(pos, finalSnippet)
		if cursorOffset != -1 {
			newLoc := calculateNewLocation(finalSnippet[:cursorOffset], startPos)
			editor.Buf.Cursor.Loc = newLoc
		}
		editor.Buf.Cursor.Relocate()
		SetStatus("Snippet inserted")
		return
	}

	if CurrentFile.FemtoBuffer.Modified() {
		saveFile()
	}

	infoBefore, errBefore := os.Stat(CurrentFile.FName)

	cmdString := replaceVariablesInMacro(macroName)
	output := XeqOutErr(cmdString)

	infoAfter, errAfter := os.Stat(CurrentFile.FName)

	if errBefore == nil && errAfter == nil {
		if infoAfter.ModTime().After(infoBefore.ModTime()) {
			SetStatus("Refreshing document since it has been modified")
			CurrentFile.Reload()
			editor.Buf = CurrentFile.FemtoBuffer
		}
	}
	out := fmt.Sprintf("%s\n", output)
	MsgBox = MsgBox.OK("Macro : "+macroName, out, nil, 0, "main", editor)
	pages.AddPage("msgBox", MsgBox.Popup(), true, false)
	pages.ShowPage("msgBox")
}

func calculateNewLocation(textBeforeCursor string, start femto.Loc) femto.Loc {
	lines := strings.Split(textBeforeCursor, "\n")
	numLines := len(lines) - 1

	newLoc := start
	newLoc.Y += numLines

	if numLines > 0 {
		// Si on a changé de ligne, la colonne repart de l'index du dernier fragment
		newLoc.X = len(lines[numLines])
	} else {
		// Si on est sur la même ligne, on ajoute simplement le nombre de caractères
		newLoc.X += len(lines[0])
	}

	return newLoc
}

// ****************************************************************************
// replaceVariablesInMacro()
// ****************************************************************************
func replaceVariablesInMacro(k string) string {
	// %D  : Full directory of current file
	// %P  : Parent directory of current file
	// %F  : Full file name with directory and extension of current file
	// %f  : File name without path and with extension of current file
	// %e  : File name without path nor extension of current file
	// %L  : Line number of current file in editor
	// %T  : Current timestamp
	// %H  : Home directory of current user
	// %U  : Current user name
	// %s  : OS path separator
	// %GU : GitHub user from config file
	// %GK : GitHub key from config file
	// %GE : GitHub email from config file
	out := Macros[k]
	userDir, _ := os.UserHomeDir()
	r := strings.NewReplacer(
		"%D", utils.EscapeSpaces(filepath.Dir(CurrentFile.FName)),
		"%P", utils.EscapeSpaces(filepath.Base(filepath.Dir(CurrentFile.FName))),
		"%F", utils.EscapeSpaces(CurrentFile.FName),
		"%f", utils.EscapeSpaces(filepath.Base(CurrentFile.FName)),
		"%e", utils.EscapeSpaces(filepath.Base(strings.TrimSuffix(filepath.Base(CurrentFile.FName), filepath.Ext(CurrentFile.FName)))),
		"%T", time.Now().Format("20060102-150405"),
		"%L", strconv.Itoa(CurrentFile.FemtoBuffer.Cursor.Y+1),
		"%s", string(os.PathSeparator),
		"%H", userDir,
		"%U", os.Getenv("USER"),
		"%GU", config.GithubUser,
		"%GK", config.GithubKey,
		"%GE", config.GithubEmail,
	)
	out = r.Replace(out)
	return out
}

// ****************************************************************************
// replaceVariablesInTemplate()
// ****************************************************************************
func replaceVariablesInTemplate(template string) string {
	// %D  : Full directory of current file
	// %P  : Parent directory of current file
	// %F  : Full file name with directory and extension of current file
	// %f  : File name without path and with extension of current file
	// %e  : File name without path nor extension of current file
	// %L  : Line number of current file in editor
	// %T  : Current timestamp
	// %H  : Home directory of current user
	// %U  : Current user name
	// %s  : OS path separator
	// %GU : GitHub user from config file
	// %GK : GitHub key from config file
	// %GE : GitHub email from config file

	out := template
	userDir, _ := os.UserHomeDir()
	r := strings.NewReplacer(
		"%D", utils.EscapeSpaces(filepath.Dir(CurrentFile.FName)),
		"%P", utils.EscapeSpaces(filepath.Base(filepath.Dir(CurrentFile.FName))),
		"%F", utils.EscapeSpaces(CurrentFile.FName),
		"%f", utils.EscapeSpaces(filepath.Base(CurrentFile.FName)),
		"%e", utils.EscapeSpaces(filepath.Base(strings.TrimSuffix(filepath.Base(CurrentFile.FName), filepath.Ext(CurrentFile.FName)))),
		"%T", time.Now().Format("20060102-150405"),
		"%L", strconv.Itoa(CurrentFile.FemtoBuffer.Cursor.Y+1),
		"%s", string(os.PathSeparator),
		"%H", userDir,
		"%U", os.Getenv("USER"),
		"%GU", config.GithubUser,
		"%GK", config.GithubKey,
		"%GE", config.GithubEmail,
	)
	out = r.Replace(out)
	return out
}

// ****************************************************************************
// editMacrosFile()
// ****************************************************************************
func editMacrosFile() {
	SetStatus("Editing macros file")
	openFile(conf.GetMacrosPath(), false)
}

// ****************************************************************************
// refreshMacrosMenu()
// ****************************************************************************
func refreshMacrosMenu() {
	ReadMacros()
	macroEntries = nil
	snippetEntries = nil
	commandEntries = nil
	sortedMacros := slices.Sorted(maps.Keys(Macros))
	for _, m := range sortedMacros {
		mName, mContent := m, Macros[m]

		entry := MenuEntry{
			Label:  mName,
			Action: func() { XeqMacro(mName) },
		}

		if strings.HasPrefix(mContent, "insert:") {
			snippetEntries = append(snippetEntries, entry)
		} else {
			commandEntries = append(commandEntries, entry)
		}
	}
	macroEntries = append(macroEntries, MenuEntry{
		Label:      "Run Command",
		SubEntries: commandEntries,
	})
	macroEntries = append(macroEntries, MenuEntry{
		Label:      "Insert Snippet",
		SubEntries: snippetEntries,
	})
	macroEntries = append(macroEntries, MenuEntry{IsSeparator: true})
	macroEntries = append(macroEntries, MenuEntry{Label: "Edit macros file", Action: func() { editMacrosFile() }})
}

// ****************************************************************************
// GetSnippets()
// ****************************************************************************
func GetSnippets() map[string]string {
	snippets := make(map[string]string)
	// On suppose que tes macros sont stockées dans conf.Macros (map[string]string)
	for name, content := range Macros {
		if strings.HasPrefix(content, "insert:") {
			snippets[name] = content
		}
	}
	return snippets
}
