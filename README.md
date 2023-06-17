# Simple Network Latency Monitor
A simple network latency monitor written in go.
It exposes a metrics endpoint with all monitored ip's latency as lables to `ip_latency`. The intent is to scrape the metrics with grafana agent and send wherever you'd like (I use [grafana cloud](https://grafana.com/products/cloud/) )


## Usage:
Add the ip's you'd like to monitor to the `iplist.txt`, each line is a new IP address:
```
8.8.8.8
8.8.4.4
1.1.1.1
```

Compile and run
metrics will be available at http://localhost:8082/metrics


### Run it a service on Ubuntu:

Once you've tested it, Build the application:
```
go build -o net_monitor
```

Create a limited user to run the service under:
```
sudo useradd -r -s /bin/false netmon
```

Create the service file `/etc/systemd/system/net_monitor.service` with the following content:
```
[Unit]
Description=Net Monitor service

[Service]
ExecStart=/path/to/your/net_monitor
WorkingDirectory=/path/to/your
Restart=always
User=netmon
Group=nogroup
Environment=PATH=/usr/bin:/usr/local/bin

[Install]
WantedBy=multi-user.target
```
Adjust the permissions on setcap so a limited user has the permission to open raw sockets without running as root:
```
sudo setcap cap_net_raw+ep /path/to/your/net_monitor
```

Then you can start your service and enable it on boot:
```
sudo systemctl start net_monitor
sudo systemctl enable net_monitor

```
