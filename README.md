# pinger
Track ping RTT for a list of websites.

## Clone the repository
You can use the `go get` command to fetch this repository into your Go workspace source directory (typically `~/go/src/`).

`go get github.com/perryizgr8/pinger`

## Run the server
On the pi (or any Linux machine) you need to run this command to let our program send ICMP pings.

`sudo sysctl -w net.ipv4.ping_group_range="0 2147483647"`

Read more about this on the `go-ping` GitHub [page](https://github.com/go-ping/ping#linux).

Start the server using `go run .`. This should fetch all dependencies and start listening on port 8080.

## Open the page
Simply point your browser to `localhost:8080` if you're running it on your PC/laptop. Or use your pi's IP address if you are running the server on the pi.

## Blog post
Read a bit about the project in my [blog post](https://perryizgr8.github.io/raspberry-pi/2021/01/06/track-ping-rtt-part1.html).

