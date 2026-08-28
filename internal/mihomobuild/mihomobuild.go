package mihomobuild

import (
	"os"
	"os/exec"
)

type Kind string

const (
	KindDomain Kind = "domain"
	KindIPCIDR Kind = "ipcidr"
)

func Convert(mihomoBin string, kind Kind, inPath, outPath string) error {
	cmd := exec.Command(mihomoBin, "convert-ruleset", string(kind), "text", inPath, outPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
