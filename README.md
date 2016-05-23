# consul IP address manager

consul IPAM allocates IPv4 and IPv6 addresses out of a specified address range.

Plugin is based on the code of [CNI IPAM host-local plugin](https://github.com/containernetworking/cni/tree/master/plugins/ipam/host-local) but with different Backend type.
Allocator code was taken as is, as long as this README. Main.go has slight modifications. Config.go was moved into separate module to share the configuration with backend code.

## Usage

### Obtain an IP

Given the following network configuration:

```
{
    "name": "default",
    "ipam": {
        "type": "consul",
        "consul_addr": "127.0.0.1",
        "consul_port": "8500",
        "dc": "dc1",
        "subnet": "203.0.113.0/24"
    }
}
```

#### Using the command line interface

```
$ export CNI_COMMAND=ADD
$ export CNI_CONTAINERID=f81d4fae-7dec-11d0-a765-00a0c91e6bf6
$ ./consul < $conf
```

```
{
    "ip4": {
        "ip": "203.0.113.1/24"
    }
}
```

## Backends

By default ipmanager stores IP allocations on the local filesystem using the IP address as the file name and the ID as contents. For example:

```
$ ls /var/lib/cni/networks/default
```
```
203.0.113.1	203.0.113.2
```

```
$ cat /var/lib/cni/networks/default/203.0.113.1
```
```
f81d4fae-7dec-11d0-a765-00a0c91e6bf6
```

## Configuration Files


```
{
	"name": "ipv6",
    "ipam": {
		    "type": "consul",
        "consul_addr": "127.0.0.1",
        "consul_port": "8500",
        "dc": "dc1",
        "subnet": "3ffe:ffff:0:01ff::/64",
        "range-start": "3ffe:ffff:0:01ff::0010",
        "range-end": "3ffe:ffff:0:01ff::0020",
        "routes": [
          { "dst": "3ffe:ffff:0:01ff::1/64" }
        ]
	}
}
```

```
{
  "name": "ipv4",
	"ipam": {
		"type": "consul",
    "consul_addr": "127.0.0.1",
    "consul_port": "8500",
    "dc": "dc1",
		"subnet": "203.0.113.1/24",
		"range-start": "203.0.113.10",
		"range-end": "203.0.113.20",
		"routes": [
			{ "dst": "203.0.113.0/24" }
		]
	}
}
```
