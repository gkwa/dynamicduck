# dynamicduck

A command-line tool for randomly sampling JSON datasets.

## Purpose

`dynamicduck` is autilityto process JSON datasets, randomly select items, and output the results while preserving the original data structure.

## Features

- Read JSON data from stdin or input file
- Randomly select N items from the dataset
- Preserve original JSON metadata
- Flexible output options (stdout or file)
- Verbose logging levels
- Seen items tracking to avoid duplicates
- Reproducible sampling with seed option

## Usage Example

```bash
# Sample 10 items from a large dataset, tracking seen items
rm -f /tmp/seen /tmp/results
dynamicduck --count 10 --seen-file /tmp/seen -vv --input large_dataset.json | noblenewtonia parse-json --input - >/tmp/results

# Sample items from stdin
cat data.json | dynamicduck > output.json
```

## Command-Line Options

| Option          | Description                                | Default    |
| --------------- | ------------------------------------------ | ---------- |
| `--count`       | Number of items to select                  | 10         |
| `--input`       | Input file path (stdin if not specified)   | stdin      |
| `--out-file`    | Output file path (stdout if not specified) | stdout     |
| `--seed`        | Seed for random number generation          | time-based |
| `--seen-file`   | File to track seen items                   | not set    |
| `-v, --verbose` | Increase verbosity level                   | 0          |

### Advanced Options

#### Verbosity Levels

- No flags: Minimal output
- `-v`: Basic operations info
- `-vv`: Detailed operations
- `-vvv`: Debug information

#### Sampling Options

```bash
# Select specific number of items
dynamicduck --count 20 --input data.json

# Use a fixed seed for reproducible sampling
dynamicduck --seed 12345 --input data.json

# Avoid sampling previously seen items
dynamicduck --seen-file /tmp/seen --input data.json
```
