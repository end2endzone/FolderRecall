//go:build windows

package spinner

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

// -----------------------------------------------------------------------
// Braille support detection
// -----------------------------------------------------------------------
//
// Legacy Windows consoles (cmd.exe / powershell.exe hosted by conhost.exe) render text using GDI
// with whatever TrueType font the console window is configured to use.
//
// On Windows 10 that font defaults to Consolas, which does not suppots glyphs for the Braille Patterns characters or Unicode characters.
// Printing such characters typically show boxes or blank characters, even if the code page is set to UTF-8.
// On consoles where Braille glyphs are unreliable we should fall back to the ASCII animation.
// However, we can hope to get support by changing the user's console font with SetCurrentConsoleFontEx() which affects the current console window.
// It does not always work. It is per Windows version, per font, per terminal.
//
// Some facts:
// - Fonts 'Consolas' and 'Lucida Console' does not supports Braille or Unicode characters.
// - Changing the font to 'Segoe UI Symbol' does supports Braille or Unicode characters,
//   but it also actually changes the font to 'Raster Fonts' whcih do not renders properly.
// - Font 'MS Gothic' does supports Braille or Unicode characters,
//   but it renders backslash characters incorrectly as a Y with double-strikethrough.
// - Fonts 'NSimSun', 'SimSun-ExtG' with a size of 18 does supports Braille or Unicode characters and seams appropriate,
//   but they are likely not installed on all systems.
//
// Windows Terminal draws text with DirectWrite() and performs automatic per-glyph font fallback.
// Braille characters render correctly regardless of the configured font.
// Windows Terminal sets the environment variable `WT_SESSION`.
// This package uses this variable to detect Windows Terminal and enable support.
//
// VSCode integrated terminal and its "JavaScript Debug Terminal" render through xterm.js inside a Chromium webview.
// It does supports Braille characters properly regardless of font.
// VSCode sets environment variable `TERM_PROGRAM=vscode` for its terminals instances.
// This package uses this variable to detect VSCode and enable support.

// -----------------------------------------------------------------------
// WIN32 api functions
// -----------------------------------------------------------------------

var (
	kernel32                   = syscall.NewLazyDLL("kernel32.dll")
	SetConsoleOutputCP         = kernel32.NewProc("SetConsoleOutputCP")
	GetCurrentConsoleFontEx    = kernel32.NewProc("GetCurrentConsoleFontEx")
	SetCurrentConsoleFontEx    = kernel32.NewProc("SetCurrentConsoleFontEx")
	CreateFileW                = kernel32.NewProc("CreateFileW")
	GetConsoleScreenBufferInfo = kernel32.NewProc("GetConsoleScreenBufferInfo")
	SetConsoleWindowInfo       = kernel32.NewProc("SetConsoleWindowInfo")
)

// -----------------------------------------------------------------------
// UTF-8 output code page
// -----------------------------------------------------------------------

// Cope page for UTF-8 in Windows
const cpUTF8 = 65001

// -----------------------------------------------------------------------
// Switching the console font
// -----------------------------------------------------------------------

const lfFaceSize = 32 // LF_FACESIZE

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
	NFont      uint32             // The index of the font in the system's console font table.
	FontSize   COORD              // A COORD structure that contains the width and height of each character in the font, in logical units. The X member contains the width, while the Y member contains the height.
	FontFamily uint32             // The font pitch and family. For information about the possible values for this member, see the description of the tmPitchAndFamily member of the TEXTMETRIC structure.
	FontWeight uint32             // The font weight. The weight can range from 100 to 1000, in multiples of 100. For example, the normal weight is 400, while 700 is bold.
	FaceName   [lfFaceSize]uint16 // The name of the typeface (such as Courier or Arial). Wide characters
}

const (
	genericRead    = 0x80000000
	genericWrite   = 0x40000000
	fileShareRead  = 0x00000001
	fileShareWrite = 0x00000002
	openExisting   = 3
)

// -----------------------------------------------------------------------
// Package windows initialization
// -----------------------------------------------------------------------

func init() {
	SupportsBrailleCharacters = detectBrailleSupport()

	// Force "Segoe UI Symbol" font to support Braille characters
	/*if !SupportsBrailleCharacters {
		fmt.Printf("Changing font to Segoe UI Symbol...\n")
		size := COORD{X: 0, Y: 16}
		err := ChangeConsoleFontWithSize("Segoe UI Symbol", &size)
		if err != nil {
			panic(err)
		}

		SupportsBrailleCharacters = true
		fmt.Printf("Font is now 'Segoe UI Symbol' {%v,%v}\n", size.X, size.Y)
	}*/

	// Enable UTF-8 supports for other characters such as `✔`.
	enableUTF8ConsoleOutput()
}

func detectBrailleSupport() bool {
	if os.Getenv("WT_SESSION") != "" {
		return true // Windows Terminal
	}
	if os.Getenv("TERM_PROGRAM") == "vscode" {
		return true // VSCode
	}
	return false // plain conhost-hosted cmd.exe or powershell.exe
}

// enableUTF8ConsoleOutput changes the console's code page to UTF-8 (65001).
// This is worth doing even on consoles that fall back to ASCII spinner frames, because it also fixes any other Unicode chatacters.
// For example, the "✔" character requires UTF-8 code page to be properly printed.
//
// This only changes the *encoding* used to interpret bytes written to the console.
// It does not affect which glyphs the console's font can renders.
func enableUTF8ConsoleOutput() {
	if SetConsoleOutputCP.Find() != nil {
		return // unavailable on this Windows version; ignore silently
	}
	_, _, _ = SetConsoleOutputCP.Call(uintptr(cpUTF8))
}

func printConsoleFontIndex(cfi CONSOLE_FONT_INFOEX, prefix string) {
	fmt.Printf("{\n")
	fmt.Printf("  %sNFont=%v\n", prefix, cfi.NFont)
	fmt.Printf("  %sFontSize={%v,%v}\n", prefix, cfi.FontSize.X, cfi.FontSize.Y)
	fmt.Printf("  %sFontFamily=%v\n", prefix, cfi.FontFamily)
	fmt.Printf("  %sFontWeight=%v\n", prefix, cfi.FontWeight)
	fmt.Printf("}\n")
}

// ChangeConsoleFont changes the current console font.
func ChangeConsoleFont(fontName string) error {
	return ChangeConsoleFontWithSize(fontName, nil)
}

// ChangeConsoleFontWithSize changes the current console font and font size, if specified.
func ChangeConsoleFontWithSize(fontName string, fontSize *COORD) error {
	h, err := openConsoleOutputHandle()
	if err != nil {
		return err
	}
	defer syscall.CloseHandle(h)

	var cfi CONSOLE_FONT_INFOEX
	cfi.CbSize = uint32(unsafe.Sizeof(cfi))

	ret, _, err := GetCurrentConsoleFontEx.Call(
		uintptr(h),
		0, // FALSE (get current window size metrics)
		uintptr(unsafe.Pointer(&cfi)))
	if ret == 0 {
		return fmt.Errorf("failed to read console font: %v", err)
	}

	// Debug
	// printConsoleFontIndex(cfi, "before.")

	// Explicitly clear the layout index to force lookup by FaceName
	cfi.NFont = 0

	// Set X to 0 to prevent vector/raster aspect ratio warping.
	// Windows automatically calculates the width matching the Y height.
	cfi.FontSize.X = 0

	// Force modern family and normal font style (no italic or bold)
	cfi.FontFamily = 54  // FF_MODERN
	cfi.FontWeight = 400 // FW_NORMAL

	// Convert desired font (e.g., "Consolas") into a UTF-16 slice
	utf16Font, err := syscall.UTF16FromString(fontName)
	if err != nil {
		return err
	}
	if len(utf16Font) > 32 {
		return fmt.Errorf("length of font name is more than 32 characters: %s", fontName)
	}

	// Clean the font face string to prevent string truncation fallbacks.
	// Reset array to zeros first
	cfi.FaceName = [lfFaceSize]uint16{}
	// Copy UTF-16 values into the fixed-size structure buffer
	copy(cfi.FaceName[:], utf16Font)

	// Force the size if specified
	if fontSize != nil {
		cfi.FontSize = *fontSize
	}

	// Debug
	// printConsoleFontIndex(cfi, "patched.")

	ret, _, err = SetCurrentConsoleFontEx.Call(uintptr(h), 0, uintptr(unsafe.Pointer(&cfi)))
	if ret == 0 {
		return err
	}

	ret, _, err = GetCurrentConsoleFontEx.Call(uintptr(h), 0, uintptr(unsafe.Pointer(&cfi)))
	if ret == 0 {
		return err
	}

	// Debug
	// printConsoleFontIndex(cfi, "after.")

	// Force the console host window to re-draw its frame.
	// Without this, the terminal will keep rendering characters using the old font buffer cache
	/*var csbi CONSOLE_SCREEN_BUFFER_INFO
	stdoutHandle := uintptr(h)
	ret, _, err = GetConsoleScreenBufferInfo.Call(stdoutHandle, uintptr(unsafe.Pointer(&csbi)))
	if ret == 0 {
		return err
	}
	ret, _, err = SetConsoleWindowInfo.Call(stdoutHandle, uintptr(1), uintptr(unsafe.Pointer(&csbi.SrWindow)))
	if ret == 0 {
		return err
	}*/

	return nil
}

// EnableLucidaConsoleFont changes the current console font to "Lucida Console".
// This font supports Braille Pattern glyph on legacy Windows 10 consoles.
// Lucida Console ships with Windows by default.
func EnableLucidaConsoleFont() error {
	err := ChangeConsoleFont("Lucida Console")
	return err
}

// EnableLucidaConsoleFont changes the current console font to "Lucida Console".
// This fonc supports Braille Pattern glyph on legacy Windows 10 consoles according to Microsoft Word.
// Segoe UI Symbol font is available with recent Windows 10 installations.
func EnableSegoeUISymbolFont() error {
	err := ChangeConsoleFont("Segoe UI Symbol")
	return err
}

func openConsoleOutputHandle() (syscall.Handle, error) {
	utf16Name, err := syscall.UTF16PtrFromString("CONOUT$")
	if err != nil {
		return 0, err
	}
	h, _, err := CreateFileW.Call(
		uintptr(unsafe.Pointer(utf16Name)),
		uintptr(genericRead|genericWrite),
		uintptr(fileShareRead|fileShareWrite),
		0,
		uintptr(openExisting),
		0,
		0,
	)

	handle := syscall.Handle(h)
	if handle == syscall.InvalidHandle {
		return 0, err
	}
	return handle, nil
}
