# LightningTip
A simple way to accept tips via the Lightning Network on your website. If want to tip me you can find my instance of LightningTip [here](https://michael1011.at/lightning.html).

[robclark56](https://github.com/robclark56) forked LightningTip and rewrote the backend in **PHP**. His fork is called [LightningTip-PHP](https://github.com/robclark56/lightningtip) and is a great alternative if you are not able to run the executable.

<img src="https://i.imgur.com/0mOEgTf.gif" width="240">

## How to install

To get all necessary files for setting up LightningTip you can either [download a prebuilt version](https://github.com/michael1011/lightningtip/releases) or [compile from source](#how-to-build).

LightningTip is using [LND](https://github.com/lightningnetwork/lnd) as backend. Please make sure it is installed and fully synced before you install LightningTip.

The default config file location is `$HOME/.lightningtip/lightningTip.conf`. The [sample config](https://github.com/michael1011/lightningtip/blob/master/sample-lightningTip.conf) contains everything you need to know about the configuration. To use a custom config file location use the flag `--config filename`. You can use all keys in the config as command line flag. Command line flags *always* override values in the config.

The next step is embedding LightningTip on your website. Upload all files excluding `lightningTip.html` to your webserver. Copy the contents of the head tag from `lightningTip.html` into the head section of the HTML file you want to show LightningTip in. The div below the head tag is LightningTip itself. Paste it into any place in the already edited HTML file on your server.

There is a light theme available for LightningTip. If you want to use it **add** this to the head tag of your HTML file:

```html
<link rel="stylesheet" href="lightningTip_light.css">
```

**Do not use LightningTip on XHTML** sites. That causes some weird scaling issues.

Make sure that the executable of **LightningTip is always running** in the background. It is an [API](https://github.com/michael1011/lightningtip/wiki/API-documentation) to connect LND and the widget on your website. **Do not open the URL you are running LightingTip on in your browser.** All this will do is show an error.

If you are not running LightningTip on the same domain or IP address as your webserver, or not on port 8081, change the variable `requestUrl` (which is in the first line) in the file `lightningTip.js` accordingly.

When using LightningTip behind a proxy make sure the proxy supports [EventSource](https://developer.mozilla.org/en-US/docs/Web/API/EventSource). Without support for it the users will not see the "Thank you for your tip!" screen.

That's it! The only two things you need to take care about is keeping the LND node online and making sure that your incoming channels are sufficiently funded to receive tips. LightningTip will take care of everything else.

## How to build

First of all make sure [Golang](https://golang.org/) and [Dep](https://github.com/golang/dep) are both correctly installed. Golang version 1.10 or newer is recommended.

If you want to return a picture, make sure you install ImageMagick MagickWand C API. For examnple (Ubuntu/Debian): sudo apt-get install libmagickwand-dev

To enable it, edit Makefile and enable the line with IMAGICK : = -tags imagick 
```
go get -d github.com/michael1011/lightningtip
cd $GOPATH/src/github.com/michael1011/lightningtip


make && make install
```

To start run `$GOPATH/bin/lightningtip` or follow the instructions below to setup a service to run LightningTip automatically.

## Upgrading

Make sure you stop any running LightningTip process before upgrading, then pull from source as follows:

```
cd $GOPATH/src/github.com/michael1011/lightningtip
git pull

make && make install
```

## Starting LightningTip Automatically

LightningTip can be started automatically via Systemd, or Supervisord, as outlined in the following wiki documentation:

* [Running LightningTip with systemd](https://github.com/michael1011/lightningtip/wiki/Running-LightningTip-with-systemd)
* [Running LightningTip with supervisord](https://github.com/michael1011/lightningtip/wiki/Running-LightningTip-with-supervisord)

## Reverse Proxy Recipes

In instances where the default LightningTip SSL configuration options are not working, you may want to explore running a reverse proxy to LightningTip as outlined in the following wiki documentation:

* [LightningTip via Apache2 reverse proxy](https://github.com/michael1011/lightningtip/wiki/LightningTip-via-Apache2-reverse-proxy)
* [LightningTip via Nginx reverse proxy](https://github.com/michael1011/lightningtip/wiki/LightningTip-via-Nginx-reverse-proxy)
