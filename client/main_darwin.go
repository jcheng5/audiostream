package main

import (
  "log"
  "os/exec"
)

func checkPrereqs() bool {
  if !installed("mpg123") {
    log.Fatal("mpg123 not installed; try `brew install mpg123`")
    return false
  }
  return true
}

func createCommand() *exec.Cmd {
  return exec.Command("mpg123", "-")
}
