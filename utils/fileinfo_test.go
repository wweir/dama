package utils

import (
	"os/exec"

	"github.com/wweir/utils/log"
)

func Example_NewFileInfoFromLs() {
	out, err := exec.Command("ls", "-alT").Output()
	if err != nil {
		log.Fatalw("ls fail", "err", err, "output", string(out))
	}

	for _, fi := range NewFileInfoFromLs(string(out)) {
		log.Infow("fileinfo", "name", fi.Name(), "mode", fi.Mode(), "time", fi.ModTime())
	}

	// Output: 1
}
