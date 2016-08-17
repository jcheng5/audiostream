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

  log.Println("Joining group 224.0.0.118:200")

  addr, err := net.ResolveUDPAddr("udp", "224.0.0.118:200")
  if err != nil {
    log.Println("Failed to resolve multicast group address")
    return
  }

  ifi, err := net.InterfaceByName(os.Args[1])
  if err != nil {
    log.Println("Couldn't retrieve network iface '", os.Args[1], "': ", err)
    return
  }
  conn, err := net.ListenMulticastUDP("udp", ifi, addr)
  if err != nil {
    log.Println("Failure listening to multicast group: ", err)
    return
  }
  defer conn.Close()

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

  var maxCounter uint64 = 0
  // This loop receives bytes from the http connection
  // and writes the results to the "received" channel
  for {
    count, _, err := conn.ReadFromUDP(buf)
    log.Println("Received")
    if count > 0 {
      counter, n := binary.Uvarint(buf[0:8])
      if n <= 0 {
        log.Println("Malformed counter")
      }
      if maxCounter >= counter {
        log.Println("Packet out of order: ", counter, " <= ", maxCounter)
      } else {
        maxCounter = counter
      }

      newSlice := make([]byte, count-8)
      copy(newSlice, buf[8:count])
      received <- newSlice
    }
    if err != nil {
      break
    }
  }

}

func transfer(recv chan []byte, stdin io.WriteCloser) {
  for {
    buf := <-recv
    _, err := stdin.Write(buf)
    if err != nil {
      log.Fatal(err)
    }
  }
}
