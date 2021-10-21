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

## Compilation
### Build **chaos** binary
```
	git clone https://github.com/cylau0/chaos
	cd chaos
	go mod tidy
	go build
```

### Building a Docker Image
```
	docker build -t test-server .
```


## Prepare prepare a MongoDB server up and running, listing to 27017 port
```
	mkdir -p ${HOME}/db
	docker run -d -v ${HOME}/db:/data/db -p 27017:27017 mongo:4.4
```

## Running Server

### Running the the compiled binary **chaos**
 * Root privilege is required for the server listening on port 80
```
	sudo ./chaos
```

### Running the Docker Image test-server
 * Prepare **env** file
```
	DOCKER_NIC_IP=$(ip a show dev docker0 | grep inet | awk '{print $2}' | sed 's|/.*||')
	echo MONGODB_URL=mongodb://${DOCKER_NIC_IP}:27017 > env
```

 * Run the test-server image
```
	docker run -d --rm -p 80:80 --env-file env test-server
```	

## RestFul API Usage

### Enquiry Latest Price
```
	curl http://localhost/price
```

 * Sample Output
```
{"ts":"2021-10-21T18:49:36.278Z","price":63501}
```

### Enquiry Price at **2021-10-21 18:47:00 UTC**
 * Linear Intepolation will be done from 2 consecutive sampling points containing the given Time
```
	curl http://localhost/price/2021-10-21T18:47:00
```

 * Sample Output
```
{"ts":"2021-10-21T18:47:00Z","price":63501}
```

### Enquiry Average Price Between  **2021-10-21 18:44:00 UTC** and **2021-10-21 18:47:00 UTC**
 * Linear Intepolation will be done if needed
```
	curl 'http://localhost/average?from=2021-10-21T18:44:00&to=2021-10-21T18:47:00'
```
 * Sample Output, the Average Price 
```
{"from":"2021-10-21T18:44:00Z","to":"2021-10-21T18:47:00Z","price":63609.151426606564}
```

## Run Test 
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


