package lore

import (
	"fmt"
	"strings"
)

const signOff = "🐙 Autopus <noreply@autopus.co>"

// BuildCommit는 LoreEntry와 메시지로 git commit 메시지를 생성한다.
func BuildCommit(entry *LoreEntry, message string) (string, error) {
	msg := strings.TrimSpace(message)
	if msg == "" {
		return "", fmt.Errorf("커밋 메시지는 비어있을 수 없습니다")
	}

	msg = strings.TrimSpace(strings.TrimSuffix(msg, signOff))
	trailers := FormatTrailers(entry)

	if trailers == "" {
		return msg + "\n\n" + signOff, nil
	}

	// 메시지와 트레일러 사이에 빈 줄 추가 (git 컨벤션)
	return msg + "\n\n" + trailers + "\n\n" + signOff, nil
}
