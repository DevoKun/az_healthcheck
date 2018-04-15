# Availability Zone (AZ) Healthcheck



## Summary

* Monitors all endpoints in an AZ to provide an end-to-end healthcheck status.
* Provides an HTTP webservice that can be consumed by an ELB or Route53 healthcheck.
* Writes a healthcheck status file on the local filesystem that can be consumed by monitoring software.

![AZ Healthcheck Diagram](https://raw.githubusercontent.com/DevoKun/az_healthcheck/master/azhealthcheck.png)



## Configuration

* Configuration is done using a simple **YAML** file called **azhealthcheck.yaml**
* **azhealthcheck** will look for the configuration file in **/etc/azhealthcheck.yaml** or **$(pwd)/azhealthcheck.yaml**



**FILE**: azhealthcheck.yaml

```yaml
---
allowed_failed_checks: 0
options:
  status_file_name: '/var/run/az_health_check.status'

hosts:
  identme:
    url: "http://ident.me/"
    headers:
      "X-Browser-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/56.0.2924.87 Safari/537.36"

  slashdot:
    url: "https://www.slashdot.org/"

```



### Support for MutualSSL *(Client Certs)*

* Normally when connecting to a remote https server endpoint the only requirement is that the server have valid SSL/TLS certificates in use.
* If the server requires that the client be connecting using valid SSL/TLS certificates, azhealthcheck can support the requirement.
* azhealthcheck supports the use of PEM formatted SSL certificate files.
* Pass the location to the Client Certificate and Key files in as **clientcertfilename** and **clientkeyfilename** variables in the per-host yaml configuration

```yaml

hosts:

  apache2:
    name: apache2
    url: https://0.0.0.0/
    clientcertfilename:   /etc/ssl/certs/test.crt
    clientkeyfilename:    /etc/ssl/private/test.key

```



#### MutualSSL Enabled in Apache2

```apa
  ### Apache2 VHost MutualSSL Support Enabled
  SSLVerifyClient      require
  SSLVerifyDepth       2
  SSLCACertificateFile "/etc/ssl/certs/mutualssl_ca.pem"
```



#### MutualSSL Support Enabled in AZHealthcheck on a per-host basis

```yaml

---
browserAgent: azh
check_mk_service_name: azhealthcheck
checkInterval: 3000
port: 3000

hosts:

  prodFrontend:
    name: prodFrontend
    url: https://frontend.production/healthcheck
    headers:
      'X-Browser-Agent': 'AZ HealthCheck'
    clientcertfilename:   /etc/ssl/certs/test.crt
    clientkeyfilename:    /etc/ssl/private/test.key

  prodBackend
    name: prodBackend
    url: https://backend.production/healthcheck
    headers:
      'X-Browser-Agent': 'AZ HealthCheck'
    clientcertfilename:   /etc/ssl/certs/test.crt
    clientkeyfilename:    /etc/ssl/private/test.key

```





## Run with Supervisor

```bash

apt-get install -y supervisor


cat << EOF > /etc/supervisor/supervisord.conf
[unix_http_server]
file=/var/run/supervisor.sock
chmod=0700

[supervisord]
logfile=/var/log/supervisor/supervisord.log
pidfile=/var/run/supervisord.pid
childlogdir=/var/log/supervisor
nodaemon=false

[rpcinterface:supervisor]
supervisor.rpcinterface_factory = supervisor.rpcinterface:make_main_rpcinterface

[supervisorctl]
serverurl=unix:///var/run/supervisor.sock

[include]
files = /etc/supervisor/conf.d/*.conf
EOF


cat << EOF > /etc/supervisor/conf.d/azhealthcheck.conf
[program:azhealthcheck]
command        = /usr/local/bin/azhealthcheck
startsecs      = 5
stopwaitsecs   = 3600
stopasgroup    = false
killasgroup    = true
stdout_logfile = /var/log/azhealthcheck-stdout.log
stderr_logfile = /var/log/azhealthcheck-stderr.log
EOF


service supervisor restart

```





## Testing

### Start an HTTP listener on port 80

```bash
sudo python -m SimpleHTTPServer 80
```



### Start an HTTP listener on port 8080

```shell
sudo python -m SimpleHTTPServer 8080
```



### Start AZHealthCheck

* The **defaults** will check for two HTTP services running on localhost **tcp:80** and **tcp:8080**.

```shell
./azhealthcheck
```


