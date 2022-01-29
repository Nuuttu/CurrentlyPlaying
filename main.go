package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"syscall"
	"time"
	"unsafe"

	"github.com/lxn/walk"
	"golang.org/x/sys/windows"
)

var (
	mod                     = windows.NewLazyDLL("user32.dll")
	procGetWindowText       = mod.NewProc("GetWindowTextW")
	procGetWindowTextLength = mod.NewProc("GetWindowTextLengthW")
)

type (
	HANDLE uintptr
	HWND   HANDLE
)

var csong string

func main() {

	mw, err := walk.NewMainWindow()
	if err != nil {
		log.Fatal(err)
	}

	ni, err := walk.NewNotifyIcon(mw)
	if err != nil {
		log.Fatal(err)
	}
	defer ni.Dispose()

	// We load our icon from a file.
	if fileExists("./ico256.ico") {
		icon, err := walk.Resources.Icon("./ico256.ico")
		if err != nil {
			log.Fatal(err)
		}
		// Set the icon and a tool tip text.
		if err := ni.SetIcon(icon); err != nil {
			log.Fatal(err)
		}
	}
	if err := ni.SetToolTip("Click for info or use the context menu to exit."); err != nil {
		log.Fatal(err)
	}
	// When the left mouse button is pressed, bring up our balloon.

	ni.MouseDown().Attach(func(x, y int, button walk.MouseButton) {
		if button != walk.LeftButton {
			return
		}
		if err := ni.ShowMessage(
			"CurrentSong By Tuomo",
			`Select Spotify window by selecting Spotify window after clicking "Select Winow"`); err != nil {

			log.Fatal(err)
		}
	})
	// We put a select window -action into the context menu.
	selectWindowAction := walk.NewAction()
	if err := selectWindowAction.SetText("&Select Window to Capture"); err != nil {
		log.Fatal(err)
	}
	selectWindowAction.Triggered().Attach(func() { getForeWindow(ni) })
	if err := ni.ContextMenu().Actions().Add(selectWindowAction); err != nil {
		log.Fatal(err)
	}
	// Adding file location opener to the context menu
	openLocationAction := walk.NewAction()
	if err := openLocationAction.SetText("&Open CurrentSong.txt"); err != nil {
		log.Fatal(err)
	}
	openLocationAction.Triggered().Attach(func() { openFileLocation() })
	if err := ni.ContextMenu().Actions().Add(openLocationAction); err != nil {
		log.Fatal(err)
	}
	// We put an exit action into the context menu.
	exitAction := walk.NewAction()
	if err := exitAction.SetText("E&xit"); err != nil {
		log.Fatal(err)
	}
	exitAction.Triggered().Attach(func() { walk.App().Exit(0) })
	if err := ni.ContextMenu().Actions().Add(exitAction); err != nil {
		log.Fatal(err)
	}

	// The notify icon is hidden initially, so we have to make it visible.
	if err := ni.SetVisible(true); err != nil {
		log.Fatal(err)
	}
	ni.ShowMessage("CurrentSong by Tuomo", "Start from Notify Area Icon menu")
	mw.Run()

	// ###############################################################################################
	// ###############################################################################################
	// ###############################################################################################
	// ###############################################################################################
	// ###############################################################################################
	// ###############################################################################################
	// ###############################################################################################
	// ###############################################################################################

	defer clearTextFromCurrentSong()

}

var looping bool = true

func looploop(hwnd uintptr) {
	runtime.GOMAXPROCS(1)
	looping = true
	for {
		nsong := GetWindowText(HWND(hwnd))
		if nsong == "Spotify Premium" {
			nsong = ""
		}
		if csong != nsong {
			csong = nsong
			if csong == "Spotify Premium" {
				csong = ""
			}
			err := ioutil.WriteFile("currentSong.txt", []byte(csong), 0777)
			if err != nil {
				log.Fatal(err)
			}
		}
		time.Sleep(time.Second)
		if !looping {
			break
		}
	}
}

func getForeWindow(ni *walk.NotifyIcon) {
	looping = false
	hwnd1 := getWindow("GetForegroundWindow")
	for {
		time.Sleep(time.Second)
		hwnd2 := getWindow("GetForegroundWindow")
		if hwnd1 != hwnd2 {
			break
		}
	}
	hwnd := getWindow("GetForegroundWindow")
	if hwnd != 0 {
		text := GetWindowText(HWND(hwnd))
		fmt.Println("window selected >>>>> ", text)
		ni.ShowMessage("Selected Window",
			text)
		go looploop(hwnd)
	}
}

func clearTextFromCurrentSong() {
	csong := ""
	data := []byte(csong)
	err := ioutil.WriteFile("currentSong.txt", data, 0777)
	if err != nil {
		log.Fatal(err)
	}
}

func getWindow(funcName string) uintptr {
	proc := mod.NewProc(funcName)
	hwnd, _, _ := proc.Call()
	return hwnd
}

func GetWindowTextLength(hwnd HWND) int {
	ret, _, _ := procGetWindowTextLength.Call(
		uintptr(hwnd))

	return int(ret)
}

func GetWindowText(hwnd HWND) string {
	textLen := GetWindowTextLength(hwnd) + 1

	buf := make([]uint16, textLen)
	procGetWindowText.Call(
		uintptr(hwnd),
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(textLen))

	return syscall.UTF16ToString(buf)
}

func fileExists(fp string) bool {
	info, err := os.Stat(fp)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func openFileLocation() {

	ex, err := os.Executable()
	if err != nil {
		log.Panic("asdasd", err)
	}
	currentSongPath := filepath.Dir(ex) + `\currentSong.txt`
	fmt.Println("exPath", currentSongPath)

	cmd := exec.Command(`explorer`, `/select,`+currentSongPath)
	fmt.Println(cmd)
	if err := cmd.Run(); err != nil {
		log.Println(err)
		fmt.Println("Ignore the error...")
	}
	// WHY THE ERROR!?!??!?!? It works
}
