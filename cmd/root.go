package cmd

import (
	"fmt"
	"os"

	"github.com/FacileStudio/tiroir/lib"
	"github.com/spf13/cobra"
)

var Version = "dev"

var rootCmd = &cobra.Command{
	Use:     "tiroir",
	Short:   "Encrypted env manager applied into a shell",
	Long:    `tiroir keeps shell variables encrypted at rest; 'tiroir export' emits export statements a shell applies into the current environment.`,
	Version: Version,
	SilenceUsage:  true,
	SilenceErrors: true,
	CompletionOptions: cobra.CompletionOptions{HiddenDefaultCmd: true},
	Run: func(cmd *cobra.Command, args []string) {
		if err := cmd.Help(); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	},
}

// Execute runs the CLI and returns the first error, if any.
func Execute() error {
	return rootCmd.Execute()
}

func store() (*env.EnvStore, error) {
	return env.New()
}

// init builds the command tree. `set` gets non-interspersed flag parsing so
// that `set KEY VALUE` accepts a VALUE starting with a dash: pflag would
// otherwise read `-value` as an unknown flag and reject it before the command
// runs, however the shell quoted it. `tiroir set -- KEY -value` still covers a
// key that itself starts with a dash.
func init() {
	rootCmd.SetVersionTemplate("{{.Name}} {{.Version}}\n")
	setCmd.Flags().SetInterspersed(false)
	rootCmd.AddCommand(getCmd, setCmd, listCmd, deleteCmd, exportCmd)
}