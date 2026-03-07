package main

// ****************************************************************************
// IMPORTS
// ****************************************************************************
import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/gdamore/tcell/v2"
	"github.com/pgavlin/femto"
	"github.com/rivo/tview"
)

// ****************************************************************************
// TYPES
// ****************************************************************************
type efile struct {
	FemtoBuffer *femto.Buffer
	FemtoView   *femto.View
	FName       string
	Encoding    string
	Modified    bool
}

// ****************************************************************************
// VARS
// ****************************************************************************
var (
	efiles      []*efile
	CurrentFile efile
)

// ****************************************************************************
// openFile()
// ****************************************************************************
func openFile(filename string) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		SetStatus(fmt.Sprintf("Could not read %v", filename))
		CurrentFile.FName = filename
		CurrentFile.FemtoBuffer = femto.NewBufferFromString(string(""), CurrentFile.FName)
		CurrentFile.FemtoBuffer.Settings["keepautoindent"] = true
		CurrentFile.FemtoBuffer.Settings["softwrap"] = true
		CurrentFile.FemtoBuffer.Settings["scrollbar"] = true
		CurrentFile.FemtoBuffer.Settings["statusline"] = false
		// We create the view and open the buffer in the editor
		CurrentFile.FemtoView = femto.NewView(CurrentFile.FemtoBuffer)
		editor.OpenBuffer(CurrentFile.FemtoBuffer)
		SetTheme("monokai")
		SetStatus(fmt.Sprintf("Creating new file %v", filename))
	} else {
		CurrentFile.FName = filename
		CurrentFile.FemtoBuffer = femto.NewBufferFromString(string(content), CurrentFile.FName)
		CurrentFile.FemtoBuffer.Settings["keepautoindent"] = true
		CurrentFile.FemtoBuffer.Settings["softwrap"] = true
		CurrentFile.FemtoBuffer.Settings["scrollbar"] = true
		CurrentFile.FemtoBuffer.Settings["statusline"] = false

		CurrentFile.FemtoView = femto.NewView(CurrentFile.FemtoBuffer)
		editor.OpenBuffer(CurrentFile.FemtoBuffer)
		SetTheme("monokai")
	}
}

// ****************************************************************************
// newFile()
// ****************************************************************************
func newFile() {
	CurrentFile = efile{}
	CurrentFile.FName = "noname"
	CurrentFile.FemtoBuffer = femto.NewBufferFromString("", CurrentFile.FName)
	CurrentFile.FemtoBuffer.Settings["keepautoindent"] = true
	CurrentFile.FemtoBuffer.Settings["softwrap"] = true
	CurrentFile.FemtoBuffer.Settings["scrollbar"] = true
	CurrentFile.FemtoBuffer.Settings["statusline"] = false
	CurrentFile.FemtoView = femto.NewView(CurrentFile.FemtoBuffer)
	editor.OpenBuffer(CurrentFile.FemtoBuffer)
	SetTheme("monokai")
}

// ****************************************************************************
// saveFile()
// ****************************************************************************
func saveFile() {
	if CurrentFile.FName == "" {
		SetStatus("[red]Error : No file name defined (use 'Save as')[-]")
		return
	}

	content := editor.Buf.String()
	err := os.WriteFile(CurrentFile.FName, []byte(content), 0644)
	if err != nil {
		SetStatus(fmt.Sprintf("[red]Error saving file : %v[-]", err))
		return
	}

	CurrentFile.FemtoBuffer.IsModified = false // Set the buffer as not modified after saving
	SetStatus(fmt.Sprintf("File saved : %s", filepath.Base(CurrentFile.FName)))
}

// ****************************************************************************
// showSaveAsDialog()
// ****************************************************************************
func showSaveAsDialog() {
	inputField := tview.NewInputField().
		SetLabel("Save as : ").
		SetFieldWidth(30)

	inputField.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			path := inputField.GetText()
			if path != "" {
				CurrentFile.FName = path
				saveFile()
				pages.RemovePage("saveAs")
				app.SetFocus(editor)
			}
		} else if key == tcell.KeyEsc {
			pages.RemovePage("saveAs")
			app.SetFocus(editor)
		}
	})

	// On centre l'input dans une petite fenêtre
	modal := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(inputField, 3, 1, true).
			AddItem(nil, 0, 1, false), 40, 1, true).
		AddItem(nil, 0, 1, false)

	pages.AddPage("saveAs", modal, true, true)
}
