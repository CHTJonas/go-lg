package web

import (
	"log"
	"os/exec"
	"syscall"
)

func run(cmd *exec.Cmd) []byte {
	stdout, err := cmd.Output()
	if err != nil {
		log.Println("error executing command", cmd)
		if exitErr, k := err.(*exec.ExitError); k {
			if exitErr.Exited() {
				log.Printf("PID %d exited with code %d\n%s", exitErr.Pid(), exitErr.ExitCode(), string(exitErr.Stderr))
			} else {
				if ws, l := exitErr.Sys().(syscall.WaitStatus); l {
					log.Println("PID", exitErr.Pid(), "was signalled:", ws.Signal())
				} else {
					log.Println("PID", exitErr.Pid(), "exited due to signal")
				}
			}
		} else {
			log.Println(err)
		}
	}
	return stdout
}
