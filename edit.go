package main

// ****************************************************************************
// IMPORTS
// ****************************************************************************
import (
	"fmt"
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
	CurrentFile *efile
)

// ****************************************************************************
// openFile()
// ****************************************************************************
func openFile(filename string) {
	/*
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
	*/
	// Vérification si déjà ouvert
	for i, f := range efiles {
		if f.FName == filename {
			switchDocument(i)
			return
		}
	}

	var newBuf *femto.Buffer
	content, err := os.ReadFile(filename)
	if err != nil {
		SetStatus(fmt.Sprintf("Could not read %v", filename))
		newBuf = femto.NewBufferFromString(string(""), filename)
		SetStatus(fmt.Sprintf("Creating new file %v", filename))
	} else {
		newBuf = femto.NewBufferFromString(string(content), filename)
	}

	newBuf.Settings["keepautoindent"] = true
	newBuf.Settings["softwrap"] = true
	newBuf.Settings["scrollbar"] = true
	newBuf.Settings["statusline"] = false

	newEFile := &efile{
		FemtoBuffer: newBuf,
		FName:       filename,
		Encoding:    "UTF-8",
		Modified:    false,
	}
	efiles = append(efiles, newEFile)
	switchDocument(len(efiles) - 1)
}

// ****************************************************************************
// newFile()
// ****************************************************************************
func newFile() {
	/*
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
	*/

	tempName := "noname"
	newBuf := femto.NewBufferFromString("", tempName)
	newEFile := &efile{
		FemtoBuffer: newBuf,
		FName:       tempName,
		Encoding:    "UTF-8",
		Modified:    false,
	}
	efiles = append(efiles, newEFile)
	switchDocument(len(efiles) - 1)
	SetStatus("New file created")
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

// ****************************************************************************
// switchDocument()
// ****************************************************************************
func switchDocument(index int) {
	if index < 0 || index >= len(efiles) {
		return
	}
	CurrentFile = efiles[index]
	editor.Buf = CurrentFile.FemtoBuffer
	editor.OpenBuffer(CurrentFile.FemtoBuffer)
	refreshFileMenu()
	SetStatus("Current file: " + filepath.Base(CurrentFile.FName))
	SetTheme("monokai")
}

// ****************************************************************************
// getFilePagination()
// ****************************************************************************
func getFilePagination() string {
	if len(efiles) == 0 || CurrentFile == nil {
		return "0/0"
	}

	currentIndex := 0
	for i, f := range efiles {
		if f == CurrentFile {
			currentIndex = i + 1 // +1 car les utilisateurs comptent à partir de 1
			break
		}
	}

	return fmt.Sprintf("%d/%d", currentIndex, len(efiles))
}

// ****************************************************************************
// nextFile()
// ****************************************************************************
func nextFile() {
	if len(efiles) <= 1 {
		return
	}

	// 1. Trouver l'index actuel
	idx := -1
	for i, f := range efiles {
		if f == CurrentFile {
			idx = i
			break
		}
	}

	// 2. Calculer le suivant (modulo pour boucler)
	nextIdx := (idx + 1) % len(efiles)
	switchDocument(nextIdx)
}

// ****************************************************************************
// prevFile()
// ****************************************************************************
func prevFile() {
	if len(efiles) <= 1 {
		return
	}

	// 1. Trouver l'index actuel
	idx := -1
	for i, f := range efiles {
		if f == CurrentFile {
			idx = i
			break
		}
	}

	// 2. Calculer le précédent (avec gestion de l'index négatif)
	prevIdx := (idx - 1 + len(efiles)) % len(efiles)
	switchDocument(prevIdx)
}
