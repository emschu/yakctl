# yakctl

**A simple command line tool to control your [Yakuake](https://kde.org/applications/en/yakuake) terminal 
instance by describing sessions (tabs and properties) to ease workflows.**

Yakuake is a "drop-down terminal emulator based on Konsole technologies" and is part of the KDE software suite.

This tool makes use of the dbus-interface Yakuake offers and is written in [Golang](https://golang.org/).

## Command
```bash
NAME:
   yakctl - Control the yakuake terminal and your terminal sessions

USAGE:
   yakctl [global options] command [command options] [arguments...]

VERSION:
   1.0.0

COMMANDS:
   clear, c    Clear all sessions and terminals
   profile, p  Manage defined profiles, default: list available profiles
   exec, e     Execute a command in all or specific terminals
   status, s   List status (=sessions, terminals) of the current yakuake instance
   help, h     Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --config value  configuration file (default: "/home/worker/.yakctl.yml")
   --verbose       verbose log output (default: false)
   --help, -h      show help (default: false)
   --version, -v   print the version (default: false)

COPYRIGHT:
   yakctl  2020  https://github.com/emschu/yakctl

   This program comes with ABSOLUTELY NO WARRANTY.
   This is free software, and you are welcome
   to redistribute it under the conditions of the
   GPLv3 (https://www.gnu.org/licenses/gpl-3.0.txt).

```

## Setup
1. Install this project by installing it via `go install github.com/emschu/yakctl@latest`
2. Create a `~/.yakctl.yml` configuration file and read the configuration section below.
3. Start using `yakctl`

### Requirements
- Yakuake of the KDE project needs to be installed
- `qdbus` command, check with `which qdbus`

## Configuration
You can find this example in `.yakctl.yml` of this repository.
It contains all supported configuration options and is expected to be at `~/.yakctl.yml`. 
Alternatively you can provide a path to a configuration file using the command line option `--config <path>`.   

You should note:
- the `clear` flag means that all your yakuake tabs will be closed, except the protected ones. To also remove the latter, use `force: true`
- commands listed in `commands` are executed before `terminalX`

```yml
---
profiles:
  - name: default
    clear: true
    force: true
    tabs:
      - name: raspi1_ssh
        monitorActivity: true
        monitorSilence: true
        disableInput: true
        protected: true
        commands:
          - ssh pi@10.10.10.11
          - echo 'hello world says the pi'
      - name: raspi2_ssh
        commands:
          - ssh pi@10.10.10.12
        protected: true
      - name: left-right split terminal
        split: lr
        commands:
          - cd /var/www/html
        terminal1:
          - top
        terminal2:
          - htop
      - name: top-bottom split terminal
        split: tb
        disableInput: true
        commands:
          - "echo 'hello'"
          - echo 'hello2'
          - echo 'hello3'
      - name: quad_tab
        split: quad
        monitorActivity: true
        monitorSilence: true
        disableInput: true
        commands:
          - "echo 'all'"
        terminal1:
          - echo "terminal1"
        terminal3:
          - echo "terminal3"
        terminal4:
          - echo "terminal4"
      - name: go-shell
        commands:
          - cd ~/go/src
  - name: other_workspace
    tabs:
      - name: raspi3_ssh
        commands:
          - "ssh pi@10.10.10.154"
        protected: true
```

## Examples

```bash 
## List your defined profiles
$ yakctl profile
OR
$ yakctl p

## Open first profile
$ yakctl profile open 1
OR
$ yakctl p o 1

## Execute "echo 'hello world'" in ALL open terminals of yakuake
$ yakctl exec echo 'hello world' 
```

## License
**GPL v3** - for details see the [full license text](./LICENSE).

## More information
- [Yakuake KDE GitLab repository](https://invent.kde.org/utilities/yakuake)