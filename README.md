# gomod_vanity_daemon

Super simple Go daemon for supplying go-import HTML meta tags in response to go get requests, in order to allow so-called 'vanity' URLs. This is less featureful than many other solutions, but I thought it might be useful for some. It is intended to run on a UNIX-like server (tested on Linux and FreeBSD), on the same machine as something like Apache or nginx, which should proxy go get requests to it. For example, nginx can be configured with its proxy_pass parameter to send all requests with "?go-get=1" appended to gomod_vanity_daemon, by placing the following inside a location directive in nginx.conf like so:

```
if ($args = "go-get=1") {
    proxy_pass http://127.0.0.1:8000;
}
```

The daemon basically allows the user to map all golang modules of the form domain.name/modulename to an online git repo programmatically. It removes the need to have separate meta tags defined for each module and repo. For example, if the user has three go modules with the names:

```
example.com/moduleone
example.com/moduletwo
example.com/modulethree
```

gomod_vanity_daemon will automatically translate all of these to their respective git paths, and produce an output (using moduleone as an example) of the form:

```html
<meta name="go-import" content="example.com/moduleone git https://github.com/ExampleUser/moduleone">
```

It is also smart enough to deal with semantic versioning so that modules with /v2 etc. will also be placed in a meta tag correctly. To start the daemon in the foreground and not as a service, run with the -no_daemon flag. To run the service manually without the service scripts below, it must execute as root, after which it will drop all of its privileges permanently.

## Building

To build to your own requirements, clone this repo and then inside the source directory, open gomod_vanity_daemon.go and change the following to your requirements:

```golang
const (
        domainName = "example.com"
        gitRoot = "github.com/ExampleUser"
        portNumber = "8000"
)
```

Then run:

```
go build
```

## Installing + starting on boot

Copy the gomod_vanity_daemon binary to a central location such as /usr/local/bin. To execute as a service, write a servece script for your respective init system - I've included examples that can be used with FreeBSD init and Linux's systemd, in the rc_scripts folder. To install on either (assuming installation in /usr/local/bin) follow the relevant section (as root):

FreeBSD:
```
cp rc_scripts/freebsd/gomod_vanity_daemon /usr/local/etc/rc.d
chmod +x /usr/local/etc/rc.d/gomod_vanity_daemon
sysrc gomod_vanity_daemon_enable="YES"
```

Linux (systemd):
```
cp rc_scripts/linux_systemd/gomod_vanity_daemon.service /etc/systemd/system
systemctl daemon-reload
systemctl enable gomod_vanity_daemon.service
```

After either of these, the gomod_vanity_daemon daemon will be enabled for starting on bootup. To start immediately, use the respective service command:

FreeBSD:
```
service gomod_vanity_daemon start
```

Linux (systemd):
```
systemctl start gomod_vanity_daemon.service
```
