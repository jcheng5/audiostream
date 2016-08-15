package main

import (
  "io"
  "log"
  "net/http"
  "os"
  "os/exec"
)

func installed(prog string) bool {
  cmd := exec.Command("which", prog)
  return cmd.Run() == nil
}

func main() {
  if !checkPrereqs() {
    os.Exit(1)
    return
  }

  if len(os.Args) < 2 {
    log.Fatal("Usage: client <url>")
  }

  url := os.Args[1]

  log.Println("Connecting to " + url)

  res, err := http.Get(url)
  if err != nil {
    log.Fatal(err)
  }
  defer res.Body.Close()

  cmd := createCommand()
  stdin, err := cmd.StdinPipe()
  if err != nil {
    log.Fatal(err)
  }
  err = cmd.Start()
  if err != nil {
    log.Fatal(err)
  }

  received := make(chan []byte)
  buf := make([]byte, 1024)

  // Start a goroutine for copying byte slices from "received"
  // and dropping them into stdin
  go transfer(received, stdin)

  // This loop receives bytes from the http connection
  // and writes the results to the "received" channel
  for {
    count, err := res.Body.Read(buf)
    if count > 0 {
      newSlice := make([]byte, count)
      copy(newSlice, buf[0:count])
      received <- newSlice
    }
    if err != nil {
      break
    }
  }

}

func transfer(recv chan []byte, stdin io.WriteCloser) {
  for {
    buf := <- recv
    _, err := stdin.Write(buf)
    if err != nil {
      log.Fatal(err)
    }
  }
}
