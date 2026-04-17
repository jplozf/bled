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
		fmt.Fprintln(wMac, "# %D : Full directory of current file")
		fmt.Fprintln(wMac, "# %P : Parent directory of current file")
		fmt.Fprintln(wMac, "# %F : Full file name with directory and extension of current file")
		fmt.Fprintln(wMac, "# %f : File name without path and with extension of current file")
		fmt.Fprintln(wMac, "# %e : File name without path nor extension of current file")
		fmt.Fprintln(wMac, "# %L : Line number of current file in editor")
		fmt.Fprintln(wMac, "# %T : Current timestamp")
		fmt.Fprintln(wMac, "# %H : Home directory of current user")
		fmt.Fprintln(wMac, "# %U : Current user name")
		fmt.Fprintln(wMac, "# %s : OS path separator")
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
	if CurrentFile.FemtoBuffer.Modified() {
		saveFile()
	}

	macroName := k.(string)
	SetStatus("Executing macro : [" + macroName + "]")

	infoBefore, errBefore := os.Stat(CurrentFile.FName)

	cmdString := replaceVariablesInMacro(macroName)
	output := XeqOutErr(cmdString)

	infoAfter, errAfter := os.Stat(CurrentFile.FName)

	if errBefore == nil && errAfter == nil {
		if infoAfter.ModTime().After(infoBefore.ModTime()) {
			refreshDocument()
		}
	}
	out := fmt.Sprintf("%s\n", output)
	MsgBox = MsgBox.OK("Macro : "+macroName, out, nil, 0, "main", editor)
	pages.AddPage("msgBox", MsgBox.Popup(), true, false)
	pages.ShowPage("msgBox")
}

// ****************************************************************************
// replaceVariablesInMacro()
// ****************************************************************************
func replaceVariablesInMacro(k string) string {
	// %D : Full directory of current file
	// %P : Parent directory of current file
	// %F : Full file name with directory and extension of current file
	// %f : File name without path and with extension of current file
	// %e : File name without path nor extension of current file
	// %L : Line number of current file in editor
	// %T : Current timestamp
	// %H : Home directory of current user
	// %U : Current user name
	// %s : OS path separator
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
	sortedMacros := slices.Sorted(maps.Keys(Macros))
	for _, k := range sortedMacros {
		macroEntries = append(macroEntries, MenuEntry{Label: k, Action: func() { XeqMacro(k) }})
	}
	macroEntries = append(macroEntries, MenuEntry{IsSeparator: true})
	macroEntries = append(macroEntries, MenuEntry{Label: "Edit macros file", Action: func() { editMacrosFile() }})
}
