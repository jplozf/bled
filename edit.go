package main

// ****************************************************************************
// IMPORTS
// ****************************************************************************
import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pgavlin/femto"
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
	ReadOnly    bool
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
func openFile(filename string, readOnly bool) {
	// Is already open ? Just switch to it
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
		ReadOnly:    readOnly,
	}
	efiles = append(efiles, newEFile)
	switchDocument(len(efiles) - 1)
}

// ****************************************************************************
// newFile()
// ****************************************************************************
func newFile() {
	// Find a unique name for the new file (noname, noname-01, noname-02, etc.)
	tempName := "noname"
	counter := 1

	// OLoop until we find a name that is not already used by an open file
	for {
		found := false
		for _, f := range efiles {
			if f.FName == tempName {
				found = true
				break
			}
		}
		if !found {
			break // The name is unique, we can use it
		}
		// Otherwise, generate the next name and check again
		tempName = fmt.Sprintf("noname-%02d", counter)
		counter++
	}

	// Create a new buffer with the unique name
	newBuf := femto.NewBufferFromString("", tempName)

	// Set the buffer settings for the new file
	newBuf.Settings["keepautoindent"] = true
	newBuf.Settings["softwrap"] = true
	newBuf.Settings["scrollbar"] = true
	newBuf.Settings["statusline"] = false

	newEFile := &efile{
		FemtoBuffer: newBuf,
		FName:       tempName,
		Encoding:    "UTF-8",
		Modified:    false,
		ReadOnly:    false,
	}
	efiles = append(efiles, newEFile)
	switchDocument(len(efiles) - 1)
	SetStatus(fmt.Sprintf("New file created : %s", tempName))
}

// ****************************************************************************
// saveFile()
// ****************************************************************************
func closeFile(index int) {
	if index < 0 || index >= len(efiles) {
		return
	}

	// Remove the file from the list
	efiles = append(efiles[:index], efiles[index+1:]...)

	// If there are still files open, switch to the first one
	if len(efiles) > 0 {
		switchDocument(0)
	} else {
		newFile()
	}
}

// ****************************************************************************
// saveFile()
// ****************************************************************************
func saveFile() {
	if CurrentFile.ReadOnly {
		SetStatus("[red]Error: Cannot save a read-only file[-]")
		return
	}
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
	SetTheme(config.Theme)
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
			currentIndex = i + 1 // +1 since we want one-based index for display
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

	// Find the current index
	idx := -1
	for i, f := range efiles {
		if f == CurrentFile {
			idx = i
			break
		}
	}

	// Calculate the next index (with wrap-around)
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

	// Find the current index
	idx := -1
	for i, f := range efiles {
		if f == CurrentFile {
			idx = i
			break
		}
	}

	// Calculate the previous index (with wrap-around)
	prevIdx := (idx - 1 + len(efiles)) % len(efiles)
	switchDocument(prevIdx)
}

// ****************************************************************************
// GoBottom()
// ****************************************************************************
func GoBottom() {
	var loc femto.Loc
	loc.X = 0
	loc.Y = CurrentFile.FemtoBuffer.End().Y
	CurrentFile.FemtoBuffer.Cursor.GotoLoc(loc)
	editor.OpenBuffer(CurrentFile.FemtoBuffer)
	SetStatus("Go to bottom")
}

// ****************************************************************************
// GoTop()
// ****************************************************************************
func GoTop() {
	var loc femto.Loc
	loc.X = 0
	loc.Y = 0
	CurrentFile.FemtoBuffer.Cursor.GotoLoc(loc)
	editor.OpenBuffer(CurrentFile.FemtoBuffer)
	SetStatus("Go to top")
}

// ****************************************************************************
// GoLine()
// ****************************************************************************
func GoLine(l int) {
	if l < 1 {
		SetStatus("Jump outside bounds")
		GoTop()
	} else {
		if l <= CurrentFile.FemtoBuffer.LinesNum() {
			var loc femto.Loc
			loc.X = 0
			loc.Y = l - 1
			CurrentFile.FemtoBuffer.Cursor.GotoLoc(loc)
			editor.OpenBuffer(CurrentFile.FemtoBuffer)
			SetStatus(fmt.Sprintf("Go to line #%d", l))
		} else {
			SetStatus("Jump too far")
			GoBottom()
		}
	}
}
