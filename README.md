# twchart

[![GitHub go.mod Go version (subdirectory of monorepo)](https://img.shields.io/github/go-mod/go-version/calvinmclean/twchart?filename=go.mod)](https://github.com/calvinmclean/twchart/blob/main/go.mod)
![GitHub Workflow Status](https://img.shields.io/github/actions/workflow/status/calvinmclean/twchart/main.yml?branch=main)
[![License](https://img.shields.io/github/license/calvinmclean/twchart)](https://github.com/calvinmclean/twchart/blob/main/LICENSE)
[![Go Reference](https://pkg.go.dev/badge/github.com/calvinmclean/twchart.svg)](https://pkg.go.dev/github.com/calvinmclean/twchart)

twchart is used to document and display sessions using a Thermoworks thermometer that is recorded in the cloud. Using the exported CSV of thermometer data, this program uses Apache ECharts to display other data on top of the graph.

The original inspiration and use-case for this was tracking bread-making since timing and temperature are important aspects of the process. On top of the timeseries temperature data, this is used to track different stages and adhoc notes throughout the process. In order to use it, you should track your notes using this flexible format, then `POST` to the server after the process is complete.


## Notes format

The following format can be used to easily take time-based notes record stages. The server is able to parse this and display on top of a timeseries graph.

```
[Session Name]
Date: 2006-01-02

[Probe Name] Probe: 1
[Probe Name] Probe: [...]
[Probe Name] Probe: [n]

Note: 3:04PM: [Notes...]
[Stage Name]: 3:04PM
Note: 3:04PM: [Notes...]
[Stage Name]: 3:04PM

Done: 3:04PM
```

- Dates and times here use [Go's formatting conventions](https://pkg.go.dev/time#pkg-constants)
- `3:04PM` timestamps can be replaced with elapsed durations (`3m`, `1h30m`, etc.)
- `[]`: brackets above are placeholders for any text. Do not include the brackets. Do not use colons in text
- Everything must be in chronological order
- Notes and stages can happen at any time
- You can have any number of notes and stages

### Bread Example

Break-making is easy to track using timestamps since steps are pretty spread out.

<details>
<summary>Click to expand</summary>

```
Ciabatta
Date: 2025-05-24

Ambient Probe: 1
Oven Probe: 2

Note: 8:10PM: preparing to make biga

Preferment: 8:10PM
Note: 8:13PM: finished mixing biga

Bulk ferment: 7:00AM
Note: 8:00AM: 10 stretch and folds

Final Proof: 9:00AM
Note: 9:00AM: shaped dough

Bake: 10:30AM
Done: 10:55AM

Note: 12:00PM: bread is delicious and crunchy
```

</details>

### Coffee Roasting Example

Coffee roasting happens pretty quickly, so it's easier to record durations instead of timestamps.

<details>
<summary>Click to expand</summary>

```
Coffee
Date: 2025-05-24

Ambient Probe: 1
Bean Probe: 2

Note: 8:00PM: preheat

Drying: 1m
Note: 1m: fan 9, heat 5

Maillard: 4m
Note: 4m: fan 7, heat 7

Development: 7m
Note: 7m: fan 5, heat 6
Note: 7m30s: first crack

Cooling: 8m30s

Done: 10m30s
```

</details>

## Installation

```shell
go install github.com/calvinmclean/twchart/cmd/twchart@latest
```

```shell
docker pull ghcr.io/calvinmclean/twchart:latest
```

## Usage

Run the server with existing data:
```shell
twchart serve --dir data/
```
This assumes you have a nested directory structure with `.txt` and `.csv` files with matching filenames (`bread.txt`/`bread.csv`)

See the `docker-compose.yml` file for an example using docker.

### Uploading Data

Here is my process, but something else might work better for you.

1. During baking/roasting/bbqing, record notes in a Notes app
2. When it's done, use a Siri shortcut to make this request:
  ```shell
  curl \
    -X POST \
    -H "Content-Type: text/plain" \
    --data-binary "@example.txt" \
    localhost:8080/sessions
  ```
3. Download the exported CSV from Thermoworks Cloud and upload to the server with a Siri shortcut making this request:
  ```shell
  curl \
    -X POST \
    -H "Content-Type: text/csv" \
    --data-binary "@example.csv" \
    localhost:8080/sessions/upload-csv
  ```
  - The `/upload-csv` endpoint will load the CSV data into the most recently-created Session
