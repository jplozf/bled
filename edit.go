package main

// ****************************************************************************
// IMPORTS
// ****************************************************************************
import (
	"fmt"
	"io/ioutil"

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
