
# Consul Loader

[![Build Status](https://travis-ci.org/natebrennand/consul_loader.svg?branch=master)](https://travis-ci.org/natebrennand/consul_loader)



Consul Loader creates is a simple way to read [Consul](http://consul.io) KV configurations from and into JSON files.
This allows for easy transfer of configurations as well as easier editing control of configurations.
Consul is very friendly but can be tedious to rename large sets of keys.


Tested with Consul 0.4.1 & 0.5.1.


NOTE: when exporting to JSON, all values will be in a string format.
This does not matter when inserting to Consul.


## Using


```
$ ./consul_loader -h
Usage of ./consul_loader:
  -destJSON="": file to export values to
  -destKey="": key to move values to
  -rename=false: place as a rename instead of a insertion
  -srcJSON="": file to import values from
  -srcKey="": key to move values from
```




#### Examples




Loading "data.json" into the key, "density":
```
./consul_loader -srcJSON data.json -destKEY=density
```

before:
```js
data.json = {
  "key": "value",
  "number": 2
}
consul = {}
```

after:
```js
consul = {
  "key": "value",
  "number": 2
}
```








--------

Renaming "density-test" to "density":
```
./consul_loader -srcKey density-test -destKEY density -rename true
```



before:
```js
consul = {
  "density-test": {
    "key": "value",
    "number": 2
  }
}
```

after:
```js
consul = {
  "density": {
    "key": "value",
    "number": 2
  }
}
```











-----------

Adding "database" to "density":
```
./consul_loader -srcKey database -destKEY density
```





before:
```js
consul = {
  "density": {
    "key": "value",
    "number": 2
  },
  "database": {
    "host": "localhost",
    "port": 6379
  }
}
```

after:
```js
consul = {
  "density": {
    "key": "value",
    "number": 2,
    "database": {
      "host": "localhost",
      "port": 6379
    }
  },
  "database": {
    "host": "localhost",
    "port": 6379
  }
}
```







## Installation

```bash
go get github.com/natebrennand/consul_loader
```


