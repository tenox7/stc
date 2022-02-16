# STC - Syncthing Cli

This is an unofficial command line tool for [Syncthing](https://syncthing.net/).
Stc can be used to quicky check status of Syncthing from a terminal / command line
without need of a Web Browser. For example over a remote ssh connection without port
forwarding, or if you have large number of machines to query.

## Usage

The tool allows to connect to a remote endpoint by using `--target=` flag. 
By default `http://127.0.0.1:8384`.

You have to specify API Key by using `--api_key=` flag. The API Key can
be obtained from the GUI or `config.xml` (`grep apikey config.xml`).

In the current version stc will display folder and device status information

## Flags

```text
  --api_key   - Syncthing api key
  --target    - URL of the Syncthing target
```

## Legal

* Copyright 2022 Google LLC
* Licensed under Apache 2.0
* This is not an official Google product
