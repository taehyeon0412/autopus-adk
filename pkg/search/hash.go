package search

import (
	"bufio"
	"fmt"
	"os"

	"github.com/cespare/xxhash/v2"
)

// HashLine은 파일의 한 줄과 그 해시이다.
type HashLine struct {
	LineNumber int    // 줄 번호 (1부터 시작)
	Content    string // 줄 내용
	Hash       string // xxhash 해시 (16진수 문자열)
}

// HashFile은 파일을 줄별로 읽어 해시를 계산한다.
// 출력 형식: LINE#HASH (예: "42#a1b2c3d4")
func HashFile(path string) ([]HashLine, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("파일 열기 실패 %q: %w", path, err)
	}
	defer f.Close()

	var lines []HashLine
	scanner := bufio.NewScanner(f)
	lineNum := 1

	for scanner.Scan() {
		content := scanner.Text()
		h := xxhash.Sum64String(content)
		lines = append(lines, HashLine{
			LineNumber: lineNum,
			Content:    content,
			Hash:       fmt.Sprintf("%x", h),
		})
		lineNum++
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("파일 읽기 실패: %w", err)
	}

	return lines, nil
}

// FormatHashLine은 HashLine을 "LINE#HASH" 형식 문자열로 반환한다.
func FormatHashLine(l HashLine) string {
	return fmt.Sprintf("%d#%s", l.LineNumber, l.Hash)
}
