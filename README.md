# Foreman
It is a [foreman](https://github.com/ddollar/foreman) implementation in GO.

## Description
Foreman is a manager for [Procfile-based](https://en.wikipedia.org/wiki/Procfs) applications. Its aim is to abstract away the details of the Procfile format, and allow you to run your services directly.

## Features
- Run procfile-backed apps.
- Able to run with dependency resolution.

## Procfile
Procfile is simply `key: value` format like:
```yaml
app1:
    cmd: ping -c 1 google.com 
    run_once: false 
    checks:
        cmd: ps aux | grep google
    deps: 
        - redis
app2:
    cmd: ping -c 5 yahoo.com
    checks:
        cmd: ps aux | grep yahoo

redis:
    cmd: redis-server --port 6010
    checks:
        cmd: redis-cli -p 6010 ping
        tcp_ports: [6010]
```
**Here** we defined three services `app1`, `app2` and `redis` with check commands and dependency matrix

## How to use
**First:** modify the procfile with processes or services you want to run.

**second**: Simply run with command: 
```sh
$ ./foreman
```

