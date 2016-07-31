package main

import (
  "io"
  "log"
  "net/http"
  "os"
  "os/exec"
)

func main() {
  if !checkPrereqs() {
    os.Exit(1)
    return
  }

  log.Fatal(http.ListenAndServe(":8080", http.HandlerFunc(handler)))
}

func installed(prog string) bool {
  cmd := exec.Command("which", prog)
  return cmd.Run() == nil
}

func handler(w http.ResponseWriter, r *http.Request) {
  log.Println("Client " + r.RemoteAddr + " connected")

  w.Header().Add("Content-Type", "audio/mpeg")

  cmd := createCommand()

  mp3data, err := cmd.StdoutPipe()
  if err != nil {
    log.Println("Failed to get pipe")
    w.WriteHeader(500)
    return
  }

  err = cmd.Start()
  if err != nil {
    log.Println("Failed to start program")
    w.WriteHeader(500)
    return
  }

  cn, ok := w.(http.CloseNotifier)
  if !ok {
    log.Println("Failed to cast response to CloseNotifier")
    w.WriteHeader(500)
    return
  }

  closeNotifyChan := cn.CloseNotify()
  go func() {
    for {
      <-closeNotifyChan
      cmd.Process.Kill()
    }
  }()

  io.Copy(w, mp3data)

  cmd.Wait()

  log.Println("Client " + r.RemoteAddr + " disconnected")
}
