package utils

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type FileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
}

// NewFileInfoFromLs parse ls -al output to FileInfos, eg:
// total 64
// drwxr-xr-x    1 root     root          4096 Jan 29 03:25 .
// drwxr-xr-x    1 root     root          4096 Jan 29 03:25 ..
// -rwxr-xr-x    1 root     root             0 Jan 29 03:25 .dockerenv
// drwxr-xr-x    2 root     root          4096 Aug 20 10:30 bin
// ...
func NewFileInfoFromLs(lsout string) []*FileInfo {
	infos := []*FileInfo{}
	lines := strings.Split(lsout, "\n")

	// single file
	if len(lines) == 1 {
		return []*FileInfo{parseLine(lines[0])}
	}

	// directory, should be at least [total . .. empty]
	if len(lines) < 4 {
		panic("invalid ls output" + lsout)
	}
	for i, line := range lines[1 : len(lines)-1] {
		if i == 1 { // ..
			continue
		}

		infos = append(infos, parseLine(line))
	}

	return infos
}

func parseLine(line string) *FileInfo {
	secs := make([]string, 0, 9)
	for _, sec := range strings.Split(line, " ") {
		if sec != "" {
			secs = append(secs, sec)
		}
	}

	// 9 columes
	if len(secs) < 9 {
		panic("invalid ls line: " + line)
	}
	size, _ := strconv.ParseInt(secs[4], 10, 64)
	return &FileInfo{
		name:    secs[8],
		size:    size,
		mode:    parseMode(secs[0]),
		modTime: parseModTime(strings.Join(secs[5:8], " ")),
	}
}

// parseMode is revert of os.FileMode formater
func parseMode(mode string) os.FileMode {
	fileMode := os.FileMode(0)

	mode = strings.TrimSuffix(mode, "@")
	special := mode[:len(mode)-9]
	const str = "dalTLDpSugct?"
	for _, c := range special {
		for i := range str {
			if c == rune(str[i]) {
				fileMode += (os.FileMode(1) << (32 - 1 - i))
			}
		}
	}

	rwx := mode[len(mode)-9:]
	for idx, shift := 0, 8; idx < 9; idx, shift = idx+1, shift-1 {
		if rwx[idx] != '-' {
			fileMode += os.FileMode(1 << shift)
		}
	}
	return fileMode
}

// parseModTime parse modTime from ls output
func parseModTime(ts string) time.Time {
	time.Now().Year()
	t, err := time.Parse("Jan 2 15:04", ts)
	if err != nil {
		panic(err)
	}
	return t
}

func (i *FileInfo) SetName(name string) { i.name = name }
func (i *FileInfo) Name() string        { return i.name }
func (i *FileInfo) Size() int64         { return i.size }
func (i *FileInfo) Mode() os.FileMode   { return i.mode }
func (i *FileInfo) ModTime() time.Time  { return i.modTime }
func (i *FileInfo) IsDir() bool         { return i.mode.IsDir() }
func (i *FileInfo) Sys() interface{}    { return nil }
