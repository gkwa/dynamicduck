package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gkwa/dynamicduck/internal/config"
	"github.com/gkwa/dynamicduck/internal/logger"
	"github.com/gkwa/dynamicduck/internal/model"
	"github.com/gkwa/dynamicduck/internal/parser"
	"github.com/gkwa/dynamicduck/internal/sampler"
	"github.com/spf13/cobra"
)

var cfg = config.NewConfig()

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "dynamicduck",
	Short: "A tool for randomly sampling JSON data",
	Long: `dynamicduck is a command-line tool for processing a JSON dataset, 
randomly selecting items, and outputting the same shape to stdout or file.

It reads data from stdin or an input file, randomly selects N items from 
the Items array, and outputs the result in the same JSON format.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		// Check for mutually exclusive flags
		if cfg.Seed != 0 && cfg.SeenFile != "" {
			return fmt.Errorf("cannot use both --seed and --seen-file together, as they serve different purposes")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		log := logger.NewLogger(cfg.VerbosityLevel)

		log.Debug("Starting dynamicduck")
		log.Debug(fmt.Sprintf("Configuration: %+v", cfg))

		// Read and parse the input
		log.Verbose("Reading input data")
		data, err := parser.ReadInput(cfg.InputFile)
		if err != nil {
			return fmt.Errorf("failed to read input: %w", err)
		}

		// Parse the JSON data
		log.Verbose("Parsing JSON data")
		jsonData, err := parser.ParseJSON(data)
		if err != nil {
			return fmt.Errorf("failed to parse JSON: %w", err)
		}

		log.VeryVerbose(fmt.Sprintf("Found %d items in input", len(jsonData.Items)))

		// Validate the data
		if len(jsonData.Items) == 0 {
			log.Error("No items found in the input data")
			return fmt.Errorf("no items found in input data")
		}

		if cfg.Count <= 0 {
			log.Error("Count must be greater than 0")
			return fmt.Errorf("count must be greater than 0")
		}

		// Check for seen file
		if cfg.SeenFile != "" {
			log.Verbose(fmt.Sprintf("Using seen items file: %s", cfg.SeenFile))

			// Ensure the directory for the seen file exists
			seenDir := filepath.Dir(cfg.SeenFile)
			if err := os.MkdirAll(seenDir, 0o755); err != nil {
				return fmt.Errorf("failed to create directory for seen items file: %w", err)
			}
		}

		// Sample the items
		log.Verbose(fmt.Sprintf("Randomly selecting %d items", cfg.Count))
		if cfg.Seed != 0 {
			log.Verbose(fmt.Sprintf("Using seed: %d", cfg.Seed))
		}

		sampledData, err := sampler.SampleItems(jsonData, cfg.Count, cfg.Seed, cfg.SeenFile)
		if err != nil {
			// If we couldn't sample any items because all have been seen,
			// log a specific message but don't fail the command
			if cfg.SeenFile != "" && err.Error() != "" && len(sampledData.Items) == 0 {
				log.Error(fmt.Sprintf("All items have been seen already: %v", err))
				log.Verbose("Consider removing or resetting the seen items file to start over")

				// Return empty result
				sampledData = &model.JSONData{
					Items:            []map[string]interface{}{},
					Count:            0,
					ScannedCount:     jsonData.ScannedCount,
					ConsumedCapacity: jsonData.ConsumedCapacity,
				}
			} else {
				return fmt.Errorf("failed to sample items: %w", err)
			}
		}

		// Warn if we selected fewer items than requested
		if len(sampledData.Items) < cfg.Count {
			log.Verbose(fmt.Sprintf("Selected only %d items (requested %d) due to lack of unseen items",
				len(sampledData.Items), cfg.Count))
		}

		// Output the result
		log.Verbose("Generating output")
		output, err := parser.GenerateOutput(sampledData)
		if err != nil {
			return fmt.Errorf("failed to generate output: %w", err)
		}

		if cfg.OutputFile != "" {
			log.Verbose(fmt.Sprintf("Writing output to file: %s", cfg.OutputFile))
			// Ensure output directory exists
			outputDir := filepath.Dir(cfg.OutputFile)
			if err := os.MkdirAll(outputDir, 0o755); err != nil {
				return fmt.Errorf("failed to create output directory: %w", err)
			}

			err = os.WriteFile(cfg.OutputFile, output, 0o644)
			if err != nil {
				return fmt.Errorf("failed to write output file: %w", err)
			}
		} else {
			// Write to stdout
			log.Debug("Writing output to stdout")
			fmt.Print(string(output))
		}

		log.Debug("dynamicduck completed successfully")
		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().IntVar(&cfg.Count, "count", 10, "Number of random items to select")
	rootCmd.Flags().StringVar(&cfg.InputFile, "input", "", "Read from file instead of stdin")
	rootCmd.Flags().StringVar(&cfg.OutputFile, "out-file", "", "Write to file instead of stdout")
	rootCmd.Flags().Int64Var(&cfg.Seed, "seed", 0, "Seed for random number generation (0 = use time)")
	rootCmd.Flags().StringVar(&cfg.SeenFile, "seen-file", "", "File to track seen items to avoid duplicates")
	rootCmd.Flags().CountVarP(&cfg.VerbosityLevel, "verbose", "v", "Increase verbosity level")
}
