# PECK

A simple REST API for Postgresql. Its intent is to enable the use of SQL and abstracting away the need for specific database drivers, to simplify common CRUD operations with REST calls, and handle errors and transactions automatically.


Peck was not designed to be used by clients directly. but was designed to be used by an internal API.

```
[Client] <--> [API] <--> [Peck] <--> [database]
```


# Documentation

See the wiki [https://github.com/jakobii/peck/wiki](https://github.com/jakobii/peck/wiki)