# STC - Syncthing Cli

This is an unofficial command line tool for [Syncthing](https://syncthing.net/).
Stc can be used to quicky check status of Syncthing from a terminal / command line
without need of a Web Browser. For example on a remote machine over ssh, without port
forwarding or if you have large number of machines to query.

## Usage

The tool by default connects to `http://127.0.0.1:8384` however you can target
different machines by using `--target` flag.

You must specify API Key by using etiher `--apikey` flag or `APIKEY` environment
variable. The API Key can be obtained from the Syncthing Web UI or `config.xml`
(`grep apikey config.xml`).

## Flags

```text
  --apikey    - Syncthing API Key
  --target    - URL of the Syncthing target
```

## Legal

* Copyright 2022 Google LLC
* Licensed under Apache 2.0
* This is not an official Google product
