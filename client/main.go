package main

import (
  "io"
  "log"
  "net/http"
  "os"
  "os/exec"
  "time"
)

type WriterFunc func(p []byte) (int, error)

func (w WriterFunc) Write(p []byte)(int, error) {
  return w(p)
}

func installed(prog string) bool {
  cmd := exec.Command("which", prog)
  return cmd.Run() == nil
}

type Metrics struct {
  ticker *time.Ticker
  receiveData chan int
  entries []float32
  pos int
  totalEntries int
  currentBytes int
  currentStartTime time.Time
}
func NewMetrics(duration time.Duration, maxEntries int) *Metrics {
  ticker := time.NewTicker(duration)

  m := &Metrics{ticker, make(chan int), make([]float32, maxEntries), -1, 0, 0, time.Now()}

  go func() {
    for {
      select {
      case <-ticker.C:
        if m.pos < 0 {
          m.pos = 0
        } else {
          secs := float32(time.Now().Sub(m.currentStartTime).Seconds())
          m.entries[m.pos] = float32(m.currentBytes) / 1024 * 8 / secs
          log.Printf("%f", m.entries[m.pos])
          m.pos = (m.pos + 1) % len(m.entries)
          m.totalEntries += 1
        }
        m.currentBytes = 0
        m.currentStartTime = time.Now()
      case bytes := <-m.receiveData:
        m.currentBytes += bytes
      }
    }
  }()

  return m
}
func (m *Metrics) Write(p []byte) (int, error) {
  m.receiveData <- len(p)
  return len(p), nil
}
func (m *Metrics) Entries() []float32 {
  if m.totalEntries == m.pos {
    entries := make([]float32, m.pos)
    copy(entries, m.entries[0:m.pos])
    return entries
  } else {
    entries := make([]float32, len(m.entries))
    copy(entries, m.entries[m.pos:])
    copy(entries[len(m.entries[m.pos:]):], m.entries[0:m.pos])
    return entries
  }
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

  metrics := NewMetrics(1 * time.Second, 3 * 60 * 60)
  sink := io.MultiWriter(stdin, metrics)
  for ; err == nil; _, err = io.CopyN(sink, res.Body, 1024) {
  }
  if err != io.EOF {
    log.Fatal(err)
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
