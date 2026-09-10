package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get KEY",
	Short: "Print a value from the store",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		s, err := store()
		if err != nil {
			return err
		}
		val, ok, err := s.Get(args[0])
		if err != nil {
			return err
		}
		if !ok {
			return fmt.Errorf("tiroir: key %q not set", args[0])
		}
		fmt.Println(val)
		return nil
	},
}

var setCmd = &cobra.Command{
	Use:   "set KEY VALUE",
	Short: "Write a value to the store",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		s, err := store()
		if err != nil {
			return err
		}
		return s.Set(args[0], args[1])
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List keys in the store",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		s, err := store()
		if err != nil {
			return err
		}
		keys, err := s.List()
		if err != nil {
			return err
		}
		for _, k := range keys {
			fmt.Println(k)
		}
		return nil
	},
}

var deleteCmd = &cobra.Command{
	Use:   "delete KEY",
	Short: "Remove a key from the store",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		s, err := store()
		if err != nil {
			return err
		}
		return s.Delete(args[0])
	},
}

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Print the store as KEY=value lines",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		s, err := store()
		if err != nil {
			return err
		}
		lines, err := s.Export()
		if err != nil {
			return err
		}
		for _, line := range lines {
			fmt.Fprintln(os.Stdout, line)
		}
		return nil
	},
}