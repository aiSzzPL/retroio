package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"retroio/spectrum"
	"retroio/spectrum/tap"
	"retroio/spectrum/tzx"
	"retroio/storage"
)

var speccyReadCmd = &cobra.Command{
	Use:                   "read FILE",
	Short:                 "Read a ZX Spectrum tape file",
	Long:                  `Read the contents of a ZX Spectrum emulator TAP or TZX tape file.`,
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

		var dsk spectrum.Image
		dskType := mediaType(spectrumMediaType, filename)

		switch dskType {
		case "tap":
			dsk = tap.New(reader)
		case "tzx":
			dsk = tzx.New(reader)
		default:
			fmt.Printf("Unsupported media type: '%s'", dskType)
			return
		}

		if err := dsk.Read(); err != nil {
			fmt.Println("Media read error!")
			fmt.Println(err)
			os.Exit(1)
		}

		if spectrumBasListing {
			dsk.DisplayBASIC()
		} else {
			cmd.Help()
			fmt.Println("\nPlease select '--bas' for BASIC program listing.")
		}
	},
}

func init() {
	speccyReadCmd.Flags().StringVarP(&spectrumMediaType, "media", "m", "", `Media type, default: file extension`)
	speccyReadCmd.Flags().BoolVar(&spectrumBasListing, "bas", false, `BASIC program listing`)
	spectrumCmd.AddCommand(speccyReadCmd)
}
