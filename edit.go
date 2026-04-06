package main

// ****************************************************************************
// IMPORTS
// ****************************************************************************
import (
	"bled/conf"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
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
		isText, _ := isTextFile(filename)
		if !isText {
			SetStatus("The file is not a valid text file")
			filename = filename + ".txt"
			newBuf = femto.NewBufferFromString(string(""), filename)
		} else {
			newBuf = femto.NewBufferFromString(string(content), filename)
		}
	}

	// FORCE PEP8 : Convert tabs to spaces for Python files
	if strings.HasSuffix(strings.ToLower(filename), ".py") {
		modifiedLines := convertTabsToSpaces(newBuf, 4)
		if modifiedLines > 0 {
			SetStatus(fmt.Sprintf("PEP8: %d lines converted to spaces", modifiedLines))
		}
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
// closeFile()
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
	currentDir = filepath.Dir(CurrentFile.FName)
	if IsFileGitTracked(CurrentFile.FName) {
		GITInfos = GetGITOneTagForFile(CurrentFile.FName)
	} else {
		GITInfos = fmt.Sprintf("%s %s %s", conf.APP_ICON, conf.APP_NAME, getFullVersion())
	}
	menuBar.rebuildBar()
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
// gotoLocation()
// ****************************************************************************
func gotoLocation(input string) {
	if CurrentFile == nil || CurrentFile.FemtoView == nil {
		return
	}

	buf := CurrentFile.FemtoBuffer
	view := CurrentFile.FemtoView
	var targetLine int
	totalLines := buf.NumLines

	// Nettoyage de l'entrée (minuscules et sans espaces)
	input = strings.ToLower(strings.TrimSpace(input))

	switch {
	case input == "top" || input == "start":
		targetLine = 0

	case input == "bottom" || input == "end":
		targetLine = totalLines - 1

	case strings.HasSuffix(input, "%"):
		percentStr := strings.TrimSuffix(input, "%")
		percent, err := strconv.Atoi(percentStr)
		if err != nil || percent < 0 || percent > 100 {
			SetStatus("Invalid percentage (ex: 50%)")
			return
		}
		targetLine = (totalLines * percent) / 100

	case strings.HasPrefix(input, "+"):
		relativeStr := input[1:] // Remove the '+' sign
		relative, err := strconv.Atoi(relativeStr)
		if err != nil {
			SetStatus("Relative line number invalid (ex: +10)")
			return
		}
		targetLine = view.Cursor.Y + relative

	case strings.HasPrefix(input, "-"):
		relativeStr := input[1:] // Remove the '-' sign
		relative, err := strconv.Atoi(relativeStr)
		if err != nil {
			SetStatus("Relative line number invalid (ex: -10)")
			return
		}
		targetLine = view.Cursor.Y - relative

	default:
		line, err := strconv.Atoi(input)
		if err != nil {
			SetStatus("Unknown command (ex: top, 50, -10, 80%)")
			return
		}
		targetLine = line - 1
	}

	if targetLine < 0 {
		targetLine = 0
	}
	if targetLine >= totalLines {
		targetLine = totalLines - 1
	}
	var loc femto.Loc
	loc.X = 0
	loc.Y = targetLine
	CurrentFile.FemtoBuffer.Cursor.GotoLoc(loc)
	editor.OpenBuffer(CurrentFile.FemtoBuffer)
	SetStatus(fmt.Sprintf("Go to line %d on %d", targetLine+1, totalLines))
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

	isSensitive := searchPanel.caseCheck.IsChecked()

	for i := 0; i < buf.NumLines; i++ {
		line := string(buf.Line(i))
		searchQuery := query

		if !isSensitive {
			line = strings.ToLower(line)
			searchQuery = strings.ToLower(query)
		}
		count += strings.Count(line, searchQuery)
	}

	foundLoc := findNextLoc(query, 0, 0)
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
		SetStatus("No current search query")
		return
	}

	if CurrentFile == nil || CurrentFile.FemtoView == nil {
		SetStatus("Error : No file open or view not initialized")
		return
	}

	view := CurrentFile.FemtoView

	if view.Cursor == nil {
		SetStatus("Error : Cursor not initialized")
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

	for y := startY; y < buf.NumLines; y++ {
		isSensitive := searchPanel.caseCheck.IsChecked()
		line := string(buf.Line(y))
		target := query
		currentLine := line

		if !isSensitive {
			currentLine = strings.ToLower(line)
			target = strings.ToLower(query)
		}

		offset := 0
		if y == startY {
			offset = startX
		}

		if offset < len(currentLine) {
			idx := strings.Index(currentLine[offset:], target)
			if idx != -1 {
				return &femto.Loc{X: offset + idx, Y: y}
			}
		}
	}

	return nil // No match found
}

// ****************************************************************************
// replaceCurrent()
// ****************************************************************************
func replaceCurrent() {
	query := searchPanel.searchInput.GetText()
	if query == "" {
		SetStatus("Query field is empty")
		return
	}

	view := CurrentFile.FemtoView
	buf := CurrentFile.FemtoBuffer
	replaceText := searchPanel.replaceInput.GetText()

	// We get the current line where the cursor is
	line := string(buf.Line(view.Cursor.Y))

	// Logic for detection (pay attention to case sensitivity if the option is checked)
	matchLine := line
	matchQuery := query
	if !searchPanel.caseCheck.IsChecked() {
		matchLine = strings.ToLower(line)
		matchQuery = strings.ToLower(query)
	}

	// Is the word right under the cursor ?
	if strings.HasPrefix(matchLine[view.Cursor.X:], matchQuery) {
		// Replacement
		buf.Replace(
			femto.Loc{X: view.Cursor.X, Y: view.Cursor.Y},
			femto.Loc{X: view.Cursor.X + len(query), Y: view.Cursor.Y},
			replaceText,
		)
		SetStatus("One occurrence has been replaced")
		// We jump to the next match after replacement
		lastSearchQuery = query // Update the global variable to ensure jumpToNextMatch uses the correct query
		jumpToNextMatch()
	} else {
		// If the cursor is not on a match, we search for the next one first
		lastSearchQuery = query // Update the global variable to ensure jumpToNextMatch uses the correct query
		jumpToNextMatch()
		SetStatus("Cursor moved to the next occurrence")
	}
}

// ****************************************************************************
// replaceAll()
// ****************************************************************************
func replaceAll() {
	query := searchPanel.searchInput.GetText()
	if query == "" {
		SetStatus("No current search query")
		return
	}
	lastSearchQuery = query // Update the global variable to ensure consistency

	buf := CurrentFile.FemtoBuffer
	replaceText := searchPanel.replaceInput.GetText()
	// isSensitive := searchPanel.caseCheck.IsChecked()
	count := 0

	// We go through the lines from bottom to top to avoid shifting the X/Y indices
	// while replacing texts of different lengths
	for y := buf.NumLines - 1; y >= 0; y-- {
		line := string(buf.Line(y))
		// Search logic (case sensitive or not)
		// ... (utiliser strings.ReplaceAll ou regex ici) ...

		// Exemple simplifié :
		if strings.Contains(line, lastSearchQuery) {
			newLine := strings.ReplaceAll(line, lastSearchQuery, replaceText)
			buf.Replace(femto.Loc{X: 0, Y: y}, femto.Loc{X: len(line), Y: y}, newLine)
			count++
		}
	}
	SetStatus("All occurrence(s) replaced")
	// SetStatus(fmt.Sprintf("All %d occurrence(s) replaced", count))
}
