# consul IP address manager

Consul IPAM allocates IPv4 and IPv6 addresses out of a specified address range and stores allocations in [Consul](https://www.consul.io/) KV store backend. 

Plugin is based on the code of [CNI IPAM host-local plugin](https://github.com/containernetworking/cni/tree/master/plugins/ipam/host-local) but with different Backend type.
Allocator code was taken as is, as long as this README. Main.go has slight modifications. Config.go was moved into separate module to share the configuration with backend code.

## TODO
- If network already defined from consul use it for allocation regardless setting in the client.
- Create ```*Store``` type struct and modify interface, which will to support disk backend from [CNI IPAM host-local plugin](https://github.com/containernetworking/cni/tree/master/plugins/ipam/host-local) and consul backend depending on the plugin configuration. That will allow to use host-local plugin with different backends. 
- Store mac address in Consul as well. That could be useful for EVPN solution with [BaGPipe CNI plugin](https://github.com/murat1985/bagpipe-bgp).

## Install

### Option 1
Use go get for installation
````
git get github.com/murat1985/cni-ipam-consul
````

Plugin would be install into $GOBIN, e.g.:
```
~/bagpipe/bin/cni-ipam-consul
```

### Option 2
The second way to install plugin altogether with other CNI plugins and IPAM plugins. Clone CNI repositority: [CNI](https://github.com/containernetworking/cni)

Make sure that GOPATH environment variable is set

```
cd $GOPATH
git clone https://github.com/containernetworking/cni
cd cni/plugins/ipam
```

Clone bagpipe CNI plugin into plugins/main/cni-ipam-consul

```
git clone https://github.com/murat1985/cni-ipam-consul cni-ipam-consul
cd ../../
```

Build plugins

```
./build
```

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
