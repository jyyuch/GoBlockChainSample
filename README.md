# GoBlockChainSample
try web3 API, block chain index

# Require
## go
```shell
go version                        
# go version go1.15.5 darwin/amd64
```

## Docker Desktop
Ref: https://www.docker.com/products/docker-desktop
version: 3.6.0（3.6.0.5487）

## PostgreSQL
```shell
docker run -p 5432:5432 --name eth-block-indexer -e POSTGRES_PASSWORD=mysecretpassword -d postgres
```
1. user name: postgres
1. password: mysecretpassword
1. db name: postgres

**PS: need check port 5432 is available**

# build
```shell
go build
```

# run
```shell
go run main.go
```