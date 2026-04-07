# bc125at

A command-line tool for programming the Uniden BC125AT radio scanner over USB serial.

Supports reading and writing channels and settings using JSON or CSV files.

This is a Go port of [bc125py](https://github.com/itsmaxymoo/bc125py) by itsmaxymoo.

## Requirements

- Uniden BC125AT scanner powered on and connected via USB
- Go 1.21+ (to build from source)

## Installation

### Pre-built binaries

Download from the `build/dist/` directory for your platform:

| Platform | Binary |
|----------|--------|
| Linux (amd64) | `bc125at-0.1.0-linux-amd64` |
| Linux (arm64) | `bc125at-0.1.0-linux-arm64` |
| macOS (Intel) | `bc125at-0.1.0-darwin-amd64` |
| macOS (Apple Silicon) | `bc125at-0.1.0-darwin-arm64` |
| Windows (amd64) | `bc125at-0.1.0-windows-amd64.exe` |

### Build from source

```sh
make build       # produces build/bc125at
make release     # cross-compiles all platforms into build/dist/
```

## Usage

```
bc125at [--port <port>] [--verbose] <command> [args]
```

The serial port is auto-detected. If multiple ports are found you will be prompted to choose. Use `--port` to specify one directly (e.g. `/dev/ttyACM0`, `COM3`).

`--verbose` prints the raw AT commands sent to and received from the scanner.

### Commands

#### `import <file>`

Read data from the scanner and save to a file.

```sh
bc125at import backup.json          # full backup (settings + all channels)
bc125at import channels.csv --csv   # channels only, CSV format
bc125at import - --csv              # write CSV to stdout
```

#### `export <file>`

Write data from a file to the scanner. The file is validated before anything is written.

```sh
bc125at export backup.json          # restore full backup (settings + channels)
bc125at export channels.csv --csv   # write channels from CSV
bc125at export - --csv              # read CSV from stdin
```

#### `validate <file>`

Check a backup file for errors without connecting to the scanner.

```sh
bc125at validate backup.json
bc125at validate channels.csv --csv
```

#### `wipe`

Delete all 500 channels from the scanner. Requires typed confirmation.

```sh
bc125at wipe
```

Make a backup first: `bc125at import backup.json`

#### `unlock`

Send the EPG (Exit Program) command to rescue a scanner stuck in Program Mode.

```sh
bc125at unlock
```

#### `shell [file]`

Open an interactive AT command shell, or run a script file non-interactively.

```sh
bc125at shell               # interactive
bc125at shell script.txt    # run script file
bc125at shell -             # read commands from stdin
```

Command history is saved to `~/.bc125at_history` (up to 500 entries). Use `--clear-history` to reset it.

Built-in shell commands: `help`, `history`, `quit`/`exit`

Common AT commands (send `PRG` first for most channel/settings operations):

| Command | Description |
|---------|-------------|
| `MDL` | Get model name |
| `VER` | Get firmware version |
| `PRG` | Enter program mode |
| `EPG` | Exit program mode |
| `CIN,<n>` | Read channel n (1–500) |
| `DCH,<n>` | Delete channel n |
| `VOL[,<n>]` | Get/set volume (0–15) |
| `SQL[,<n>]` | Get/set squelch (0–15) |
| `BLT[,<m>]` | Get/set backlight (`AO`/`10`/`30`/`SQ`) |

Lines beginning with `#` are treated as comments in script files.

#### `tones`

List all valid CTCSS/DCS tone identifiers for use in CSV/JSON files.

```sh
bc125at tones
```

## File Formats

### JSON

Full backup including scanner settings and all channels:

```json
{
  "model": "BC125AT",
  "firmware": "1.00.06",
  "settings": {
    "backlight": "AO",
    "battery_save": 0,
    "key_beep": 3,
    "key_lock": false,
    "priority": 0,
    "volume": 10,
    "squelch": 3,
    "contrast": 7,
    "scan_banks": "1111111111",
    "weather_scan": 0
  },
  "channels": [...]
}
```

### CSV

Channels only. Required columns: `index`, `frequency_mhz`, `modulation`, `ctcss_dcs`, `delay`.

```csv
index,name,frequency_mhz,modulation,ctcss_dcs,delay,locked_out,priority
1,Fire Disp,154.2800,FM,none,2,0,0
2,Police,460.1000,NFM,ctcss_100.0,2,0,1
```

## Channel Fields

| Field | Description | Valid values |
|-------|-------------|--------------|
| `index` | Channel number | 1–500 |
| `name` | Channel name | Up to 16 characters |
| `frequency_mhz` | Frequency in MHz | 25–54, 108–174, 225–380, 400–512 |
| `modulation` | Modulation type | `AUTO`, `FM`, `NFM`, `AM` |
| `ctcss_dcs` | Tone squelch | `none`, `search`, `no_tone`, `ctcss_<freq>`, `dcs_<code>` |
| `delay` | Squelch delay (seconds) | `-10`, `-5`, `0`, `1`–`5` |
| `locked_out` | Exclude from scan | `0`/`1` or `true`/`false` |
| `priority` | Priority channel | `0`/`1` or `true`/`false` |

## Settings Fields

| Field | Description | Valid values |
|-------|-------------|--------------|
| `backlight` | Backlight mode | `AO` (always on), `10`, `30`, `SQ` (squelch) |
| `battery_save` | Battery charge timer (hours) | `0` (off) – `16` |
| `key_beep` | Key beep level | `0`–`99` |
| `key_lock` | Key lock enabled | `true`/`false` |
| `priority` | Priority scan mode | `0` (off), `1` (on), `2` (plus) |
| `volume` | Volume level | `0`–`15` |
| `squelch` | Squelch level | `0`–`15` |
| `contrast` | Display contrast | `1`–`15` |
| `scan_banks` | Active scan banks | 10-character string of `0`/`1` |
| `weather_scan` | Weather alert | `0` (off), `1` (on) |

## Troubleshooting

### `cdc_acm probe failed with error -22` / device not detected

If you see the following in `dmesg`:

```
cdc_acm 2-2:1.0: Zero length descriptor references
cdc_acm 2-2:1.0: probe with driver cdc_acm failed with error -22
```

The `cdc_acm` driver is failing to claim the device. Load the `usbserial` driver manually with the BC125AT's vendor and product IDs:

```sh
sudo modprobe usbserial vendor=0x1965 product=0x0017
```

The scanner should then appear as `/dev/ttyUSB0` (or similar).

## License

BSD 2-Clause. See [LICENSE](LICENSE).
