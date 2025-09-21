package main

import (
	"fmt"
	"govirtfsstat/stat"
	"strconv"
	"syscall"

	"github.com/spf13/cobra"
)

func main() {
	rootCmd := cobra.Command{}

	statCmd := cobra.Command{
		Use:  "stat <path>",
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			path := args[0]

			st := stat.Stat(path)
			fmt.Printf("st: %+v\n", st)
			switch st.Mode & syscall.S_IFMT {
			case syscall.S_IFDIR:
				fmt.Println("dir")
			case syscall.S_IFLNK:
				fmt.Println("symlink")
			case syscall.S_IFIFO:
				fmt.Println("fifo")
			case syscall.S_IFCHR:
				fmt.Println("character device")
			case syscall.S_IFBLK:
				fmt.Println("block device")
			case syscall.S_IFREG:
				fmt.Println("regular file")
			case syscall.S_IFSOCK:
				fmt.Println("socket file")
			}
		},
	}

	uidCmd := cobra.Command{
		Use:  "set-uid <path> <uid>",
		Args: cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			path := args[0]
			uid, err := strconv.ParseUint(args[1], 10, 32)
			if err != nil {
				panic(err)
			}

			err = stat.SetUid(path, uint32(uid))
			if err != nil {
				panic(err)
			}
		},
	}
	gidCmd := cobra.Command{
		Use:  "set-gid <path> <gid>",
		Args: cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			path := args[0]
			gid, err := strconv.ParseUint(args[1], 10, 32)
			if err != nil {
				panic(err)
			}

			stat.SetGid(path, uint32(gid))
		},
	}

	rootCmd.AddCommand(&statCmd, &uidCmd, &gidCmd)

	rootCmd.Execute()
}
