package main

import (
	"govirtfsstat/tar"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	cmd := cobra.Command{
		Use:  "",
		Args: cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			tarPath := args[0]
			destPath := args[1]

			tarReader, err := os.Open(tarPath)
			if err != nil {
				panic(err)
			}
			defer tarReader.Close()
			err = tar.Untar(tarReader, destPath)
			if err != nil {
				panic(err)
			}
		},
	}

	cmd.Execute()
}
