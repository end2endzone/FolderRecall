package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/end2endzone/FolderRecall/internal/build"
	ole "github.com/go-ole/go-ole"
	// ole2 "github.com/go-ole/go-ole/oleutil"
)

// Config holds all the command-line argument values
type Config struct {
	CommandPrint   bool
	CommandMonitor bool
	DatabasePath   string
	Interval       int
	NoHeader       bool
	Verbose        bool
	Version        bool
	Help           bool
}

const DefaultInterval = 30
const InvalidInterval = 0

func main() {
	os.Exit(run(os.Args[1:]))
}

func printHeader() {
	fmt.Fprintf(os.Stdout, "fldrecall - Logs or recalls your Windows File Explorer navigation history.\n")
}

func printVersion(verbose bool) {
	fmt.Fprintf(os.Stdout, "Version %s.\n", GetProductVersionString())

	if verbose {
		metadata := build.GetBuildMetadata()
		fmt.Fprintf(os.Stdout, "Build metadata:%s\n", metadata)
	}
}

func reportArgumentParsingError(format string, args ...any) {
	// Prefix the message with a clear "Error:" label
	fullFormat := "Error: " + format + "\n"

	// Print the formatted string directly to standard error
	fmt.Fprintf(os.Stderr, fullFormat, args...)
}

// hasArgument checks if a specific argument exists in the raw command-line arguments.
// The function also automatically checks with the `-` and `--` prefixes.
// For example,  hasArgument("version") returns true when `--version` is one of the arguments.
// Returns true when the given value is found. Returns false otherwise.
func hasArgument(value string) bool {
	// Loop through arguments, skipping the first one (program path)
	for _, arg := range os.Args[1:] {
		// Specific value
		if arg == value {
			// Found exact match
			return true
		}

		// Try the `-` and `--` prefixes
		if arg == "--"+value || arg == "-"+value {
			// Found exact match
			return true
		}
	}
	return false
}

// newAppFlagSet initializes the FlagSet and binds fields directly to the Config struct
func newAppFlagSet(cfg *Config) *flag.FlagSet {
	fs := flag.NewFlagSet("fldrecall", flag.ContinueOnError)

	// Bind string flags directly to the struct fields
	fs.BoolVar(&cfg.CommandPrint, "print", false, "|Print the current list of directories in File Explorer.")
	fs.BoolVar(&cfg.CommandMonitor, "monitor", false, "|Monitor and log navigation history.")
	fs.StringVar(&cfg.DatabasePath, "dbpath", "", "<path>|Path to the database file to store navigation history.")
	fs.IntVar(&cfg.Interval, "interval", InvalidInterval, "<value>|Interval time in seconds between snapshots.")
	fs.BoolVar(&cfg.NoHeader, "no-header", false, "|Do not show product header when running a command.")
	fs.BoolVar(&cfg.Verbose, "verbose", false, "|Enable verbose output for the command.")
	fs.BoolVar(&cfg.Version, "version", false, "|Show the product version.")

	// The flag library automatically registers `--help`, `-help` and `-h` flags.
	// We register the help flag specifically to make sure this flag is printed in our custom fs.Usage() function.
	fs.BoolVar(&cfg.Help, "help", false, "|Show this usage message.")

	// Attach the custom usage printer
	fs.Usage = func() {
		// This function can be called for multiple reasons:
		// 1. Parsing errors. When called, the parsing error is already printed to fs.Output().
		// 2. One of `--help`, `-help` or `-h` flags was used.
		//    Since we specifically register a manual `help` flag in our flagset, the flag library will not
		//    call this function for `--help` and `-help` but it does for `-h`.

		// Force usage text to be displayed on stdout by temporary switching output to stdout
		originalOutput := fs.Output() // os.Stderr or os.Stdout
		fs.SetOutput(os.Stdout)

		// Show an empty line between the error and the usage text (on stdout)
		fmt.Fprintln(os.Stdout)

		// Print usage text
		printUsage(fs)

		// Restore output to whatever it was set
		fs.SetOutput(originalOutput)
	}

	return fs
}

// getOrderedFlags visits all flags in a flagset and make a slice with them.
// The returned slices is also ordered so that `version` and `help` flags are last.
func getOrderedFlags(fs *flag.FlagSet) []*flag.Flag {
	flags := make([]*flag.Flag, 0)

	// Create a slice with all flags, skipping some flags...
	fs.VisitAll(func(f *flag.Flag) {
		if f.Name == "verbose" ||
			f.Name == "version" ||
			f.Name == "help" ||
			f.Name == "no-header" {
			return // skip
		}
		flags = append(flags, f)
	})

	// Add our bottom flags at the end of the list.
	flags = append(flags, fs.Lookup("no-header"))
	flags = append(flags, fs.Lookup("verbose"))
	flags = append(flags, fs.Lookup("version"))
	flags = append(flags, fs.Lookup("help"))

	return flags
}

// StringSplitAtLast splits a given string at the last occurance of the given separator.
func StringSplitAtLast(s, separator string) []string {
	// Find the last occurrence of the separator
	i := strings.LastIndex(s, separator)

	// If the separator is not found, return the original string as a single-element slice
	if i == -1 {
		return []string{s}
	}

	// Slice the string before and after the last separator
	return []string{
		s[:i],
		s[i+len(separator):],
	}
}

// printUsage print a usage string that output each arguments and then examples.
func printUsage(fs *flag.FlagSet) {
	output := fs.Output() // os.Stderr

	// Print static usage header
	const usageText = `Usage:
    fldrecall --print [--no-header] [--verbose]
    fldrecall --monitor [--no-header] --dbpath <path> [--no-header] [--verbose]
    fldrecall --version [--verbose]
    fldrecall --help
	`
	fmt.Fprintln(output, usageText)

	// Print all flags and their descriptions in a 2 columns layout.
	// Column 0 is the name of the argument and its value descriptor such as `<path>`.
	// Column 1 is the flag's usage description. The description can contain \n character to force the following text to be displayed on the next line.
	// For example:
	// |----------------------------|---------------------------------------------------|
	// `  --dbpath <path>            Path to the database file to store the             `
	// `                             navigation history.                                `

	fmt.Fprintln(output, "Flags:")

	orderedFlags := getOrderedFlags(fs)
	for _, f := range orderedFlags {
		// Split our custom usage metadata format: `<placeholder>|Description string`.
		// Using StringSplitAtLast() instead of strings.SplitN(f.Usage, "|", 2) to support
		// placeholders that contains optional names.
		// For example `<path|uuid>`.
		parts := StringSplitAtLast(f.Usage, "|")

		placeholder := ""
		description := f.Usage

		if len(parts) == 2 {
			placeholder = parts[0]
			description = parts[1]
		}

		// Construct flag component (for example "--install <path>")
		flagStr := "  --" + f.Name
		if placeholder != "" {
			flagStr += " " + placeholder
		}

		// Split the description into multiple lines to handle clean indentation alignment
		lines := strings.Split(description, "\n")

		// Target column where the 2nd column (description text) must begin
		const targetCol = 30

		// Print the first line
		if len(flagStr) < targetCol {
			// Pad the remaining space up to column targetCol
			padding := strings.Repeat(" ", targetCol-len(flagStr))
			fmt.Fprintf(output, "%s%s%s\n", flagStr, padding, lines[0])
		} else {
			// If the flag declaration is too long that it breaks our alignement...

			// Print it on its own line
			fmt.Fprintln(output, flagStr)

			// Then align the description
			padding := strings.Repeat(" ", targetCol)
			fmt.Fprintf(output, "%s%s\n", padding, lines[0])
		}

		// Print any multi-line description wrap-arounds exactly at column targetCol
		for i := 1; i < len(lines); i++ {
			padding := strings.Repeat(" ", targetCol)
			fmt.Fprintf(output, "%s%s\n", padding, lines[i])
		}
	}
	fmt.Fprintln(output)

	// Print static examples
	const exampleText = `Examples:
	fldrecall --print
	fldrecall --monitor --dbpath %%USERPROFILE%%\fldrecall.db --interval 30
	`

	fmt.Fprintf(output, exampleText) //Bug: can not use fmt.Fprintln() without error: "fmt.Fprintln call has possible Printf formatting directive %U"
}

func run(args []string) int {
	var cfg Config
	fs := newAppFlagSet(&cfg)

	// Manually parse for `--no-header` and `--version` arguments before calling fs.Parse().
	// In case of parsing errors, the error will be printed before the flag library will call our custom fs.Usage().
	// So we must print the application's header or version before doing the actual parsing.
	cfg.NoHeader = hasArgument("no-header")
	cfg.Verbose = hasArgument("verbose")
	cfg.Version = hasArgument("version")

	// Should we only print the version ?
	if cfg.Version {
		printVersion(cfg.Verbose)
		return 0
	}

	// Print application header, unless specified not to
	if !cfg.NoHeader {
		printHeader()
		printVersion(false)
	}

	var err error

	// Parse arguments
	fs.SetOutput(os.Stderr) // parsing errors should be printed to stderr
	err = fs.Parse(args)
	fs.SetOutput(os.Stdout) // after parsing, following outputs should be printed to stdout
	if err != nil {

		// The flag library automatically registers `--help`, `-help` and `-h` flags.
		// When specified on the command line, these flags reports the specific error `flag.ErrHelp` on parsing.
		// Since we specifically register a manual `help` flag in our flagset, the flag library do not report
		// the error `flag.ErrHelp` for `--help` and `-help` but it does for `-h`.
		if err == flag.ErrHelp {
			// There is no need to call printUsage(fs) since the flag library has already called fs.Usage() because of the error.
			return 0
		}

		reportArgumentParsingError("invalid arguments: %v", err)
		return 2
	}

	// Help flag set
	if cfg.Help {
		// Triggered by `--help`, `-help`
		printUsage(fs)
		return 0
	}

	// Check optional argument, set default value if unspecified
	if cfg.Interval == InvalidInterval {
		cfg.Interval = DefaultInterval
	}

	// Count how many commands are specified in the arguments
	// Do not count `--version` and `--help` as these were already processed above.
	commandsSet := 0
	for _, set := range []bool{cfg.CommandMonitor, cfg.CommandPrint} {
		if set {
			commandsSet++
		}
	}

	// Act accordingly if too none or too many are specified
	switch {
	case commandsSet == 0:
		// No command specified. Do not know that to do.
		fmt.Fprint(os.Stderr, "no command specified\n\n")

		// Then show help message.
		printUsage(fs)
		return 2
	case commandsSet > 1:
		reportArgumentParsingError("please specify only one command at a time")
		return 2
	}

	// Initialize COM library
	ole.CoInitialize(0)
	defer ole.CoUninitialize()

	// Call the actual command helpers
	switch {
	case cfg.CommandPrint:
		err = cmdPrint(cfg)
	case cfg.CommandMonitor:
		err = cmdMonitor(cfg)
	}

	// Check for an error while running a command.
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}
	return 0
}

// cmdMonitor monitors the navigation history at the given interval and store it in the database.
func cmdMonitor(cfg Config) error {
	return nil
}

// cmdPrint get the current list of directories from File Explorer and print them on the console.
func cmdPrint(cfg Config) error {
	snapshot, err := GetCurrentDirectoriesSnapshot()
	if err != nil {
		return err
	}

	snapshot.Sort()

	fmt.Printf("\n")
	fmt.Printf("Directories at %s : \n\n", snapshot.TimeStamp)
	for idx, path := range snapshot.paths {
		fmt.Printf("%02d: %s\n", idx, path)
	}

	return nil
}
