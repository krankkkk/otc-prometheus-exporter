package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	"github.com/iits-consulting/otc-prometheus-exporter/grafana"
	"github.com/spf13/cobra"
)

func main() {
	var outputPath string
	var rootCmd = &cobra.Command{
		Use:   "grafanadashboards",
		Short: "Generates Grafana dashboards from provider dashboard metadata.",
		Run: func(cmd *cobra.Command, args []string) {
			if err := os.MkdirAll(outputPath, 0755); err != nil {
				log.Fatalf("Could not create output directory %s: %v\n", outputPath, err)
			}

			for _, cfg := range allDashboardConfigs() {
				board := grafana.GenerateDashboard(cfg)

				b, err := json.MarshalIndent(board, "", "  ")
				if err != nil {
					log.Fatalf("Could not marshal dashboard %s: %v\n", cfg.Title, err)
				}

				filename := strings.ToLower(strings.ReplaceAll(cfg.UID, "otc-", "")) + ".json"
				outputFile := path.Join(outputPath, filename)
				if err := os.WriteFile(outputFile, b, 0644); err != nil {
					log.Fatalf("Could not write %s: %v\n", outputFile, err)
				}
				fmt.Printf("Generated %s\n", outputFile)
			}
		},
	}
	rootCmd.Flags().StringVar(&outputPath, "output-path", "", "Directory for generated dashboards.")
	rootCmd.MarkFlagRequired("output-path") //nolint:errcheck

	if err := rootCmd.Execute(); err != nil {
		log.Fatalln(err)
	}
}
