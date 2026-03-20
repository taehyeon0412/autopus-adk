package lore

import (
	"fmt"
	"strings"
)

// BuildCommit는 LoreEntry와 메시지로 git commit 메시지를 생성한다.
func BuildCommit(entry *LoreEntry, message string) (string, error) {
	if message == "" {
		return "", fmt.Errorf("커밋 메시지는 비어있을 수 없습니다")
	}

	trailers := FormatTrailers(entry)
	if trailers == "" {
		return message, nil
	}

	// 메시지와 트레일러 사이에 빈 줄 추가 (git 컨벤션)
	msg := strings.TrimRight(message, "\n")
	return msg + "\n\n" + trailers, nil
}
