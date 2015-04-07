# kiri

<img src="https://mail.lavaboom.com/img/Lavaboom-logo.svg" align="right" width="200px" />

Service discovery library based on etcd. Contains a format compatible with puro.

One of the basic part's of Lavaboom's cloud setup. It was not designed with performance
in mind, but we will eventually rework it to reduce execution locks.

## Requirements

 - etcd

## Examples 

### service example

```go
sd := kiri.New([]string{"etcd address"})
sd.Store(kiri.Puro, "/puro/backends")
sd.Store(kiri.Default, "/kiri")
sd.Register("api-master", "10.10.20.38:8000", nil)
```

### query servers

```go
sd := kiri.New([]string{
    "http://10.10.20.49:4001",
    "http://10.10.20.50:4001",
    "http://10.10.20.51:4001",
})
sd.Store(kiri.Default, "/kiri")

matched, err := sd.Query("api-master", nil)
if err != nil {
    log.Fatal(err)
}

for _, match := range matched {
    log.Print(match.Address)
}
```

### discover a server

```go
sd := kiri.New([]string{
    "http://10.10.20.49:4001",
    "http://10.10.20.50:4001",
    "http://10.10.20.51:4001",
})
sd.Store(kiri.Default, "/kiri")

var conn *r.Connection
err := sd.Discover("rethinkdb", nil, func(service *Service) {
    var err error
    conn, err = r.NewConnection(&r.ConnectOpts{
        Address: service.Address,
    })
    // If err != nil, then it'll remove that server from the service discovery
    return err
})
if err != nil {
    log.Fatal(err)
}
```

## License

This project is licensed under the MIT license. Check `license` for more
information.