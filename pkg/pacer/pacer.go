// Package pacer implements a rate controlled reader or writer.
// You wrap a reader or write in a pacer and it goes at the rate you want.
// Read or Write will block as needed to implement the rate control.
// It can burst data for each read or write possibly beyond the rate but
// the doesn't allow the next read or write to occur until the rate
// will allow.
//
// There are a million ways to do rate limiting but we just do the
// well-known and simple virtual clock algorithm here.
//
// This is useful for testing.
//
package pacer

import (
	"io"
	"time"
)

// Pacer xxx
type Pacer struct {
	bytesPerSecond int
	clock          time.Time
}

func (p *Pacer) pace(cc int) {
	now := time.Now()
	if p.clock.IsZero() {
		p.clock = now
	}
	delay := time.Duration((int(time.Second) * cc) / p.bytesPerSecond)
	p.clock = p.clock.Add(delay)
	if p.clock.After(now) {
		time.Sleep(p.clock.Sub(now))
	}
}

// ReaderPacer is an io.Reader with a rate limit
type ReaderPacer struct {
	Pacer
	reader io.Reader
}

// WriterPacer is an io.Writer with a rate limit
type WriterPacer struct {
	Pacer
	writer io.Writer
}

// NewReaderPacer wraps the writer and limits its writing to the
// rate in bytes-per-second indicated.
func NewReaderPacer(r io.Reader, rate int) *ReaderPacer {
	rp := &ReaderPacer{reader: r}
	rp.bytesPerSecond = rate
	return rp
}

func (p *ReaderPacer) Read(b []byte) (int, error) {
	cc, err := p.reader.Read(b)
	if err != nil {
		return cc, err
	}
	p.pace(cc)
	return cc, nil
}

// NewWriterPacer wraps the writer and limits its writing to the
// rate in bytes-per-second indicated.
func NewWriterPacer(w io.Writer, rate int) *WriterPacer {
	wp := &WriterPacer{writer: w}
	wp.bytesPerSecond = rate
	return wp
}

func (p *WriterPacer) Write(b []byte) (int, error) {
	cc, err := p.writer.Write(b)
	if err != nil {
		return cc, err
	}
	p.pace(cc)
	return cc, nil
}
