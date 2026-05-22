package adapter

import (
	"bufio"
	"errors"
	"io"
	"strings"
)

// Event 是单个 SSE 事件。多行 data 用 \n 连接。
type Event struct {
	Event string
	Data  string
	ID    string
}

// SSEReader 按 SSE 协议从 stream 中按 event 块读取。
//
// 一个 event 块以一个空行结束;每行内部 "field: value"。
type SSEReader struct {
	br *bufio.Reader
}

// NewSSEReader 构造一个 SSE 行读取器。默认缓冲 64KB,可自动增长。
func NewSSEReader(r io.Reader) *SSEReader {
	return &SSEReader{br: bufio.NewReaderSize(r, 64*1024)}
}

// Next 读取下一个 event。返回 io.EOF 表示流结束。
func (r *SSEReader) Next() (*Event, error) {
	e := &Event{}
	has := false
	for {
		line, err := r.br.ReadString('\n')
		if err != nil {
			// EOF 时如果已有累积内容,先返回它,下次再 EOF
			if errors.Is(err, io.EOF) {
				if line != "" {
					if parsed := parseSSELine(strings.TrimRight(line, "\r\n"), e); parsed {
						has = true
					}
				}
				if has {
					return e, nil
				}
			}
			return nil, err
		}
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			if has {
				return e, nil
			}
			continue
		}
		if strings.HasPrefix(line, ":") {
			continue // 注释
		}
		if parseSSELine(line, e) {
			has = true
		}
	}
}

// parseSSELine 解析一行 SSE field;返回是否产出了内容。
func parseSSELine(line string, e *Event) bool {
	i := strings.IndexByte(line, ':')
	if i < 0 {
		return false
	}
	field := line[:i]
	val := strings.TrimPrefix(line[i+1:], " ")
	switch field {
	case "data":
		if e.Data != "" {
			e.Data += "\n" + val
		} else {
			e.Data = val
		}
		return true
	case "event":
		e.Event = val
		return true
	case "id":
		e.ID = val
		return true
	}
	return false
}
