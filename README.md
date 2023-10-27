# iot_detector
A multithreaded scanner for detecting IoT devices, which allows you to identify a device on the Internet based on a specific request to it. This allows you to more accurately determine the device and avoid false positives or honeypot


## Usage
```bash
Usage of ./iot_detector:
  -a    Auto URL scheme
  -l string
        List of ip,port
  -p string
        Ports to scan (e.g. 22,80,443,1000-2000)
  -r string
        Json request file
  -t int
        Thread count (default 1000)
  -v    Verbose
```


## JSON request format
```json
{
    "path": "/cgi-bin/target.cgi",
    "method": "POST",
    "headers": {
        "Content-Type": "text/xml; charset=utf-8",
        "X-Requested-With": "XMLHttpRequest"
    },
    "data": "action=get_device_info&method=1",
    "search": "<DeviceName>DIR-123</DeviceName>"
}
```

* `path` - A specific URL, when accessed, returns the model or version of the device

* `method` - Request method GET or POST

* `headers` - Request header

* `data` - Post request body

* `search` - Success detection pattern


## Run or install
```bash
$ go build .
$ ./iot_detector ...
```
