# kiri
puro-compatible etcd-based golang service discovery

```go
sd := kiri.New([]string{"etcd address"})
sd.Store(kiri.Puro, "/puro/backends")
sd.Store(kiri.Default, "/kiri")
sd.Register("api-master", "10.10.20.38:8000", nil)
```