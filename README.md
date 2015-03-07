
# Consul Loader

[![Build Status](https://travis-ci.org/natebrennand/consul_loader.svg?branch=master)](https://travis-ci.org/natebrennand/consul_loader)



Consul Loader creates is a simple way to read [Consul](http://consul.io) KV configurations from and into JSON files.
This allows for easy transfer of configurations as well as easier editing control of configurations.



## Using


```
$ ./consul_loader -h
Usage of ./consul_loader:
  -destJSON="": file to export values to
  -destKey="": key to move values to
  -srcJSON="": file to import values from
  -srcKey="": key to move values from
```




#### Examples

Loading "data.json" into the key, "density":
```
./consul_loader -srcJSON data.json -destKEY=density
```

Moving "density-test" to "density":
```
./consul_loader -srcKey density-test -destKEY density
```





## Installation

```bash
go get github.com/natebrennand/consul_loader
```


