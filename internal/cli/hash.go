package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/insajin/autopus-adk/pkg/search"
)

// newHashCmd는 hash 서브커맨드를 생성한다.
func newHashCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "hash <file>",
		Short: "Hash each line of a file using xxhash",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			lines, err := search.HashFile(args[0])
			if err != nil {
				return fmt.Errorf("파일 해시 실패: %w", err)
			}

			for _, l := range lines {
				fmt.Println(search.FormatHashLine(l))
			}
			return nil
		},
	}
}
