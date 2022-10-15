[![Language](https://img.shields.io/badge/Language-Go-blue.svg)](https://golang.org/)
![License](https://img.shields.io/badge/license-Apache-blue.svg)

## WellWood

![WellWood](https://cdn.nlark.com/yuque/0/2022/png/29306112/1656687604705-cefdbe9e-3242-4173-871f-fdb11fcacd83.png)

## build 
```
git clone https://github.com/eraft-io/eraft.git
cd eraft
make
```

## quick start

run ekv server
```
./ekv-server -id 0
./ekv-server -id 1
./ekv-server -id 2
```
## write & read

```
redis-cli -p 17088[leader port]
```
