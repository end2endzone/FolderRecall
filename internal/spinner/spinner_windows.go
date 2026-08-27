//go:build windows

package spinner

import (
	"syscall"
	"unsafe"
)

var SupportsBrailleCharacters = true

type COORD struct {
	X int16
	Y int16
}

type SMALL_RECT struct {
	Left   int16
	Top    int16
	Right  int16
	Bottom int16
}

type CONSOLE_SCREEN_BUFFER_INFO struct {
	DwSize              COORD
	DwCursorPosition    COORD
	WAttributes         uint16
	SrWindow            SMALL_RECT
	DwMaximumWindowSize COORD
}

type CONSOLE_FONT_INFOEX struct {
	CbSize     uint32
	NFont      uint32
	DwFontSize COORD
	FontFamily uint32
	FontWeight uint32
	FaceName   [32]uint16
}

func changeFont(stdoutHandle uintptr, kernel32 *syscall.LazyDLL) {
	setCurrentConsoleFontEx := kernel32.NewProc("SetCurrentConsoleFontEx")
	getConsoleScreenBufferInfo := kernel32.NewProc("GetConsoleScreenBufferInfo")
	setConsoleWindowInfo := kernel32.NewProc("SetConsoleWindowInfo")

	if setCurrentConsoleFontEx.Find() != nil {
		SupportsBrailleCharacters = false
		return
	}

	targetFont := "Lucida Console"
	utf16Font, err := syscall.UTF16FromString(targetFont)
	if err != nil || len(utf16Font) > 32 {
		SupportsBrailleCharacters = false
		return
	}

	var cfi CONSOLE_FONT_INFOEX
	cfi.CbSize = uint32(unsafe.Sizeof(cfi))
	cfi.NFont = 0
	cfi.DwFontSize.X = 0
	cfi.DwFontSize.Y = 16
	cfi.FontFamily = 54  // FF_MODERN
	cfi.FontWeight = 400 // FW_NORMAL
	copy(cfi.FaceName[:], utf16Font)

	ret, _, _ := setCurrentConsoleFontEx.Call(stdoutHandle, uintptr(0), uintptr(unsafe.Pointer(&cfi)))
	if ret == 0 {
		SupportsBrailleCharacters = false
		return
	}

	var csbi CONSOLE_SCREEN_BUFFER_INFO
	retcsbi, _, _ := getConsoleScreenBufferInfo.Call(stdoutHandle, uintptr(unsafe.Pointer(&csbi)))
	if retcsbi != 0 && setConsoleWindowInfo.Find() == nil {
		_, _, _ = setConsoleWindowInfo.Call(stdoutHandle, uintptr(1), uintptr(unsafe.Pointer(&csbi.SrWindow)))
	}
}

func init() {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")

	// 1. Force the console to use UTF-8 output code page (65001)
	setConsoleOutputCP := kernel32.NewProc("SetConsoleOutputCP")
	_, _, _ = setConsoleOutputCP.Call(uintptr(65001))

	createFile := kernel32.NewProc("CreateFileW")
	getConsoleMode := kernel32.NewProc("GetConsoleMode")
	setConsoleMode := kernel32.NewProc("SetConsoleMode")

	const enableVirtualTerminalProcessing = 0x0004

	// FIX: If standard GetStdHandle is wrapped inside a pipe stream,
	// open "CONOUT$" directly to gain access to the raw Windows screen layout buffer.
	conoutPath, _ := syscall.UTF16PtrFromString("CONOUT$")

	// GENERIC_READ = 0x80000000, GENERIC_WRITE = 0x40000000
	const genericReadWrite = 0x80000000 | 0x40000000
	// FILE_SHARE_WRITE = 0x00000002
	const fileShareWrite = 0x00000002
	// OPEN_EXISTING = 3
	const openExisting = 3

	ret, _, _ := createFile.Call(
		uintptr(unsafe.Pointer(conoutPath)),
		uintptr(genericReadWrite),
		uintptr(fileShareWrite),
		0,
		uintptr(openExisting),
		0,
		0,
	)

	if ret != 0 && ret != uintptr(syscall.InvalidHandle) {
		var mode uint32

		retMode, _, _ := getConsoleMode.Call(ret, uintptr(unsafe.Pointer(&mode)))
		if retMode != 0 {
			mode |= enableVirtualTerminalProcessing

			retSetMode, _, _ := setConsoleMode.Call(ret, uintptr(mode))
			if retSetMode == 0 {
				SupportsBrailleCharacters = false
			}
		} else {
			SupportsBrailleCharacters = false
		}

		changeFont(ret, kernel32)

		// Clean up our targeted window handle
		closeHandle := kernel32.NewProc("CloseHandle")
		_, _, _ = closeHandle.Call(ret)
	} else {
		SupportsBrailleCharacters = false
	}
}
