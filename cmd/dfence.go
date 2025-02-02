package cmd

import (
	"fmt"
	"os"

	"github.com/chavacava/dfence/internal/infra"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var logLevel string

var rootCmd = &cobra.Command{
	Use:   "dfence",
	Short: "Dependency fences",
	Long: color.New(color.FgHiYellow).Sprintf(`
         ________                   
    ____/ / ____/__  ____  ________ 
   / __  / /_  / _ \/ __ \/ ___/ _ \
  / /_/ / __/ /  __/ / / / /__/  __/
  \__,_/_/    \___/_/ /_/\___/\___/ 
																		 
  Understand and control your dependencies`),

	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		viper.Set("logger", buildlogger(logLevel))
	},
}

// Execute executes this command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize()
	rootCmd.PersistentFlags().StringVar(&logLevel, "log", "info", "log level: none, error, warn, info, debug")
}

func buildlogger(level string) infra.Logger {
	nop := func(string, ...interface{}) {}
	debug, info, warn, err := nop, nop, nop, nop
	switch level {
	case "none":
		// do nothing
	case "debug":
		debug = buildLoggerFunc("[DEBUG] ", color.New(color.FgCyan))
		fallthrough
	case "info":
		info = buildLoggerFunc("", color.New(color.FgGreen))
		fallthrough
	case "warn":
		warn = buildLoggerFunc("", color.New(color.FgHiYellow))
		fallthrough
	default:
		err = buildLoggerFunc("", color.New(color.FgHiRed))
	}

	fatal := buildLoggerFunc("", color.New(color.BgRed))
	return infra.NewLogger(debug, info, warn, err, fatal)
}

func buildLoggerFunc(prefix string, c *color.Color) infra.LoggerFunc {
	return func(msg string, vars ...interface{}) {
		fmt.Println(c.Sprintf(prefix+msg, vars...))
	}
}
