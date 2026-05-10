package cmd

import (
	"fmt"

	"github.com/kdobmayer/kond-test-go-cli/config"
	"github.com/kdobmayer/kond-test-go-cli/output"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage pipeline configuration",
	Long:  `Get, set, and list configuration values.`,
}

var configGetCmd = &cobra.Command{
	Use:   "get [key]",
	Short: "Get a configuration value",
	Args:  cobra.ExactArgs(1),
	RunE:  runConfigGet,
}

var configSetCmd = &cobra.Command{
	Use:   "set [key] [value]",
	Short: "Set a configuration value",
	Args:  cobra.ExactArgs(2),
	RunE:  runConfigSet,
}

var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configuration values",
	RunE:  runConfigList,
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize configuration with defaults",
	RunE:  runConfigInit,
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configListCmd)
	configCmd.AddCommand(configInitCmd)
	configSetCmd.Flags().String("config-file", "", "Path to config file")
	configListCmd.Flags().Bool("json", false, "Output as JSON array")
}

func runConfigGet(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	val, err := cfg.Get(args[0])
	if err != nil {
		return err
	}

	fmt.Fprintln(cmd.OutOrStdout(), val)
	return nil
}

func runConfigSet(cmd *cobra.Command, args []string) error {
	configFile, _ := cmd.Flags().GetString("config-file")

	var cfg *config.Config
	var err error
	if configFile != "" {
		cfg, err = config.LoadFrom(configFile)
	} else {
		cfg, err = config.Load()
	}
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	if err := cfg.Set(args[0], args[1]); err != nil {
		return err
	}

	if configFile != "" {
		err = cfg.SaveTo(configFile)
	} else {
		err = cfg.Save()
	}
	if err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Set %s = %s\n", args[0], args[1])
	return nil
}

func runConfigList(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	jsonFlag, _ := cmd.Flags().GetBool("json")
	format := outputFormat
	if jsonFlag {
		format = "json"
	}
	formatter := output.NewFormatter(format, cmd.OutOrStdout())
	headers := []string{"KEY", "VALUE"}
	var rows []output.TableRow

	keys := cfg.ListKeys()
	type kvPair struct {
		Key   string `json:"key" yaml:"key"`
		Value string `json:"value" yaml:"value"`
	}
	var pairs []kvPair

	for _, key := range keys {
		val, _ := cfg.Get(key)
		rows = append(rows, output.TableRow{Columns: []string{key, val}})
		pairs = append(pairs, kvPair{Key: key, Value: val})
	}

	return formatter.Render(headers, rows, pairs)
}

func runConfigInit(cmd *cobra.Command, args []string) error {
	cfg := config.DefaultConfig()
	if err := cfg.Save(); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Configuration initialized at %s\n", config.ConfigPath())
	return nil
}
