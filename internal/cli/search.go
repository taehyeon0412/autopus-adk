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

// newDocsCmd는 docs 서브커맨드를 생성한다.
func newDocsCmd() *cobra.Command {
	var topic string

	cmd := &cobra.Command{
		Use:   "docs <library>",
		Short: "Lookup library documentation via Context7",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client := search.NewContext7Client()

			// 라이브러리 ID 조회
			id, err := client.ResolveLibrary(args[0])
			if err != nil {
				return fmt.Errorf("라이브러리 조회 실패: %w", err)
			}

			// 문서 조회
			docs, err := client.GetDocs(id, topic)
			if err != nil {
				return fmt.Errorf("문서 조회 실패: %w", err)
			}

			fmt.Println(docs)
			return nil
		},
	}

	cmd.Flags().StringVar(&topic, "topic", "", "Documentation topic to focus on")
	return cmd
}
