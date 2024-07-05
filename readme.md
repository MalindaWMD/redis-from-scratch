# Implementation of [Build Redis from scratch](https://www.build-redis-from-scratch.dev/en/introduction) series

This is experimental implementation of [Build Redis from scratch](https://www.build-redis-from-scratch.dev/en/introduction) series i've done for understand how Redis works under the hood.

Addition to the  features covered by the article,

- this repo supports following commands
    - HGETALL
    - DEL
    - EXPIRE
    - TTL

- TTL support. For new entries and existing entries.
- aof is rewritten if the configured size has reached. Size can be configured via a .conf file. 
    Only the latest values are transfered to the new file.