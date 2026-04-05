package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/insajin/autopus-adk/pkg/search"
)

// newSearchCmd는 search 서브커맨드를 생성한다.
func newSearchCmd() *cobra.Command {
	var numResults int

	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search the web using Exa API (requires EXA_API_KEY)",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			query := ""
			for i, a := range args {
				if i > 0 {
					query += " "
				}
				query += a
			}

			client := search.NewExaClientFromEnv()
			results, err := client.Search(query, numResults)
			if err != nil {
				return fmt.Errorf("검색 실패: %w", err)
			}

			if len(results) == 0 {
				fmt.Println("검색 결과 없음")
				return nil
			}

			for i, r := range results {
				fmt.Printf("\n[%d] %s\n    %s\n    %s\n", i+1, r.Title, r.URL, r.Snippet)
			}
			return nil
		},
	}

	cmd.Flags().IntVar(&numResults, "num", 5, "Number of results to return")
	return cmd
}
