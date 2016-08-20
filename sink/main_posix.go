package main

import (
  "log"
  "os/exec"
)

func checkPrereqs() bool {
  if !installed("sox") {
    log.Fatal("sox not installed; try `apt install sox`")
    return false
  }
  return true
}

func createCommand() *exec.Cmd {
  return exec.Command("sox", "-t", "wav", "-", "-d")
}
