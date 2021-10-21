# chaos: A Simple Poller on BTCUSD with API

## Build Environment
 * Go 1.16+

### Arch Linux
```
	pacman -Syu go docker git
	systemctl enable docker
	systemctl start docker
```  

### Fedora 34 Linux
```
	dnf install golang moby-engine git
	systemctl enable docker
	systemctl start docker
	
```

### Compilation
```
	git clone https://github.com/cylau0/chaos
	cd chaos
	go mod tidy
	go build
```

### Running 
 * First prepare a mongodb database server 
```
	mkdir -p ${HOME}/db
	docker run -d -v ${HOME}/db:/data/db -p 27017:27017 mongo:4.4
```

 * Then build the test image
```
	docker build -t test-server .
```

 * Prepare **env** file
```
	DOCKER_NIC_IP=$(ip a show dev docker0 | grep inet | awk '{print $2}' | sed 's|/.*||')
	echo MONGODB_URL=mongodb://${DOCKER_NIC_IP}:27017 > env
```

 * Run the test server
```
	docker run -d --rm -p 80:80 --env-file env test-server
```	

### Run Test 
 * Run following command
```
	go test
```

 * Sample output
```
	2021/10/22 02:24:28 TestAPIServer
	2021/10/22 02:24:28 Insert : {<nil> BTC USD Yobit 64000 2021-08-01 00:00:00 +0000 UTC 0}
	2021/10/22 02:24:28 Insert : {<nil> BTC USD Yobit 63000 2021-08-01 00:01:00 +0000 UTC 0}
	2021/10/22 02:24:28 Insert : {<nil> BTC USD Yobit 62000 2021-08-01 00:02:00 +0000 UTC 0}
	2021/10/22 02:24:28 Insert : {<nil> BTC USD Yobit 61000 2021-08-01 00:03:00 +0000 UTC 0}
	2021/10/22 02:24:28 Insert : {<nil> BTC USD Yobit 60000 2021-08-01 00:04:00 +0000 UTC 0}
	2021/10/22 02:24:28 Insert : {<nil> BTC USD Yobit 59000 2021-08-01 00:05:00 +0000 UTC 0}
	2021/10/22 02:24:30 Using URL = http://localhost:8080/price
	2021/10/22 02:24:30 "GET http://localhost:8080/price HTTP/1.1" from [::1]:44642 - 200 44B in 68.464µs
	2021/10/22 02:24:30 Using URL = http://localhost:8080/price/2021-08-01T00:02:30
	2021/10/22 02:24:30 "GET http://localhost:8080/price/2021-08-01T00:02:30 HTTP/1.1" from [::1]:44642 - 200 44B in 64.111µs
	2021/10/22 02:24:30 Using URL = http://localhost:8080/average?from=2021-08-01T00:00:00&to=2021-08-01T00:05:00
	2021/10/22 02:24:30 "GET http://localhost:8080/average?from=2021-08-01T00:00:00&to=2021-08-01T00:05:00 HTTP/1.1" from [::1]:44642 - 200 74B in 56.77µs
	PASS
	ok      github.com/cylau0/chaos 2.008s
```


