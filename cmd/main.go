package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"zone-finder/workoutfile"
	"zone-finder/zones"
)

func formatOutput(zones zones.HeartRateZones) string {
	const output = `
LTHR: %v bpm
Zone 1: 0-%v
Zone 2: %v-%v
Zone 3: %v-%v
Zone 4: %v-%v
Zone 5: %v+
`
	return strings.TrimPrefix(fmt.Sprintf(
		output,
		zones.LTHR,
		zones.Zones[0].Max,
		zones.Zones[1].Min,
		zones.Zones[1].Max,
		zones.Zones[2].Min,
		zones.Zones[2].Max,
		zones.Zones[3].Min,
		zones.Zones[3].Max,
		zones.Zones[4].Min,
	), "\n")
}

func main() {
	exitCode := run(os.Args, os.Stdout, os.Stderr)
	os.Exit(exitCode)
}

func validateArgs(args []string) error {
	if len(args) == 1 {
		return errors.New("missing required argument: file path")
	} else if len(args) > 2 {
		return errors.New("we're only able to handle a single file at a time")
	}

	return nil
}

func showUsage(w io.Writer) {
	usage := `
Usage: zone-finder <file.ext>

Calculate heart rate training zones from FIT or TCX workout files using the
Lactate Threshold Heart Rate (LTHR) method.

Arguments:
  <file.ext>    Path to a workout file

Options:
  -h, --help    Show this help message

Examples:
  zone-finder workout.tcx
  zone-finder ~/Documents/garmin-run.fit

The program analyzes the last 20 minutes of your workout to determine
your LTHR, then calculates 5 training zones based on percentages of LTHR.
`

	fmt.Fprint(w, usage)
}

func checkHelpFlag(args []string) bool {
	if len(args) < 2 {
		return false
	}

	arg := args[1]
	if arg == "-h" || arg == "--help" {
		return true
	}

	return false
}

func run(args []string, stdout io.Writer, stderr io.Writer) int {
	if isHelp := checkHelpFlag(args); isHelp {
		showUsage(stdout)
		return 0
	}

	if err := validateArgs(args); err != nil {
		showUsage(stderr)
		return 1
	}

	workoutFile := args[1]
	workout, err := workoutfile.ParseFile(workoutFile)
	if err != nil {
		fmt.Fprintf(stderr, "failed to parse workout file: %v\n", err)
		return 1
	}

	hrData, err := workout.GetHRDataPoints()
	if err != nil {
		fmt.Fprintf(stderr, "failed to process heart rate data: %v\n", err)
		return 1
	}

	zones, err := zones.CalculateZonesFromHRData(hrData)
	if err != nil {
		fmt.Fprintf(stderr, "failed to calculate zones: %v\n", err)
		return 1
	}

	fmt.Fprint(stdout, formatOutput(zones))
	return 0
}
