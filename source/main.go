package main

import (
  "encoding/binary"
  "io"
  "log"
  "net"
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

  cmd := createCommand()

  mp3data, err := cmd.StdoutPipe()
  if err != nil {
    log.Println("Failed to get pipe")
    return
  }

  err = cmd.Start()
  if err != nil {
    log.Println("Failed to start program")
    return
  }
  defer cmd.Process.Kill()

  addr, err := net.ResolveUDPAddr("udp", "224.0.0.118:2016")
  if err != nil {
    log.Println("Failed to resolve multicast group address")
    return
  }
  conn, err := net.DialUDP("udp", nil, addr)
  if err != nil {
    log.Println("Error creating udp connection: ", err)
    return
  }
  defer conn.Close()

  var counter uint64 = 0
  buf := make([]byte, 8192 + 8)
  for {
    binary.PutUvarint(buf[0:8], counter)
    _, err = io.ReadFull(mp3data, buf[8:])
    if err != nil {
      // TODO: Restart?
      return
    }
    _, err = conn.Write(buf)
    if err != nil {
      log.Println("Failure on write: ", err)
      return
    }
    counter++
  }

  cmd.Wait()
}
