package benchrunner

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// scannerBufferSize caps the longest single token (one SQL statement, one
// Mongo command, one Redis SET) the scanner will hold. 4 MB is overkill for
// any one statement but cheap.
const scannerBufferSize = 4 * 1024 * 1024

// streamQueries reads newline-separated commands (Mongo .js, Redis .txt) and
// invokes onLine for the first `limit` non-empty, non-comment lines.
// limit <= 0 means "consume the whole file". Lines are trimmed and stripped
// of a trailing `;`.
func streamQueries(queryFile string, limit int, onLine func(line string) error) (int, error) {
	return streamSplit(queryFile, limit, bufio.ScanLines, onLine)
}

// streamSQLStatements reads `;`-separated SQL statements. The generator emits
// every statement on a single line, joined by `;`, so bufio.ScanLines would
// either return one multi-GB token or fail with "token too long". Splitting on
// `;` keeps each token small and lets the runner execute one statement per
// Exec call (without needing multiStatements=true on the driver).
func streamSQLStatements(queryFile string, limit int, onLine func(line string) error) (int, error) {
	return streamSplit(queryFile, limit, scanSemicolon, onLine)
}

// FirstSQLStatement returns the first non-comment SQL statement from a file
// with semicolon separators. Used by the EXPLAIN collector to pick a
// representative query.
func FirstSQLStatement(queryFile string) (string, error) {
	var first string
	_, err := streamSQLStatements(queryFile, 1, func(s string) error {
		first = s
		return nil
	})
	return first, err
}

func streamSplit(queryFile string, limit int, split bufio.SplitFunc, onLine func(string) error) (int, error) {
	f, err := os.Open(queryFile)
	if err != nil {
		return 0, fmt.Errorf("open %s: %w", queryFile, err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 64*1024), scannerBufferSize)
	scanner.Split(split)

	count := 0
	for scanner.Scan() {
		if limit > 0 && count >= limit {
			break
		}
		line := strings.TrimSpace(scanner.Text())
		if line == "" ||
			strings.HasPrefix(line, "--") ||
			strings.HasPrefix(line, "//") ||
			strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimSuffix(line, ";")
		if err := onLine(line); err != nil {
			return count, err
		}
		count++
	}
	if err := scanner.Err(); err != nil {
		return count, fmt.Errorf("scan %s: %w", queryFile, err)
	}
	return count, nil
}

// scanSemicolon is a bufio.SplitFunc that returns tokens separated by `;`.
// Tokens may span newlines — we explicitly do not break on `\n` because some
// generated SQL files have no newlines at all (the whole file is one logical
// line of `;`-joined statements).
func scanSemicolon(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := indexByte(data, ';'); i >= 0 {
		return i + 1, data[:i], nil
	}
	if atEOF {
		return len(data), data, nil
	}
	// Need more data to find the next `;`.
	return 0, nil, nil
}

func indexByte(data []byte, b byte) int {
	for i, c := range data {
		if c == b {
			return i
		}
	}
	return -1
}
