// Limited reimplementation of hyprctl using hyprland-go to show an example
// on how it can be used.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/thiagokokada/hyprland-go"
)

var (
	c *hyprland.RequestClient
	// default error/usage output
	out io.Writer = flag.CommandLine.Output()
)

// Needed for an acumulator flag, i.e.: can be passed multiple times, get the
// results in a string array
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

// Unmarshal structs as JSON and indent output
func mustMarshalIndent(v any) []byte {
	return must1(json.MarshalIndent(v, "", "   "))
}

func usage(m map[string]func(args []string)) {
	must1(fmt.Fprintf(out, "Usage of %s:\n", os.Args[0]))
	must1(fmt.Fprintf(out, "  %s [subcommand] <options>\n\n", os.Args[0]))
	must1(fmt.Fprintf(out, "Available subcommands:\n"))

	// Sort keys before printing, since Go randomises order
	subcommands := make([]string, len(m))
	i := 0
	for s := range m {
		subcommands[i] = s
		i++
	}
	sort.Strings(subcommands)
	for _, s := range subcommands {
		must1(fmt.Fprintf(out, "  - %s\n", s))
	}
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

	// Map pf subcommands to a function to handle the subcommand. Will
	// receive the subcommand arguments as parameter
	m := map[string]func(args []string){
		"activewindow": func(_ []string) {
			v := must1(c.ActiveWindow())
			must1(fmt.Printf("%s\n", mustMarshalIndent(v)))
		},
		"activeworkspace": func(_ []string) {
			v := must1(c.ActiveWorkspace())
			must1(fmt.Printf("%s\n", mustMarshalIndent(v)))
		},
		"batch": func(args []string) {
			must(batchFS.Parse(args))
			if len(batch) == 0 {
				must1(fmt.Fprintf(out, "Error: at least one '-c' is required for batch.\n"))
				os.Exit(1)
			} else {
				// Batch commands are done in the following way:
				// `[[BATCH]]command0 param0 param1; command1 param0 param1;`
				r := hyprland.RawRequest(
					fmt.Sprintf("[[BATCH]]%s", strings.Join(batch, ";")),
				)
				v := must1(c.RawRequest(r))
				must1(fmt.Printf("%s\n", v))
			}
		},
		"dispatch": func(args []string) {
			must(dispatchFS.Parse(args))
			if len(dispatch) == 0 {
				must1(fmt.Fprintf(out, "Error: at least one '-c' is required for dispatch.\n"))
				os.Exit(1)
			} else {
				v := must1(c.Dispatch(dispatch...))
				must1(fmt.Printf("%s\n", v))
			}
		},
		"kill": func(_ []string) {
			v := must1(c.Kill())
			must1(fmt.Printf("%s\n", v))
		},
		"reload": func(_ []string) {
			v := must1(c.Reload())
			must1(fmt.Printf("%s\n", v))
		},
		"setcursor": func(_ []string) {
			must(setcursorFS.Parse(os.Args[2:]))
			v := must1(c.SetCursor(*theme, *size))
			must1(fmt.Printf("%s\n", v))
		},
		"version": func(_ []string) {
			v := must1(c.Version())
			must1(fmt.Printf("%s\n", mustMarshalIndent(v)))
		},
	}

	flag.Usage = func() { usage(m) }
	flag.Parse()

	if len(os.Args) < 2 {
		flag.Usage()
		os.Exit(1)
	}

	subcommand := os.Args[1]
	if run, ok := m[subcommand]; ok {
		c = hyprland.MustClient()
		run(os.Args[2:])
	} else {
		must1(fmt.Fprintf(out, "Error: unknown subcommand: %s\n", subcommand))
		os.Exit(1)
	}
}
