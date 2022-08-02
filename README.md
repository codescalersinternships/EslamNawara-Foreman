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
        cmd: ps aux 
    deps: 
        - redis
app2:
    cmd: ping -c 5 yahoo.com
    checks:
        cmd: ps aux

redis:
    cmd: redis-server --port 5000 
    checks:
        cmd: redis-cli -p 5000 ping
        tcp_ports: [5000]
```
**Here** we defined three services `app1`, `app2` and `redis` with check commands and dependency matrix

## How to use
**First:** modify the procfile with processes or services you want to run.

**second**: Simply run with command: 
```sh
$ ./foreman
```

