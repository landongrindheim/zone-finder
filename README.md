# zone-finder

Calculate heart rate training zones from workout files using the Lactate Threshold Heart Rate (LTHR) method.

## Installation
```bash
go install github.com/landongrindheim/zone-finder@latest
```

Or build from source:
```bash
git clone https://github.com/landongrindheim/zone-finder.git
cd zone-finder
go build -o zone-finder ./cmd
```

## Usage
```bash
zone-finder
```

**Supported formats:** TCX, FIT

**Example:**
```bash
$ zone-finder ~/workouts/morning-run.tcx
LTHR: 172 bpm
Zone 1: 0-137
Zone 2: 138-151
Zone 3: 152-161
Zone 4: 162-172
Zone 5: 173+
```

## How It Works

zone-finder analyzes the last 20 minutes of your workout to determine your Lactate Threshold Heart Rate (LTHR), then calculates 5 training zones based on percentages of LTHR:

- **Zone 1** (Recovery): < 80% of LTHR
- **Zone 2** (Endurance): 80-88% of LTHR
- **Zone 3** (Tempo): 89-94% of LTHR
- **Zone 4** (Threshold): 95-100% of LTHR
- **Zone 5** (VO2 Max): > LTHR

Based on the method described by [David Roche](https://www.trailrunnermag.com/training/trail-tips-training/how-to-find-your-lactate-threshold/).

## Requirements

- Workout file must be at least 20 minutes long (ideally 30 or more)
- Heart rate data required
  - For best results, use a chest-strap or arm band heart rate monitor
- TCX or FIT format

## Development
```bash
# Run tests
go test ./...

# Run with coverage
go test -cover ./...

# Build
go build -o zone-finder ./cmd
```

## License

MIT
