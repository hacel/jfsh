package mpv

import "os/exec"

// TODO: finish the rest of this function
func Play(filename string) {
	c := exec.Command("mpv", filename)
	c.Run()
}
