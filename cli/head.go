package cli

import (
	"fmt"
	"github.com/LapisBlue/Lapitar/head"
	"github.com/LapisBlue/Lapitar/mc"
	"github.com/LapisBlue/Lapitar/util"
	"github.com/ogier/pflag"
	"image"
	"image/png"
	"os"
)

const (
	headWidth         = 128
	headHeight        = 128
	headAngle         = 45
	headSuperSampling = 4
)

func runHead(name string, args []string) int {
	flags := pflag.NewFlagSet(name, pflag.ContinueOnError)

	width := flags.IntP("width", "w", headWidth, "The width of the canvas to render on, in pixels.")
	height := flags.IntP("height", "h", headHeight, "The height of the canvas to render on, in pixels.")
	angle := flags.Float32P("angle", "a", headAngle, "The angle to render the head at, in degrees.")
	superSampling := flags.IntP("supersampling", "s", headSuperSampling,
		"The amount of super sampling to perform, as a multiplier to width and height.")
	scale := &scaling{head.DefaultScale}
	flags.Var(scale, "scale", "The scaling method to use when rendering.")
	in := flags.StringP("in", "i", input, "The source of the list of players to render. Can be either a file, STDIN or ARGS.")
	out := flags.StringP("out", "o", output, "The destination to write the result to. Can be either a file or STDOUT.")

	nohelm := flags.Bool("no-helm", false, "Don't render the helm of the skin.")
	noshadow := flags.Bool("no-shadow", false, "Don't render the shadow of the head.")
	nolighting := flags.Bool("no-lighting", false, "Don't enable lighting.")

	FlagUsage(name, flags).
		Add("").
		Add("Example:") // TODO

	if len(args) < 1 || args[0] == "help" {
		flags.Usage()
		return 1
	}

	watch := util.GlobalWatch().Start()

	if flags.Parse(args) != nil {
		return 1
	}

	players := readFrom(*in, flags.Args())
	if players == nil {
		return 1
	}

	if isStdout(*out) {
		if len(players) > 1 {
			fmt.Fprintln(os.Stderr, "You can only render one image using STDOUT")
			return 1
		}

		player := players[0]
		skin, err := mc.Download(player)
		if err != nil {
			return PrintError(err, "Failed to download skin:", player)
		}

		head, err := head.Render(skin, *angle, *width, *height, *superSampling, !*nohelm, !*noshadow, !*nolighting, scale.Get())
		if err != nil {
			return PrintError(err, "Failed to render head:", player)
		}

		err = png.Encode(os.Stdout, head)
		if err != nil {
			return PrintError(err, "Failed to write head to STDOUT")
		}

		return 0
	}

	skins := downloadSkins(players)

	fmt.Println()
	fmt.Printf("Rendering %d head(s), please wait...\n", len(skins))

	watch.Mark()
	heads := make([]image.Image, len(skins))

	var err error
	for i, skin := range skins {
		if skin == nil {
			continue
		}

		watch.Mark()

		heads[i], err = head.Render(skin, *angle, *width, *height, *superSampling, !*nohelm, !*noshadow, !*nolighting, scale.Get())
		if err != nil {
			PrintError(err, "Failed to render head:", players[i], watch)
			continue
		}

		fmt.Println("Rendered head:", players[i], watch)
	}

	fmt.Println("Finished rendering heads", watch)

	fmt.Println()
	saveResults(players, heads, *out)

	fmt.Println()
	watch.Stop()
	fmt.Println("Done!", watch)

	return 0
}
