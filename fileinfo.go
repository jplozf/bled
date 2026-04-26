package main

// ****************************************************************************
// IMPORTS
// ****************************************************************************
import (
	"bled/conf"
	"bled/utils"
	"fmt"
	"net/http"
	"os"

	"github.com/rivo/tview"
)

// ****************************************************************************
// TYPES
// ****************************************************************************
type FileInfo struct {
	Size     int64
	Mime     string
	Lines    int
	Modified string
}

// ****************************************************************************
// VARIABLES
// ****************************************************************************
var infoPanel *tview.TextView
var infoPanelActive bool

// ****************************************************************************
// getExtendedFileInfo()
// ****************************************************************************
func getExtendedFileInfo(path string) FileInfo {
	stats, err := os.Stat(path)
	if err != nil {
		return FileInfo{}
	}

	f, _ := os.Open(path)
	buffer := make([]byte, 512)
	n, _ := f.Read(buffer)
	f.Close()
	mime := http.DetectContentType(buffer[:n])

	lines := CurrentFile.FemtoBuffer.NumLines

	return FileInfo{
		Size:     stats.Size(),
		Mime:     mime + "; syntax=" + CurrentFile.FemtoBuffer.Settings["filetype"].(string) + "; type=" + CurrentFile.FemtoBuffer.Settings["fileformat"].(string) + ";",
		Lines:    lines,
		Modified: stats.ModTime().Format("02/01/2006 15:04"),
	}
}

// ****************************************************************************
// initInfoPanel()
// ****************************************************************************
func initInfoPanel() {
	infoPanel = tview.NewTextView()
	infoPanel.SetDynamicColors(true).SetRegions(true).SetBorder(true).SetTitle(" File Info ")
	infoPanel.SetBackgroundColor(conf.GetColor(config.MenuBgColor))
}

// ****************************************************************************
// func toggleInfoPanel()
// ****************************************************************************
func toggleInfoPanel() {
	if infoPanelActive {
		layout.ResizeItem(infoPanel, 0, 0)
		app.SetFocus(editor)
	} else {
		info := getExtendedFileInfo(CurrentFile.FName)

		txt := fmt.Sprintf(" [%s]Path     :[-] %s\n [%s]Size     :[-] %d bytes (%s)\n [%s]Type     :[-] %s\n [%s]Lines    :[-] %d\n [%s]Last Mod :[-] %s",
			HighlightedColor, CurrentFile.FName, HighlightedColor, info.Size, utils.HumanFileSize(float64(info.Size)), HighlightedColor, info.Mime, HighlightedColor, info.Lines, HighlightedColor, info.Modified)

		infoPanel.SetText(txt)
		layout.ResizeItem(infoPanel, 7, 0)
		app.SetFocus(infoPanel)
	}
	infoPanelActive = !infoPanelActive
}
