package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"retroio/amstrad"
	"retroio/amstrad/dsk"
	"retroio/storage"
)

var amstradReadCmd = &cobra.Command{
	Use:                   "read FILE",
	Short:                 "Read the Amstrad disk directory contents",
	Long:                  `Read the directory contents from an Amstrad emulator DSK file.`,
	Args:                  cobra.ExactArgs(1),
	DisableFlagsInUseLine: true,
	Run: func(cmd *cobra.Command, args []string) {
		filename := args[0]

		f, err := os.Open(filename)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer f.Close()
		reader := storage.NewReader(f)

		var disk amstrad.Image
		dskType := mediaType(amstradMediaType, filename)

		switch dskType {
		case "dsk":
			disk = dsk.New(reader)
		default:
			fmt.Printf("Unsupported media type: '%s'", dskType)
			return
		}

		if err := disk.Read(); err != nil {
			fmt.Println("Storage read error!")
			fmt.Println(err)
			os.Exit(1)
		}

		disk.DirectoryListing()
	},
}

func init() {
	amstradReadCmd.Flags().StringVarP(&amstradMediaType, "media", "m", "", `Media type, default: file extension`)
	amstradCmd.AddCommand(amstradReadCmd)
}
