package harness

import (
	"bufio"
	"context"
	"encoding/csv"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Feed is a fake DDE currency server: takes the auth token, then pushes frames to every client.
type Feed struct {
	ln    net.Listener
	mu    sync.Mutex
	conns []net.Conn
	last  map[string]string
}

// StartFeed listens on port and accepts in the background.
func StartFeed(port int) (*Feed, error) {
	ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		return nil, err
	}
	f := &Feed{ln: ln, last: map[string]string{}}
	go f.accept()
	return f, nil
}

func (f *Feed) accept() {
	for {
		conn, err := f.ln.Accept()
		if err != nil {
			return
		}
		go f.handshake(conn)
	}
}

// handshake reads the one-shot token, then replays the last known prices so the client sees data at once.
func (f *Feed) handshake(conn net.Conn) {
	_ = conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	if _, err := bufio.NewReader(conn).ReadString('!'); err != nil {
		_ = conn.Close()
		return
	}
	_ = conn.SetReadDeadline(time.Time{})
	f.mu.Lock()
	f.conns = append(f.conns, conn)
	for _, frame := range f.last {
		_, _ = conn.Write([]byte(frame))
	}
	f.mu.Unlock()
}

// Step pushes one quote for a symbol to every connected client.
func (f *Feed) Step(symbol string, bid, ask float64) {
	chunk := make([]string, 16)
	chunk[0], chunk[1], chunk[2] = symbol, fmt.Sprintf("%.5f", bid), fmt.Sprintf("%.5f", ask)
	chunk[3], chunk[4], chunk[6], chunk[8] = chunk[2], chunk[1], chunk[1], chunk[1]
	frame := "MRKTDATAs?<G)_" + time.Now().UTC().Format("2006-01-02 15:04:05") + "," + strings.Join(chunk, ",") + "!"

	f.mu.Lock()
	defer f.mu.Unlock()
	f.last[symbol] = frame
	alive := f.conns[:0]
	for _, c := range f.conns {
		if _, err := c.Write([]byte(frame)); err == nil {
			alive = append(alive, c)
		} else {
			_ = c.Close()
		}
	}
	f.conns = alive
}

// Path pushes a sequence of quotes with a pause between them.
func (f *Feed) Path(symbol string, quotes [][2]float64, pause time.Duration) {
	for _, q := range quotes {
		f.Step(symbol, q[0], q[1])
		time.Sleep(pause)
	}
}

// Connected says whether hst-quote currently holds a session.
func (f *Feed) Connected() bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.conns) > 0
}

// Close stops the listener and drops the clients.
func (f *Feed) Close() {
	_ = f.ln.Close()
	f.mu.Lock()
	for _, c := range f.conns {
		_ = c.Close()
	}
	f.mu.Unlock()
}

// Replay streams every data/ticks/<SYMBOL>.csv (time,bid,ask) through the feed at speed× real time until ctx ends.
func (f *Feed) Replay(ctx context.Context, dir string, speed float64) error {
	files, err := filepath.Glob(filepath.Join(dir, "*.csv"))
	if err != nil || len(files) == 0 {
		return fmt.Errorf("no tick files in %s", dir)
	}
	for _, file := range files {
		symbol := strings.TrimSuffix(filepath.Base(file), ".csv")
		rows, err := readTicks(file)
		if err != nil {
			return err
		}
		go func() {
			for i, r := range rows {
				if i > 0 {
					wait := time.Duration(float64(r.at-rows[i-1].at) * float64(time.Second) / speed)
					select {
					case <-ctx.Done():
						return
					case <-time.After(wait):
					}
				}
				f.Step(symbol, r.bid, r.ask)
			}
		}()
	}
	return nil
}

type tickRow struct {
	at       int64
	bid, ask float64
}

// readTicks parses a time,bid,ask csv with a header line.
func readTicks(file string) ([]tickRow, error) {
	fh, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer fh.Close()
	recs, err := csv.NewReader(fh).ReadAll()
	if err != nil {
		return nil, err
	}
	rows := make([]tickRow, 0, len(recs))
	for i, rec := range recs {
		if i == 0 || len(rec) < 3 {
			continue
		}
		at, _ := strconv.ParseInt(rec[0], 10, 64)
		bid, _ := strconv.ParseFloat(rec[1], 64)
		ask, _ := strconv.ParseFloat(rec[2], 64)
		rows = append(rows, tickRow{at, bid, ask})
	}
	return rows, nil
}
