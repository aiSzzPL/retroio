package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"retroio/commodore"
	"retroio/commodore/t64"
	"retroio/commodore/tap"
	"retroio/storage"
)

var commodoreMediaType string

var commodoreGeometryCmd = &cobra.Command{
	Use:   "geometry FILE",
	Short: "Read the Commodore tape file geometry",
	Long: `Read the geometry - headers and data blocks - from a Commodore emulator TAP
or T64 tape file.`,
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

		var dsk commodore.Image
		dskType := mediaType(commodoreMediaType, filename)

		switch dskType {
		case "t64":
			dsk = t64.New(reader)
		case "tap":
			dsk = tap.New(reader)
		default:
			fmt.Printf("Unsupported media type: '%s'", dskType)
			return
		}

		if err := dsk.Read(); err != nil {
			fmt.Println("Media read error!")
			fmt.Println(err)
			os.Exit(1)
		}

		dsk.DisplayGeometry()
	},
}

func init() {
	commodoreGeometryCmd.Flags().StringVarP(&commodoreMediaType, "media", "m", "", `Media type, default: file extension`)
	commodoreCmd.AddCommand(commodoreGeometryCmd)
}
