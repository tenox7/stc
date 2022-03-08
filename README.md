# STC - Syncthing Cli

Stc is a command line tool for [Syncthing](https://syncthing.net/).
It can be used to quickly check status of Syncthing from a terminal / command line
without need of a Web Browser. For example on a remote machine over ssh, without port
forwarding or if you have large number of machines to query.

```
$ stc
Host      Uptime    Version
homenas   2 weeks   v1.19.0

Folder    Paused   State   Global   Local
pics      false    idle    37 GB    37 GB
docs      false    idle    4 GB     4 GB
backups   false    idle    86 GB    86 GB

Device          Paused    Conn   Sync%   Download  Upload
office          false     true   100.0%  11 kB     11 kB
laptop          false     false  83.2%   0 B       0 B
jakob-home      false     true   100.0%  89 MB     447 kB
backup-nas      false     true   100.0%  6.3 kB    7.0 kB
*homenas        false     true   100.0%  0 B       0 B
```

## Functionality

Stc doesn't attempt to be a fully fledged CLI with all functions from the Web UI.
Instead it aims to have rudimentary functionality for operation on remote devices.

Currently supported is a basic dashboard with status and a number of commands to
perform basic tasks like restart, shutdown, rescan, override and revert, print and
clear errors, etc.

## Usage

```sh
$ stc [flags] [commands]
```

### Easy Mode

Place `stc` binary in Syncthing home directory and run it from there, or
specify it with `--homedir=/path..`. Stc will try to find the URL and API
Key on it's own.

### Advanced / Remote Mode

Stc takes `--apikey=xxx` and `--target=http://...` flags to connect to a
Syncthing service. The API Key can be obtained from the Syncthing Web UI
(Settings:General tab) or from `config.xml` file.

API Key can also be specified by `APIKEY` environmental variable.

If you place `stc` binary in the Syncthing home folder or specify `--homedir`
flag, it will try to obtain the right values from `config.xml`.

If you use TLS/SSL/https without valid certificate you can use the flag
`--ignore_cert_errors` to suppress the errors. This is considered very insecure.

## Flags

```text
  --apikey              - Syncthing API Key
  --target              - URL of the Syncthing target
  --homedir             - Path of Syncthing home directory, if specified stc
                          will try to find apikey and target from config.xml
  --ignore_cert_errors  - Ignore cert errors while using https/SSL/TLS
```

## Arguments / Commands

```text
  log           - print syncthing 'recent' log
  restart       - restart syncthing daemon
  shutdown      - shutdown syncthing daemon
  errors        - print errors visible in web UI
  clear_errors  - clear errors in the web UI
  post_error    - posts a custom error message in the web UI
  folder_errors - prints folder errors from scan or pull
  id            - print ID of this node
  reset_db      - reset the database / file index
  rescan        - rescan a folder or 'all'
  override      - override remote changed for a send-only folder (OoSync)
  revert        - revert local changes for a receive-only folder (LocAdds)
```

## Download binaries

See [Releases](https://github.com/tenox7/stc/releases)

## Legal

* Copyright 2022 Google LLC
* Licensed under Apache 2.0
* This is not an official Google product
