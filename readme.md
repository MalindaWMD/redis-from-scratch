# Implementation of [Build Redis from scratch](https://www.build-redis-from-scratch.dev/en/introduction) series

This is experimental implementation of [Build Redis from scratch](https://www.build-redis-from-scratch.dev/en/introduction) series i've done for understand how Redis works under the hood.

Addition to the  features covered by the article,

- this repo supports following commands
    - command 1 
    - command 2
    - command 3

- TTL support 
- aof is rewritten if the configured size has reached. Size can be configured via a .conf file. 
    Only the latest values are transfered to the new file.