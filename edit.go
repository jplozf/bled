package main

// ****************************************************************************
// IMPORTS
// ****************************************************************************
import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

	var newView *femto.View
	newView = femto.NewView(newBuf)

	newEFile := &efile{
		FemtoBuffer: newBuf,
		FName:       filename,
		FemtoView:   newView,
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
	var newView *femto.View
	newView = femto.NewView(newBuf)
	newEFile := &efile{
		FemtoBuffer: newBuf,
		FemtoView:   newView,
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

// ****************************************************************************
// performSearch()
// ****************************************************************************
func performSearch(query string) int {
	if query == "" || CurrentFile == nil {
		return 0
	}

	buf := CurrentFile.FemtoBuffer
	view := CurrentFile.FemtoView
	count := 0
	var foundLoc *femto.Loc

	searchQuery := strings.ToLower(query)

	for i := 0; i < buf.NumLines; i++ {
		line := strings.ToLower(string(buf.Line(i)))
		line = strings.TrimRight(line, "\r\n") // Remove trailing newline characters for accurate searching

		if strings.Contains(line, searchQuery) {
			occurences := strings.Count(line, searchQuery)

			if count == 0 {
				idx := strings.Index(line, searchQuery)
				foundLoc = &femto.Loc{X: idx, Y: i}
			}
			count += occurences
		}
	}

	if foundLoc != nil && view != nil {
		view.Cursor.Loc = *foundLoc
		editor.OpenBuffer(buf) // Refresh the view to show the new cursor position
		view.Center()
	}

	return count
}

// ****************************************************************************
// jumpToNextMatch()
// ****************************************************************************
func jumpToNextMatch() {
	if lastSearchQuery == "" {
		SetStatus("[yellow]No current search query[-]")
		return
	}

	if CurrentFile == nil || CurrentFile.FemtoView == nil {
		SetStatus("[red]Error : No file open or view not initialized[-]")
		return
	}

	view := CurrentFile.FemtoView

	if view.Cursor == nil {
		SetStatus("[red]Error : Cursor not initialized[-]")
		return
	}

	currentX := view.Cursor.X
	currentY := view.Cursor.Y

	next := findNextLoc(lastSearchQuery, currentX+1, currentY)
	currentMatchIdx++
	if currentMatchIdx > totalMatches {
		currentMatchIdx = 1
	}
	searchPanel.label.SetText(fmt.Sprintf(" [black]Match %d of %d", currentMatchIdx, totalMatches))
	if next == nil {
		next = findNextLoc(lastSearchQuery, 0, 0)
	}

	if next != nil {
		view.Cursor.X = next.X
		view.Cursor.Y = next.Y
		view.Center()
		var loc femto.Loc
		loc.X = next.X
		loc.Y = next.Y
		CurrentFile.FemtoBuffer.Cursor.GotoLoc(loc)
		editor.OpenBuffer(CurrentFile.FemtoBuffer)
	}
}

// ****************************************************************************
// findNextLoc()
// ****************************************************************************
func findNextLoc(query string, startX, startY int) *femto.Loc {
	if CurrentFile == nil || CurrentFile.FemtoBuffer == nil {
		return nil
	}

	buf := CurrentFile.FemtoBuffer
	query = strings.ToLower(query)

	if startY < buf.NumLines {
		line := strings.ToLower(string(buf.Line(startY)))
		if startX < len(line) {
			idx := strings.Index(line[startX:], query)
			if idx != -1 {
				return &femto.Loc{X: startX + idx, Y: startY}
			}
		}
	}

	for y := startY + 1; y < buf.NumLines; y++ {
		line := strings.ToLower(string(buf.Line(y)))
		idx := strings.Index(line, query)
		if idx != -1 {
			return &femto.Loc{X: idx, Y: y}
		}
	}

	return nil // No match found
}
