package main

import (
	"bled/conf"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/go-cmd/cmd"
)

// ****************************************************************************
// Xeq()
// ****************************************************************************
func Xeq(c string) {
	sCmd := strings.Fields(c)
	if len(ACmd) > 0 {
		if ACmd[len(ACmd)-1] != c {
			ACmd = append(ACmd, c)
			ICmd++
		}
	} else {
		ACmd = append(ACmd, c)
		ICmd++
	}

	if len(sCmd) > 0 {
		SetStatus(fmt.Sprintf("Running [%s]", c))
		// 1. Setup the command
		cmdOptions := cmd.Options{
			Buffered:  false, // We want streaming
			Streaming: true,
		}
		xCmd := cmd.NewCmdOptions(cmdOptions, sCmd[0], sCmd[1:]...)
		//activeCmd = xCmd // Assign to the shared variable
		xCmd.Dir = currentDir

		// 2. Open the log file once (use O_APPEND)
		fOut, err := os.OpenFile(conf.GetLogPath(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			// Handle error
			SetStatus("Error opening log file: " + err.Error())
		} else {
			SetStatus("Using " + conf.GetLogPath())
		}

		// 3. Start the command in the background
		statusChan := xCmd.Start()
		// Check if the command failed to even start (e.g., command not found)
		initialStatus := xCmd.Status()
		if initialStatus.Error != nil {
			SetStatus("Error: " + initialStatus.Error.Error())
			// Log it to the file as well
			fmt.Fprintln(fOut, "START ERROR: "+initialStatus.Error.Error())
			return
		}

		// 4. Handle Output & Lifecycle in a single Goroutine

		go func() {
			defer fOut.Close()

			// Write header to file
			fmt.Fprintf(fOut, "%s ⯈ %s\n", time.Now().Format("20060102-150405"), c)
			for {
				select {
				case line, open := <-xCmd.Stdout:
					if !open {
						xCmd.Stdout = nil
					} else {
						fmt.Fprintln(fOut, line)
					}
				case line, open := <-xCmd.Stderr:
					if !open {
						xCmd.Stderr = nil
					} else {
						fmt.Fprintln(fOut, line)
					}
				case status := <-statusChan:
					// Command finished!
					fmt.Fprintf(fOut, "%s ⯈ Done [%s] Exit Code: %d\n", time.Now().Format("20060102-150405"), c, status.Exit)
					app.QueueUpdateDraw(func() {
						SetStatus(fmt.Sprintf("Done [%s] Exit Code: %d", c, status.Exit))
					})
					return // Exit the goroutine
				}

				// If both streams are closed but status hasn't arrived,
				// we still need to wait for statusChan to avoid leaking
				if xCmd.Stdout == nil && xCmd.Stderr == nil && statusChan == nil {
					return
				}
			}
		}()

	} else {
		SetStatus("Nothing to run")
	}
}

// ****************************************************************************
// XeqOut()
// ****************************************************************************
func XeqOut(c string) string {
	// sCmd := strings.Fields(c)
	// https://stackoverflow.com/questions/47489745/splitting-a-string-at-space-except-inside-quotation-marks
	quoted := false
	sCmd := strings.FieldsFunc(c, func(r rune) bool {
		if r == '"' {
			quoted = !quoted
		}
		return !quoted && r == ' '
	})

	out := ""
	if len(sCmd) > 0 {
		cmd := exec.Command(sCmd[0], sCmd[1:]...)
		cmd.Dir = currentDir
		SetStatus(fmt.Sprintf("Executing [%s] in %s", c, cmd.Dir))
		var outb, errb bytes.Buffer
		cmd.Stdout = &outb
		cmd.Stderr = &errb
		if err := cmd.Run(); err != nil {
			out = "Error : " + err.Error()
			if exitError, ok := err.(*exec.ExitError); ok {
				out = out + fmt.Sprintf("\nExit code %d", exitError.ExitCode())
			}
		} else {
			out = outb.String()
			out = out + errb.String()
			out = out + "\nExit code 0"
		}
	} else {
		out = "Nothing to run\n\nExit code 0"
	}

	out = strings.TrimSpace(out)
	SetStatus(out)
	SetStatus(fmt.Sprintf("Done [%s]", c))
	return out
}

// ****************************************************************************
// XeqOutErr()
// ****************************************************************************
func XeqOutErr(c string) string {
	// sCmd := strings.Fields(c)
	// https://stackoverflow.com/questions/47489745/splitting-a-string-at-space-except-inside-quotation-marks
	quoted := false
	sCmd := strings.FieldsFunc(c, func(r rune) bool {
		if r == '"' {
			quoted = !quoted
		}
		return !quoted && r == ' '
	})

	out := ""
	if len(sCmd) > 0 {
		cmd := exec.Command(sCmd[0], sCmd[1:]...)
		cmd.Dir = currentDir
		SetStatus(fmt.Sprintf("Executing [%s] in %s", c, cmd.Dir))
		var outb, errb bytes.Buffer
		cmd.Stdout = &outb
		cmd.Stderr = &errb
		if err := cmd.Run(); err != nil {
			out = "Error : " + err.Error()
			if exitError, ok := err.(*exec.ExitError); ok {
				out = out + fmt.Sprintf("\nExit code %d", exitError.ExitCode())
				out = out + outb.String()
				out = out + errb.String()
			}
		} else {
			out = outb.String()
			out = out + errb.String()
			out = out + "\nExit code 0"
		}
	} else {
		out = "Nothing to run\n\nExit code 0"
	}

	out = strings.TrimSpace(out)
	SetStatus(out)
	SetStatus(fmt.Sprintf("Done [%s]", c))
	return out
}

// ****************************************************************************
// XeqRaw()
// ****************************************************************************
func XeqRaw(c string) string {
	// sCmd := strings.Fields(c)
	// https://stackoverflow.com/questions/47489745/splitting-a-string-at-space-except-inside-quotation-marks
	quoted := false
	sCmd := strings.FieldsFunc(c, func(r rune) bool {
		if r == '"' {
			quoted = !quoted
		}
		return !quoted && r == ' '
	})

	out := ""
	if len(sCmd) > 0 {
		cmd := exec.Command(sCmd[0], sCmd[1:]...)
		cmd.Dir = currentDir
		SetStatus(fmt.Sprintf("Executing [%s] in %s", c, cmd.Dir))
		var outb, errb bytes.Buffer
		cmd.Stdout = &outb
		cmd.Stderr = &errb
		if err := cmd.Run(); err != nil {
			out = "Error : " + err.Error()
			if exitError, ok := err.(*exec.ExitError); ok {
				out = out + fmt.Sprintf("\nExit code %d", exitError.ExitCode())
			}
		} else {
			out = outb.String()
			out = out + errb.String()
		}
	} else {
		out = "Nothing to run\n\nExit code 0"
	}

	out = strings.TrimSpace(out)
	SetStatus(out)
	SetStatus(fmt.Sprintf("Done [%s]", c))
	return out
}
