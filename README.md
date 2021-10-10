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

# Run application
## build
```shell
# build at project root folder
go build

# run
go run main.go
```
Note:
1. Listening and serving HTTP on localhost:8080

# Call API
## API service
```shell
curl 'http://localhost:8080/blocks?limit=1'
curl 'http://localhost:8080/blocks/13064886'
curl 'http://localhost:8080/transaction/0xff09cc3d65e71c3ac6c17e253befd62ae298b33f30f30d19e1b99523f2cd91f4'
```

## Ethereum block indexer service
```shell
# start scan
curl 'http://localhost:8080/block_indexer/scan'
```
PS: 
1. it will start scan from block 0
1. each loop scan NUM_BLOCKS_SCAN_ONCE (config.go) by NUM_ROUTINE_TO_SCAN (config.go)
1. and go on next loop

