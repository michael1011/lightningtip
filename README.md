# LightningTip
A simple way to accept tips via the Lightning Network on your website.

<img src="https://i.imgur.com/tTQnnoJ.gif" width="240">

## How to install
To get all necessary files for setting up LightningTip you can either [download a prebuilt version](https://github.com/michael1011/lightningtip/releases) or [compile from source](#how-to-install).


LightningTip is using [LND](https://github.com/lightningnetwork/lnd) as backend. Please make sure it is installed and fully synced before you install LightningTip.


The default config file location is `lightningTip.conf` in the directory you are executing LightningTip in. The [sample config](https://github.com/michael1011/lightningtip/blob/master/sample-lightningTip.conf) contains everything you need to know about the configuration. To use a custom config file location use the flag `--config filename`. You can use all keys in the config as command line flag. Command line flags *always* override values in the config.


Embedding LightningTip is also quite easy. Upload all files excluding `lightningTip.html` to your webserver. Copy the contents of the head tag of the before mentioned HTML file into a HTML file you want to show LightningTip in. The div below the head tag is LightningTip itself. Paste it into any place in the already edited HTML file on you server.


Make sure that the executable of **LightningTip is always running** in the background. It connects LND and the widget on your website. A sample configuration for systemd can be found [on this wiki page](https://github.com/michael1011/lightningtip/wiki/Running-LightningTip-with-systemd).


If you are not running LightningTip on the same domain or IP address as your webserver or not on port 8081 change the variable `requestUrl` (which is in the first line) in the file `lightningTip.js` accordingly.


When using LightningTip behind a proxy make sure the proxy supports [EventSource](https://developer.mozilla.org/en-US/docs/Web/API/EventSource). Without support for it the users will not see the "Thank you for your tip!" screen.


That's it! The only two things you need to take care about is keeping the LND node online and making sure that your channels are funded well enough to receive tips. LightningTip will take care of everything else.

## How to build
First of all make sure [Golang](https://golang.org/) and [Dep](https://github.com/golang/dep) are both correctly installed. Golang version 1.10 or newer is recommended.

```
go get github.com/michael1011/lightningtip
cd $GOPATH/src/github.com/michael1011/lightningtip
dep ensure
go install
```

## Starting LightningTip Automatically

LightningTip can be started automatically via Systemd, or Supervisord, as outlined in the following wiki documentation:

* [Running LightningTip with systemd](https://github.com/michael1011/lightningtip/wiki/Running-LightningTip-with-systemd)
* [Running LightningTip service with supervisord](https://github.com/michael1011/lightningtip/wiki/Running-LightningTip-service-with-supervisord)

## Reverse Proxy Recipes

In instances where the default LightningTip SSL configuration options are not working, you may want to explore running a reverse proxy to LightningTip as outlined in the following wiki documentation:

* [LightningTip via Apache2 reverse proxy](https://github.com/michael1011/lightningtip/wiki/LightningTip-via-Apache2-reverse-proxy)
* [LightningTip via Nginx reverse proxy](https://github.com/michael1011/lightningtip/wiki/LightningTip-via-Nginx-reverse-proxy)


