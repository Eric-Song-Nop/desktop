package desktop

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// EntryType may be Application, Link or Directory.
type EntryType int

// All entry types
const (
	Unknown     EntryType = iota // Unspecified or unrecognized
	Application                  // Execute command
	Link                         // Open browser
	Directory                    // Open file manager
)

const sectionHeaderNotFoundError = "section header not found"

func (t EntryType) String() string {
	switch t {
	case Unknown:
		return "Unknown"
	case Application:
		return "Application"
	case Link:
		return "Link"
	case Directory:
		return "Directory"
	}

	return strconv.Itoa(int(t))
}

var (
	entryHeader      = []byte("[desktop entry]")
	entryType        = []byte("type=")
	entryName        = []byte("name=")
	entryGenericName = []byte("genericname=")
	entryComment     = []byte("comment=")
	entryIcon        = []byte("icon=")
	entryPath        = []byte("path=")
	entryExec        = []byte("exec=")
	entryURL         = []byte("url=")
	entryTerminal    = []byte("terminal=true")
	entryNoDisplay   = []byte("nodisplay=true")
	entryHidden      = []byte("hidden=true")
)

var quotes = map[string]string{
	`%%`:         `%`,
	`\\\\ `:      `\\ `,
	`\\\\` + "`": `\\` + "`",
	`\\\\$`:      `\\$`,
	`\\\\(`:      `\\(`,
	`\\\\)`:      `\\)`,
	`\\\\\`:      `\\\`,
	`\\\\\\\\`:   `\\\\`,
}

// Entry represents a parsed desktop entry.
type Entry struct {
	// Type is the type of the entry. It may be Application, Link or Directory.
	Type EntryType

	// Name is the name of the entry.
	Name string

	// GenericName is a generic description of the entry.
	GenericName string

	// Comment is extra information about the entry.
	Comment string

	// Icon is the path to an icon file or name of a themed icon.
	Icon string

	// Path is the directory to start in.
	Path string

	// Exec is the command(s) to be executed when launched.
	Exec string

	// URL is the URL to be visited when launched.
	URL string

	// Terminal controls whether to run in a terminal.
	Terminal bool
}

// ExpandExec fills keywords in the provided entry's Exec with user arguments.
func (e *Entry) ExpandExec(args string) string {
	ex := e.Exec

	ex = strings.ReplaceAll(ex, "%F", args)
	ex = strings.ReplaceAll(ex, "%f", args)
	ex = strings.ReplaceAll(ex, "%U", args)
	ex = strings.ReplaceAll(ex, "%u", args)

	return ex
}

func unquoteExec(ex string) string {
	for qs, qr := range quotes {
		ex = strings.ReplaceAll(ex, qs, qr)
	}

	return ex
}

// Parse reads and parses a .desktop file into an *Entry.
func Parse(content io.Reader, buf []byte) (*Entry, error) {
	var (
		scanner         = bufio.NewScanner(content)
		scannedBytes    []byte
		scannedBytesLen int

		entry       Entry
		foundHeader bool
	)

	scanner.Buffer(buf, len(buf))
	for scanner.Scan() {
		scannedBytes = bytes.TrimSpace(scanner.Bytes())
		scannedBytesLen = len(scannedBytes)

		if scannedBytesLen == 0 || scannedBytes[0] == byte('#') {
			continue
		} else if scannedBytes[0] == byte('[') {
			if !foundHeader {
				if scannedBytesLen < 15 || !bytes.EqualFold(scannedBytes[0:15], entryHeader) {
					return nil, errors.New(sectionHeaderNotFoundError)
				}

				foundHeader = true
			} else {
				break // Start of new section
			}
		} else if scannedBytesLen >= 6 && bytes.EqualFold(scannedBytes[0:5], entryType) {
			t := strings.ToLower(string(scannedBytes[5:]))
			switch t {
			case "application":
				entry.Type = Application
			case "link":
				entry.Type = Link
			case "directory":
				entry.Type = Directory
			}
		} else if scannedBytesLen >= 6 && bytes.EqualFold(scannedBytes[0:5], entryName) {
			entry.Name = string(scannedBytes[5:])
		} else if scannedBytesLen >= 13 && bytes.EqualFold(scannedBytes[0:12], entryGenericName) {
			entry.GenericName = string(scannedBytes[12:])
		} else if scannedBytesLen >= 9 && bytes.EqualFold(scannedBytes[0:8], entryComment) {
			entry.Comment = string(scannedBytes[8:])
		} else if scannedBytesLen >= 6 && bytes.EqualFold(scannedBytes[0:5], entryIcon) {
			entry.Icon = string(scannedBytes[5:])
		} else if scannedBytesLen >= 6 && bytes.EqualFold(scannedBytes[0:5], entryPath) {
			entry.Path = string(scannedBytes[5:])
		} else if scannedBytesLen >= 6 && bytes.EqualFold(scannedBytes[0:5], entryExec) {
			entry.Exec = unquoteExec(string(scannedBytes[5:]))
		} else if scannedBytesLen >= 5 && bytes.EqualFold(scannedBytes[0:4], entryURL) {
			entry.URL = string(scannedBytes[4:])
		} else if scannedBytesLen == 13 && bytes.EqualFold(scannedBytes, entryTerminal) {
			entry.Terminal = true
		} else if (scannedBytesLen == 14 && bytes.EqualFold(scannedBytes, entryNoDisplay)) || (scannedBytesLen == 11 && bytes.EqualFold(scannedBytes, entryHidden)) {
			return nil, nil
		}
	}

	err := scanner.Err()
	if err == nil && !foundHeader {
		err = errors.New(sectionHeaderNotFoundError)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to parse desktop entry: %s", err)
	}

	return &entry, nil
}
