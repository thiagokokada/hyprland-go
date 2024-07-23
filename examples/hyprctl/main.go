// Limited reimplementation of hyprctl using hyprland-go to show an example
// on how it can be used.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/thiagokokada/hyprland-go"
)

var (
	c   *hyprland.RequestClient
	out io.Writer = flag.CommandLine.Output()
)

// https://stackoverflow.com/a/28323276
type arrayFlags []string

func (i *arrayFlags) String() string {
	return fmt.Sprintf("%v", *i)
}

func (i *arrayFlags) Set(v string) error {
	*i = append(*i, v)
	return nil
}

func must1[T any](v T, err error) T {
	must(err)
	return v
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func mustMarshalIndent(v any) []byte {
	return must1(json.MarshalIndent(v, "", "   "))
}

func main() {
	batchFS := flag.NewFlagSet("batch", flag.ExitOnError)
	var batch arrayFlags
	batchFS.Var(&batch, "c", "Command to batch, can be passed multiple times. "+
		"Please quote commands with arguments (e.g.: 'dispatch exec kitty')")

	dispatchFS := flag.NewFlagSet("dispatch", flag.ExitOnError)
	var dispatch arrayFlags
	dispatchFS.Var(&dispatch, "c", "Command to dispatch, can be passed multiple times. "+
		"Please quote commands with arguments (e.g.: 'exec kitty')")

	setcursorFS := flag.NewFlagSet("setcursor", flag.ExitOnError)
	theme := setcursorFS.String("theme", "Adwaita", "Cursor theme")
	size := setcursorFS.Int("size", 32, "Cursor size")

	m := map[string]func(){
		"activewindow": func() {
			v := must1(c.ActiveWindow())
			fmt.Printf("%s\n", mustMarshalIndent(v))
		},
		"activeworkspace": func() {
			v := must1(c.ActiveWorkspace())
			fmt.Printf("%s\n", mustMarshalIndent(v))
		},
		"batch": func() {
			batchFS.Parse(os.Args[2:])
			if len(batch) == 0 {
				fmt.Fprintf(out, "Error: at least one '-c' is required for batch.\n")
				os.Exit(1)
			} else {
				// Batch commands are done in the following way:
				// `[[BATCH]]command0 param0 param1; command1 param0 param1;`
				r := hyprland.RawRequest(
					fmt.Sprintf("[[BATCH]]%s", strings.Join(batch, ";")),
				)
				v := must1(c.RawRequest(r))
				fmt.Printf("%s\n", v)
			}
		},
		"dispatch": func() {
			dispatchFS.Parse(os.Args[2:])
			if len(dispatch) == 0 {
				fmt.Fprintf(out, "Error: at least one '-c' is required for dispatch.\n")
				os.Exit(1)
			} else {
				v := must1(c.Dispatch(dispatch...))
				fmt.Printf("%s\n", v)
			}
		},
		"kill": func() {
			v := must1(c.Kill())
			fmt.Printf("%s\n", v)
		},
		"reload": func() {
			v := must1(c.Reload())
			fmt.Printf("%s\n", v)
		},
		"setcursor": func() {
			setcursorFS.Parse(os.Args[2:])
			v := must1(c.SetCursor(*theme, *size))
			fmt.Printf("%s\n", v)
		},
		"version": func() {
			v := must1(c.Version())
			fmt.Printf("%s\n", mustMarshalIndent(v))
		},
	}

	flag.Usage = func() {
		fmt.Fprintf(out, "Usage of %s:\n", os.Args[0])
		fmt.Fprintf(out, "  %s [subcommand] <options>\n\n", os.Args[0])
		fmt.Fprintf(out, "Available subcommands:\n")
		for k := range m {
			fmt.Fprintf(out, "  - %s\n", k)
		}
	}
	flag.Parse()

	if len(os.Args) < 2 {
		flag.Usage()
		os.Exit(1)
	}

	subcmd := os.Args[1]
	if run, ok := m[subcmd]; ok {
		c = hyprland.MustClient()
		run()
	} else {
		fmt.Fprintf(out, "Error: unknown subcommand: %s\n", subcmd)
		os.Exit(1)
	}
}
