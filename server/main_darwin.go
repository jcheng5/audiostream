package main

import (
  "log"
  "os/exec"
)

func checkPrereqs() bool {
  if !installed("sox") {
    log.Fatal("sox not installed; try `brew install sox`")
    return false
  }
  if !installed("lame") {
    log.Fatal("lame not installed; try `brew install lame`")
    return false
  }
  return true
}

func createCommand() *exec.Cmd {
  return exec.Command("/bin/bash", "-c", "sox -d -t wav - 2>/dev/null | lame - -")
}
