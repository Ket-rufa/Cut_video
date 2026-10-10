package ffmpeg

import (
	"bufio"
	"bytes"
	"io"
	"regexp"
	"strconv"
)

var (
	timeRegex     = regexp.MustCompile(`time=(\d+):(\d+):(\d+)\.(\d+)`)
	durationRegex = regexp.MustCompile(`Duration:\s+(\d+):(\d+):(\d+)\.(\d+)`)
)

// scanCRLFLines is a bufio.SplitFunc that splits on \n or \r, trimming both.
// FFmpeg writes progress lines terminated with \r (overwriting the current line)
// when stderr is a pipe, or \r\n on some platforms. This handles both.
func scanCRLFLines(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	// Find \n or \r
	if i := bytes.IndexAny(data, "\r\n"); i >= 0 {
		// Skip any adjacent \r\n to avoid empty tokens
		end := i + 1
		for end < len(data) && (data[end] == '\r' || data[end] == '\n') && end == i+1 {
			// Only skip the paired \n after a \r (i.e., \r\n)
			if data[i] == '\r' && data[end] == '\n' {
				end++
			}
			break
		}
		return end, bytes.TrimRight(data[:i], "\r\n "), nil
	}
	// No newline yet — if at EOF, return what we have
	if atEOF {
		return len(data), bytes.TrimRight(data, "\r\n "), nil
	}
	return 0, nil, nil
}

// ParseProgress reads FFmpeg stderr and emits real progress values (0-99)
// by parsing the "time=HH:MM:SS.ms" fields FFmpeg writes during encoding.
//
// If totalDurationSec <= 0, the function auto-detects the source duration
// from the "Duration: HH:MM:SS.ms" line printed at startup.
//
// fn is called each time a new progress value is computed. Reaching 100
// is left to the caller after the FFmpeg process exits successfully.
func ParseProgress(r io.Reader, totalDurationSec float64, fn func(int)) {
	scanner := bufio.NewScanner(r)
	scanner.Split(scanCRLFLines)
	scanner.Buffer(make([]byte, 64*1024), 64*1024)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		// Auto-detect total duration from FFmpeg header
		if totalDurationSec <= 0 {
			if m := durationRegex.FindStringSubmatch(line); len(m) >= 5 {
				h, _ := strconv.ParseFloat(m[1], 64)
				min, _ := strconv.ParseFloat(m[2], 64)
				sec, _ := strconv.ParseFloat(m[3], 64)
				totalDurationSec = h*3600 + min*60 + sec
			}
		}

		// Parse current encoding position
		m := timeRegex.FindStringSubmatch(line)
		if m == nil || len(m) < 5 {
			continue
		}
		h, _ := strconv.ParseFloat(m[1], 64)
		min, _ := strconv.ParseFloat(m[2], 64)
		sec, _ := strconv.ParseFloat(m[3], 64)
		currentSec := h*3600 + min*60 + sec

		if totalDurationSec > 0 && fn != nil {
			pct := int(currentSec / totalDurationSec * 100)
			if pct > 99 {
				pct = 99
			}
			fn(pct)
		}
	}
}
