# Implementation of [Build Redis from scratch](https://www.build-redis-from-scratch.dev/en/introduction) series

This is experimental implementation of [Build Redis from scratch](https://www.build-redis-from-scratch.dev/en/introduction) series i've done for understand how Redis works under the hood.

Addition to the  features covered by the article,

- this repo supports following commands
    - HGETALL
    - DEL
    - EXPIRE
    - TTL

- TTL support. For new entries and existing entries.
- AOF rewriting. If a certain file size is reached default AOF will be moved to a different file and the default one will be
    re-created. When reading on startup, all the files will be read and handled in a goroutine.