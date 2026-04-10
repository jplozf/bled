package main

import (
	"bled/utils"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ****************************************************************************
// VARS
// ****************************************************************************
var (
	DlgYesNo1   *Dialog
	DlgYesNo2   *Dialog
	DlgYesNo    *Dialog
	DlgInputBox *Dialog
)

// ****************************************************************************
// TYPES
// ****************************************************************************
type GitInfo struct {
	Commit  string
	Status  string
	Branch  string
	FStatus string
}

// ****************************************************************************
// DoGitStatus()
// ****************************************************************************
func DoGitStatus() {
	if IsInsideGitWorkTree() {
		out := fmt.Sprintf("Current local commit : %s\n\n", XeqOut("git rev-parse --short HEAD"))
		out += fmt.Sprintf("%s\n", XeqOut("git status"))
		out += fmt.Sprintf("%s\n", XeqOut("git diff"))
		MsgBox = MsgBox.OK("Git Status", out, nil, 0, "main", editor)
		pages.AddPage("msgBox", MsgBox.Popup(), true, false)
		pages.ShowPage("msgBox")
	} else {
		SetStatus("No Git repository found")
	}
}

// ****************************************************************************
// DoGitLog()
// ****************************************************************************
func DoGitLog() {
	if IsInsideGitWorkTree() {
		out := fmt.Sprintf("%s", XeqOut("git log"))
		MsgBox = MsgBox.OK("Git Log", out, nil, 0, "main", editor)
		pages.AddPage("msgBox", MsgBox.Popup(), true, false)
		pages.ShowPage("msgBox")
	} else {
		SetStatus("No Git repository found")
	}
}

// ****************************************************************************
// DoGitBang()
// ****************************************************************************
func DoGitBang() {
	if !IsInsideGitWorkTree() {
		pjname := filepath.Base(filepath.Dir(currentDir))
		DlgYesNo1 = DlgYesNo1.YesNo("Git Bang", // Title
			fmt.Sprintf("This will initialize and push a Git environment\nfor your project \"%s\".\n\nAre you sure you want to proceed ?", pjname), // Message
			func(rc DlgButton, idx int) {
				if rc == BUTTON_YES {
					if config.GithubUser == "" {
						SetStatus("Github User is not yet configured")
					} else {
						DlgYesNo2 = DlgYesNo2.YesNo("Git Bang", // Title
							"The following repository should already exist :\n\nhttps://github.com/"+config.GithubUser+"/"+pjname+"\n\nHas the repository already been created ?", // Message
							func(rc DlgButton, idx int) {
								if rc == BUTTON_YES {
									url := "https://" + config.GithubKey + "@github.com/" + config.GithubUser + "/" + pjname
									Xeq("git init")
									Xeq("git add .")
									Xeq("git commit -m \"First Commit\"")
									Xeq("git remote add origin " + url)
									Xeq("git branch -M main")
									Xeq("git pull --rebase origin main")
									Xeq("git push -u origin main")
									if !utils.IsFileExist(filepath.Join(currentDir, "README.md")) {
										createTextFile(filepath.Join(currentDir, "README.md"), "#"+pjname)
										Xeq("git add README.md")
									}
									if !utils.IsFileExist(filepath.Join(currentDir, ".gitignore")) {
										createTextFile(filepath.Join(currentDir, ".gitignore"),
											`# This file is used to ignore files which are generated
# ----------------------------------------------------------------------------

github.key
*~
*.autosave
*.a
*.core
*.moc
*.o
*.obj
*.orig
*.rej
*.so
*.so.*
*_pch.h.cpp
*_resource.rc
*.qm
.#*
*.*#
core
!core/
tags
.DS_Store
.directory
*.debug
Makefile*
*.prl
*.app
moc_*.cpp
ui_*.h
qrc_*.cpp
Thumbs.db
*.res
*.rc
/.qmake.cache
/.qmake.stash

# qtcreator generated files
*.pro.user*
*.qbs.user*
CMakeLists.txt.user*

# xemacs temporary files
*.flc

# Vim temporary files
.*.swp

# Visual Studio generated files
*.ib_pdb_index
*.idb
*.ilk
*.pdb
*.sln
*.suo
*.vcproj
*vcproj.*.*.user
*.ncb
*.sdf
*.opensdf
*.vcxproj
*vcxproj.*

# MinGW generated files
*.Debug
*.Release

# Python byte code
*.pyc

# Binaries
# --------
*.dll
*.exe

# Directories with generated files
.moc/
.obj/
.pch/
.rcc/
.uic/
/build*/
`)
										Xeq("git add .gitignore")
									}
									DoGitStatus()
								} else {
									SetStatus("Aborting Git Bang")
								}
							},
							0,
							"main", editor) // Focus return
						pages.AddPage("dlgYesNo2", DlgYesNo2.Popup(), true, false)
						pages.ShowPage("dlgYesNo2")
					}
				} else {
					SetStatus("Aborting Git Bang")
				}
			},
			0,
			"main", editor) // Focus return
		pages.AddPage("dlgYesNo1", DlgYesNo1.Popup(), true, false)
		pages.ShowPage("dlgYesNo1")
	} else {
		SetStatus("Git repository already created")
	}
}

// ****************************************************************************
// DoGitInit()
// ****************************************************************************
func DoGitInit() {
	if !IsInsideGitWorkTree() {
		pjname := filepath.Base(currentDir)
		DlgYesNo1 = DlgYesNo1.YesNo("Git Init", // Title
			fmt.Sprintf("This will initialize a Git environment\nfor your project \"%s\".\n\nAre you sure you want to proceed ?", pjname), // Message
			func(rc DlgButton, idx int) {
				if rc == BUTTON_YES {
					if config.GithubUser == "" {
						SetStatus("Git User is not yet configured")
					} else {

						if !utils.IsFileExist(filepath.Join(currentDir, "README.md")) {
							createTextFile(filepath.Join(currentDir, "README.md"), "#"+pjname)
						}
						if !utils.IsFileExist(filepath.Join(currentDir, ".gitignore")) {
							createTextFile(filepath.Join(currentDir, ".gitignore"),
								`# This file is used to ignore files which are generated
# ----------------------------------------------------------------------------

github.key
*~
*.autosave
*.a
*.core
*.moc
*.o
*.obj
*.orig
*.rej
*.so
*.so.*
*_pch.h.cpp
*_resource.rc
*.qm
.#*
*.*#
core
!core/
tags
.DS_Store
.directory
*.debug
Makefile*
*.prl
*.app
moc_*.cpp
ui_*.h
qrc_*.cpp
Thumbs.db
*.res
*.rc
/.qmake.cache
/.qmake.stash

# qtcreator generated files
*.pro.user*
*.qbs.user*
CMakeLists.txt.user*

# xemacs temporary files
*.flc

# Vim temporary files
.*.swp

# Visual Studio generated files
*.ib_pdb_index
*.idb
*.ilk
*.pdb
*.sln
*.suo
*.vcproj
*vcproj.*.*.user
*.ncb
*.sdf
*.opensdf
*.vcxproj
*vcxproj.*

# MinGW generated files
*.Debug
*.Release

# Python byte code
*.pyc

# Binaries
# --------
*.dll
*.exe

# Directories with generated files
.moc/
.obj/
.pch/
.rcc/
.uic/
/build*/
`)
						}
						Xeq("git init")
						// Xeq("git add .")
						DoGitStatus()
						Xeq("git branch -M main")
						go refreshStatus()
					}
				} else {
					SetStatus("Aborting Git Init")
				}
			},
			0,
			"main", editor) // Focus return
		pages.AddPage("dlgYesNo1", DlgYesNo1.Popup(), true, false)
		pages.ShowPage("dlgYesNo1")
	} else {
		SetStatus("Git repository already created")
	}
}

// ****************************************************************************
// DoGitCommitPush()
// ****************************************************************************
func DoGitCommitPush() {
	if IsInsideGitWorkTree() {
		DlgInputBox = DlgInputBox.Input("Git Commit & Push", // Title
			"Please, enter the message :", // Message
			"",                            // Default value
			func(rc DlgButton, idx int) {
				if rc == BUTTON_OK {
					msg := DlgInputBox.Value
					if msg == "" {
						msg = fmt.Sprintf("Commited on %s", time.Now().Format("09-07-2017 17:06:06"))
					}
					out := fmt.Sprintf("Committing...\n%s", XeqOut("git commit -a -m "+msg))
					branch := XeqRaw("git rev-parse --abbrev-ref HEAD")
					out += fmt.Sprintf("\n\nPushing...\n%s", XeqOut("git push origin "+branch))

					MsgBox = MsgBox.OK("Git Commit & Push", out, nil, 0, "main", editor)
					pages.AddPage("msgBox", MsgBox.Popup(), true, false)
					pages.ShowPage("msgBox")
				} else {
					SetStatus("Aborting Git Commit & Push")
				}
			},
			0,
			"main", editor) // Focus return
		pages.AddPage("dlgInput", DlgInputBox.Popup(), true, false)
		pages.ShowPage("dlgInput")
	} else {
		SetStatus("No Git repository found")
	}
}

// ****************************************************************************
// DoGitFetch()
// ****************************************************************************
func DoGitFetch() {
	if IsInsideGitWorkTree() {
		DlgYesNo = DlgYesNo.YesNo("Git Fetch", // Title
			"The Git Fetch will fetch the remote version but no merging is processed locally.\n\nAre you sure you want to proceed ?", // Message
			func(rc DlgButton, idx int) {
				if rc == BUTTON_YES {
					out := fmt.Sprintf("Fetching...\n%s", XeqOut("git fetch origin"))
					MsgBox = MsgBox.OK("Git Fetch", out, nil, 0, "main", editor)
					pages.AddPage("msgBox", MsgBox.Popup(), true, false)
					pages.ShowPage("msgBox")
				} else {
					SetStatus("Aborting Git Fetch")
				}
			},
			0,
			"main", editor) // Focus return
		pages.AddPage("dlgYesNo", DlgYesNo.Popup(), true, false)
		pages.ShowPage("dlgYesNo")
	} else {
		SetStatus("No Git repository found")
	}
}

// ****************************************************************************
// DoGitPull()
// ****************************************************************************
func DoGitPull() {
	if IsInsideGitWorkTree() {
		branch := XeqOut("git rev-parse --abbrev-ref HEAD")
		DlgYesNo = DlgYesNo.YesNo("Git Pull", // Title
			"The Git Pull will fetch the remote version and merge it with you local branch which is currently ["+branch+"].\n\nAre you sure you want to proceed ?", // Message
			func(rc DlgButton, idx int) {
				if rc == BUTTON_YES {
					out := fmt.Sprintf("Pulling...\n%s", XeqOut("git pull origin "+branch))
					MsgBox = MsgBox.OK("Git Pull", out, nil, 0, "main", editor)
					pages.AddPage("msgBox", MsgBox.Popup(), true, false)
					pages.ShowPage("msgBox")
				} else {
					SetStatus("Aborting Git Pull")
				}
			},
			0,
			"main", editor) // Focus return
		pages.AddPage("dlgYesNo", DlgYesNo.Popup(), true, false)
		pages.ShowPage("dlgYesNo")
	} else {
		SetStatus("No Git repository found")
	}
}

// ****************************************************************************
// DoGitPush()
// ****************************************************************************
func DoGitPush() {
	if IsInsideGitWorkTree() {
		branch := XeqRaw("git rev-parse --abbrev-ref HEAD")
		out := fmt.Sprintf("Pushing...\n%s", XeqOut("git push origin "+branch))
		MsgBox = MsgBox.OK("Git Push", out, nil, 0, "main", editor)
		pages.AddPage("msgBox", MsgBox.Popup(), true, false)
		pages.ShowPage("msgBox")
	} else {
		SetStatus("No Git repository found")
	}
}

// ****************************************************************************
// DoGitCommit()
// ****************************************************************************
func DoGitCommit() {
	if IsInsideGitWorkTree() {
		DlgInputBox = DlgInputBox.Input("Git Commit", // Title
			"Please, enter the message :", // Message
			"",                            // Default value
			func(rc DlgButton, idx int) {
				if rc == BUTTON_OK {
					msg := DlgInputBox.Value
					if msg == "" {
						msg = fmt.Sprintf("Commited on %s", time.Now().Format("09-07-2017 17:06:06"))
					}
					out := fmt.Sprintf("Committing...\n%s", XeqOut("git commit -a -m "+msg))
					MsgBox = MsgBox.OK("Git Commit", out, nil, 0, "main", editor)
					pages.AddPage("msgBox", MsgBox.Popup(), true, false)
					pages.ShowPage("msgBox")
				} else {
					SetStatus("Aborting Git Commit")
				}
			},
			0,
			"main", editor) // Focus return
		pages.AddPage("dlgInput", DlgInputBox.Popup(), true, false)
		pages.ShowPage("dlgInput")
	} else {
		SetStatus("No Git repository found")
	}
}

// ****************************************************************************
// DoGitClone()
// ****************************************************************************
func DoGitClone() {
	if !IsInsideGitWorkTree() {
		DlgInputBox = DlgInputBox.Input("Git Clone", // Title
			"Please, enter the distant repository to clone locally :", // Message
			"https://github.com/repo/project.git",                     // Default value
			func(rc DlgButton, idx int) {
				if rc == BUTTON_OK {
					out := fmt.Sprintf("Cloning...\n%s", XeqOut("git clone "+DlgInputBox.Value))
					MsgBox = MsgBox.OK("Git Clone", out, nil, 0, "main", editor)
					pages.AddPage("msgBox", MsgBox.Popup(), true, false)
					pages.ShowPage("msgBox")
				} else {
					SetStatus("Aborting Git Clone")
				}
			},
			0,
			"main", editor) // Focus return
		pages.AddPage("dlgInput", DlgInputBox.Popup(), true, false)
		pages.ShowPage("dlgInput")
	} else {
		SetStatus("Already in a Git repository")
	}
}

// ****************************************************************************
// DoGitAdd()
// ****************************************************************************
func DoGitAdd(f any) {
	if IsInsideGitWorkTree() {
		DlgYesNo = DlgYesNo.YesNo("Git Add", // Title
			"This will add the file :\n\n"+f.(string)+"\n\nto the Git tracking.\n\nAre you sure you want to proceed ?", // Message
			func(rc DlgButton, idx int) {
				if rc == BUTTON_YES {
					b := filepath.Base(f.(string))
					out := fmt.Sprintf("Adding...\n%s", XeqOut("git add ./"+utils.EscapeSpaces(b)))
					MsgBox = MsgBox.OK("Git Add", out, nil, 0, "main", editor)
					pages.AddPage("msgBox", MsgBox.Popup(), true, false)
					pages.ShowPage("msgBox")
				} else {
					SetStatus("Aborting Git Add")
				}
			},
			0,
			"main", editor) // Focus return
		pages.AddPage("dlgYesNo", DlgYesNo.Popup(), true, false)
		pages.ShowPage("dlgYesNo")
	} else {
		SetStatus("No Git repository found")
	}
}

// ****************************************************************************
// DoGitAddAll()
// ****************************************************************************
func DoGitAddAll() {
	if IsInsideGitWorkTree() {
		DlgYesNo = DlgYesNo.YesNo("Git Add All", // Title
			"This will add all the files to the Git tracking.\n\nAre you sure you want to proceed ?", // Message
			func(rc DlgButton, idx int) {
				if rc == BUTTON_YES {
					out := fmt.Sprintf("Adding...\n%s", XeqOut("git add ."))
					MsgBox = MsgBox.OK("Git Add All", out, nil, 0, "main", editor)
					pages.AddPage("msgBox", MsgBox.Popup(), true, false)
					pages.ShowPage("msgBox")
				} else {
					SetStatus("Aborting Git Add All")
				}
			},
			0,
			"main", editor) // Focus return
		pages.AddPage("dlgYesNo", DlgYesNo.Popup(), true, false)
		pages.ShowPage("dlgYesNo")
	} else {
		SetStatus("No Git repository found")
	}
}

// ****************************************************************************
// IsInsideGitWorkTree()
// ****************************************************************************
func IsInsideGitWorkTree() bool {
	rc := false
	out := XeqOut("git rev-parse --is-inside-work-tree")
	if out[:4] == "true" {
		rc = true
	}
	return rc
}

// ****************************************************************************
// IsFileGitTracked()
// ****************************************************************************
func IsFileGitTracked(f string) bool {
	rc := false
	out := XeqOut("git ls-files --error-unmatch " + utils.EscapeSpaces(f))
	b := filepath.Base(f)
	if out[:len(b)] == b {
		SetStatus(fmt.Sprintf("File [%s] is git-tracked", f))
		rc = true
	} else {
		SetStatus(fmt.Sprintf("File [%s] is NOT git-tracked", f))
	}
	return rc
}

// ****************************************************************************
// createTextFile()
// ****************************************************************************
func createTextFile(fName string, text string) {
	f, err := os.Create(fName)
	if err == nil {
		defer f.Close()
		fmt.Fprintln(f, text)
	}
}

// ****************************************************************************
// GetGITInfosForFile()
// ****************************************************************************
func GetGITInfosForFile(f string) GitInfo {
	var GitCommit, GitStatus, GitBranch, GitFileStatus string

	ws := filepath.Dir(f)

	// Get GIT Commit
	commit, err2 := utils.Xeq(ws, "git", "rev-parse", "--short", "HEAD")
	if err2 != "" {
		commit = "None"
	}
	GitCommit = strings.TrimSpace(commit)

	// Get GIT Status
	status, _ := utils.Xeq(ws, "git", "status", "-s")
	if status != "" {
		GitStatus = "Pending"
	} else {
		GitStatus = "Up2Date"
	}

	// Get GIT Branch
	branch, _ := utils.Xeq(ws, "git", "rev-parse", "--abbrev-ref", "HEAD")
	/* if err3 != "" {
		branch = "Unknown"
	} */
	GitBranch = strings.TrimSpace(branch)

	// Get GIT File Status
	fstatus, _ := utils.Xeq(ws, "git", "status", "-s", f)
	if fstatus != "" {
		fstatus = fstatus[0:2]
	} else {
		fstatus = ""
	}
	GitFileStatus = fstatus

	return GitInfo{Commit: GitCommit, Status: GitStatus, Branch: GitBranch, FStatus: GitFileStatus}
}

// ****************************************************************************
// GetGITOneTagForFile()
// ****************************************************************************
func GetGITOneTagForFile(f string) string {
	gitInfo := GetGITInfosForFile(f)
	s := fmt.Sprintf("%s:%s:%s:%s:", gitInfo.FStatus, gitInfo.Commit, gitInfo.Status, gitInfo.Branch)
	SetStatus(s)
	return s
}
