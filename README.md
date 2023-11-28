# waku-store-request-maker

Proof of concept:

This repo is based on the https://github.com/alrevuelta/waku-publisher one but modified to make it
generate Store requests at certain rate.

Therefore, this project starts a simple go-waku that acts as a Store client and performs
periodic requests to the specified Waku Store nodes.

```
go build
```

Example:
```
./main --pubsub-topic="my-ptopic"  --content-topic="my-ctopic" --queries-per-second=10 --bootstrap-node="enr:-J64QB97yIfff299_1bCzp60X42dxGiJebNp_vpp-JfZ20WHZC--y_z7-Gtqo1-8SHwDK5l1vsttvgtEhfUpy1TQeOMBgmlkgnY0gmlwhAoBAAOKbXVsdGlhZGRyc4CJc2VjcDI1NmsxoQM3Tqpf5eFn4Jztm4gB0Y0JVSJyxyZsW8QR-QU5DZb-PYN0Y3CC6mCDdWRwgiMohXdha3UyAQ" --peer-store-postgres-addr="/ip4/10.1.0.10/tcp/30303/p2p/16Uiu2HAkxcXtrhB9sEVi3wzAoSZgc8ymyxRdgj4jXznHHBdxFsbY" --peer-store-sqlite-addr="/ip4/10.1.0.9/tcp/30304/p2p/16Uiu2HAmD2W31nH32oCmLw1Lc7qCg8VCFmuoaSNqyRvsJBmnNxDk"
```

Docker images are available, last 7 digits of the commit: ivansete/waku-store-request-maker:90d6a6
