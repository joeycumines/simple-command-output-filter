# `simple-command-output-filter`

Executes a `command [args...]`, filters its `stdout` based on user-defined patterns, and transparently handles `stdin`,
`stderr`, signals, and the command's exit status.

## Table of Contents

- [Synopsis](#synopsis)
- [Options](#options)
- [Pattern Matching](#pattern-matching)
    - [Syntax](#syntax)
    - [Pattern Files](#pattern-files--f---pattern-file)
    - [Behavior Without Patterns](#behavior-without-patterns)
- [Execution & Transparency](#execution--transparency)
    - [Exit Status](#exit-status)
- [Examples](#examples)

## Synopsis

```sh
simple-command-output-filter [options] [--] command [args...]
```

* `--`: Optional; separates filter options from the `command`. Essential if `command` or `args` begin with `-`.

## Options

* `-p PATTERN`, `--pattern PATTERN`: Defines a pattern. Use multiple times for multiple patterns.
* `-f FILE`, `--pattern-file FILE`: Reads patterns from `FILE` (one per line). Use multiple times.
* `-v`, `--invert-match`: Inverts the match; prints lines that *do not* match any pattern.
* `-e MODE`, `--error-mode MODE`: Alters exit status based on filtered output *if the command succeeds*. `MODE` can be:
    * `default`: (Default) Exit status primarily mirrors the command's.
    * `no-content`: Exits `1` if the filter produces *no output* (and command succeeded), else `0`.
    * `on-content`: Exits `1` if the filter produces *any output* (and command succeeded), else `0`.
* `-h`, `--help`: Displays the help message and exits.

## Pattern Matching

Filters `stdout` lines from the executed command. A line is printed if it matches *any* specified pattern (or *no*
patterns if `-v` is active).

### Syntax

Patterns are case-sensitive and matched against the entire line (implicitly anchored with `^` and `$`):

* `*`: Wildcard, matches zero or more characters (becomes `.*` in regex).
* `**`: Matches a literal asterisk character (`*`).
* Other characters are treated literally. Regex metacharacters (e.g., `.`, `+`, `?`, `(`, `)`) are automatically
  escaped.
    * Example: `foo*bar` (regex `^foo.*bar$`) matches "foodbar", "foobar".
    * Example: `config.value[0]` (regex `^config\.value\[0\]$`) matches the literal string "config.value[0]".

### Pattern Files (`-f`, `--pattern-file`)

* One pattern per line. Lines are trimmed of leading/trailing whitespace.
* `#` initiates a comment (ignored to end-of-line), unless `##` which is treated as a literal `#` in the pattern.
* Lines that are empty or contain only comments (after processing `##`) are ignored.

### Behavior Without Patterns

* **Default (no `-v`)**: If no patterns are provided, no lines from `stdout` are printed.
* **Inverted (`-v`)**: If no patterns are provided, all lines from `stdout` are printed.

## Execution & Transparency

`simple-command-output-filter` acts as a thin wrapper:

* **`stdin`**: Passed directly to the executed command.
* **`stderr`**: Passed through unmodified from the command.
* **Signals**: Forwards signals like `SIGINT` (Ctrl+C) and `SIGTERM` to the command's process group, allowing graceful
  termination.

### Exit Status

The filter aims to mirror the command's exit status, to optionally `exit 1` on
presence or lack of content, or indicate its own errors:

* **`0`**:
    * Normal successful execution: The command completed successfully (exited `0`), AND the active error mode (if any)
      condition for exit `0` was met.
    * Help displayed (`-h`, `--help`).
* **`1`**:
    * Error mode triggered: The command completed successfully (exited `0`), but the active error mode (`no-content` or
      `on-content`) condition for exit `1` was met.
    * Command failure/runtime error: The command was terminated by a signal, exited with a generic non-positive status (
      e.g., Go's internal `-1` for signals), or a runtime error occurred within the filter during command execution or
      setup.
* **`N` (Specific positive integer from the command, e.g., `5`, `127`):**
    * Command failure: The executed command explicitly exited with this positive status `N`. This exit status is
      propagated directly from the command.
* **`2`**:
    * Filter initialization error: Invalid command-line flags, no command specified, or errors loading pattern files.

## Examples

1. **Show only directories from `ls -l` (lines typically starting with 'd'):**
   ```bash
   simple-command-output-filter -p "d*" -- ls -l
   ```
2. **Exclude lines containing "DEBUG" from `my_script.sh` output:**
   ```bash
   simple-command-output-filter -v -p "*DEBUG*" -- ./my_script.sh
   ```
3. **Monitor `app.log`, showing lines with "ERROR" or "CRITICAL", using patterns from `filters.txt`:**
   `filters.txt`:
   ```
   *ERROR*
   *CRITICAL* # Show critical alerts
   TASK_##COMPLETED    # Match literal "TASK_#COMPLETED"
   ```
   Command:
   ```bash
   simple-command-output-filter -f filters.txt -- tail -f /var/log/app.log
   ```
4. **Pass arguments starting with `-` to the target command, filtering for specific output:**
   ```bash
   simple-command-output-filter -p "result:*" -- my_tool --process --input-file /dev/null -v
   ```
