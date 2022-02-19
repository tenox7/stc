# STC - Syncthing Cli

This is an unofficial command line tool for [Syncthing](https://syncthing.net/).
Stc can be used to quicky check status of Syncthing from a terminal / command line
without need of a Web Browser. For example on a remote machine over ssh, without port
forwarding or if you have large number of machines to query.

```
$ stc --apikey=XXXXXX --target=http://127.0.0.1:8384
Host      Uptime    CPU%      Version
homenas   2 weeks   0.0%      v1.19.0

Folder     Paused    State     Global    Local
pics       false     idle      37 GB     37 GB
docs       false     idle      4 GB      4 GB
backups    false     idle      86 GB     86 GB

Device          Paused    Connected  Complete  Download  Upload
office          false     true       100.0%    11 kB     11 kB
garage          false     false      100.0%    0 B       0 B
jakob-home      false     true       100.0%    89 MB     447 kB
backupnas       false     true       100.0%    6.3 kB    7.0 kB
homenas         false     true       100.0%    0 B       0 B
```

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
