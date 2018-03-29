// Edit this variable if you are not running LightningTip on the same domain or IP address as your webserver or not on port 8081
// Don't forget the "/" at the end!
var requestUrl = window.location.protocol + "//" + window.location.hostname + ":8081/";

// To prohibit multiple requests at the same time
var running = false;

var invoice;
var qrCode;

var defaultGetInvoice;

// Fixes weird bug which moved the button up one pixel when its content was changed
window.onload = function () {
    var button = document.getElementById("lightningTipGetInvoice");

    button.style.height = button.clientHeight + "px";
    button.style.width = button.clientWidth + "px";
};

// TODO: maybe don't show full invoice
// TODO: show price in dollar?
function getInvoice() {
    if (running === false) {
        running = true;

        var tipValue = document.getElementById("lightningTipAmount");

        if (tipValue.value !== "") {
            if (!isNaN(tipValue.value)) {
                var data = JSON.stringify({"Amount": parseInt(tipValue.value), "Message": document.getElementById("lightningTipMessage").value});

                var request = new XMLHttpRequest();

                request.onreadystatechange = function () {
                    if (request.readyState === 4) {
                        try {
                            var json = JSON.parse(request.responseText);

                            if (request.status === 200) {
                                console.log("Got invoice: " + json.Invoice);
                                console.log("Invoice expires in: " + json.Expiry);

                                var hash = sha256(json.Invoice);

                                console.log("Got hash of invoice: " + hash);

                                console.log("Starting listening for invoice to get settled");

                                listenInvoiceSettled(hash);

                                // Update UI
                                invoice = json.Invoice;

                                var wrapper = document.getElementById("lightningTip");

                                wrapper.innerHTML = "<a>Your tip request</a>";
                                wrapper.innerHTML += "<textarea class='lightningTipInput' id='lightningTipInvoice' onclick='copyToClipboard(this)' readonly>" + json.Invoice + "</textarea>";
                                wrapper.innerHTML += "<div id='lightningTipQR'></div>";

                                resizeInput(document.getElementById("lightningTipInvoice"));

                                wrapper.innerHTML += "<div id='lightningTipTools'>" +
                                    "<button class='lightningTipButton' id='lightningTipGetQR' onclick='showQRCode()'>QR</button>" +
                                    "<button class='lightningTipButton' id='lightningTipOpen'>Open</button>" +
                                    "<a id='lightningTipExpiry'></a>" +
                                    "</div>";

                                starTimer(json.Expiry, document.getElementById("lightningTipExpiry"));

                                document.getElementById("lightningTipTools").style.height = document.getElementById("lightningTipGetQR").clientHeight + "px";

                                document.getElementById("lightningTipOpen").onclick = function () {
                                    location.href = "lightning:" + json.Invoice;
                                };

                                running = false;

                            } else {
                                showErrorMessage(json.Error);
                            }

                        } catch (exception) {
                            console.error(exception);

                            showErrorMessage("Failed to reach backend");
                        }

                    }

                };

                request.open("POST", requestUrl + "getinvoice", true);
                request.send(data);

                var button = document.getElementById("lightningTipGetInvoice");

                defaultGetInvoice = button.innerHTML;

                button.innerHTML = "<div class='spinner'></div>";

            } else {
                showErrorMessage("Tip amount must be a number");
            }

        } else {
            showErrorMessage("No tip amount set");
        }

    } else {
        console.warn("Last request still pending");
    }

}

function listenInvoiceSettled(hash) {
    try {
        var eventSrc = new EventSource(requestUrl + "eventsource");

        eventSrc.onmessage = function (event) {
            if (event.data === hash) {
                console.log("Invoice settled");

                eventSrc.close();

                showThankYouScreen();
            }

        };

    } catch (e) {
        console.info(e);
        console.warn("Your browser does not support EventSource. Sending a request to the server every two second to check if the invoice is settled");

        var interval = setInterval(function () {
            var request = new XMLHttpRequest();

            request.onreadystatechange = function () {
                if (request.readyState === 4 && request.status === 200) {
                    var json = JSON.parse(request.responseText);

                    if (json.Settled) {
                        console.log("Invoice settled");

                        clearInterval(interval);

                        showThankYouScreen();
                    }

                }

            };

            request.open("POST", requestUrl + "invoicesettled", true);
            request.send(JSON.stringify({"InvoiceHash": hash}))

        }, 2000);

    }

}

function showThankYouScreen() {
    var wrapper = document.getElementById("lightningTip");

    wrapper.innerHTML = "<p id=\"lightningTipLogo\">⚡</p>";
    wrapper.innerHTML += "<a id='lightningTipFinished'>Thank you for your tip!</a>";
}

function starTimer(duration, element) {
    showTimer(duration, element);

    var interval = setInterval(function () {
        if (duration > 1) {
            duration--;

            showTimer(duration, element);

        } else {
            showExpired();

            clearInterval(interval);
        }

    }, 1000);

}

function showTimer(duration, element) {
    var seconds = Math.floor(duration % 60);
    var minutes = Math.floor((duration / 60) % 60);
    var hours = Math.floor((duration / (60 * 60)) % 24);

    seconds = addLeadingZeros(seconds);
    minutes = addLeadingZeros(minutes);

    if (hours > 0) {
        element.innerHTML = hours + ":" + minutes + ":" + seconds;

    } else {
        element.innerHTML = minutes + ":" + seconds;
    }

}

function showExpired() {
    var wrapper = document.getElementById("lightningTip");

    wrapper.innerHTML = "<p id=\"lightningTipLogo\">⚡</p>";
    wrapper.innerHTML += "<a id='lightningTipFinished'>Your tip request expired!</a>";
}

function addLeadingZeros(value) {
    return ("0" + value).slice(-2);
}

function showQRCode() {
    var element = document.getElementById("lightningTipQR");

    if (!element.hasChildNodes()) {
        // Show the QR code
        console.log("Showing QR code");

        // QR code was not shown yet
        if (qrCode == null) {
            createQRCode(10);
        }

        element.style.marginBottom = "1em";
        element.innerHTML = qrCode;

        var size = document.getElementById("lightningTipInvoice").clientWidth + "px";

        var qrElement = element.children[0];

        qrElement.style.height = size;
        qrElement.style.width = size;

    } else {
        // Hide the QR code
        console.log("Hiding QR code");

        element.style.marginBottom = "0";
        element.innerHTML = "";

    }

}

function createQRCode(typeNumber) {
    try {
        console.log("Creating QR code with type number: " + typeNumber);

        var qr = qrcode(typeNumber, "L");

        qr.addData(invoice);
        qr.make();

        qrCode = qr.createImgTag(6, 6);

    } catch (e) {
        console.log("Overflow error. Trying bigger type number");

        createQRCode(typeNumber + 1);
    }

}

function copyToClipboard(element) {
    element.select();

    document.execCommand('copy');

    console.log("Copied invoice to clipboard");
}

function showErrorMessage(message) {
    running = false;

    console.error(message);

    var error = document.getElementById("lightningTipError");

    error.parentElement.style.marginTop = "0.5em";
    error.innerHTML = message;

    var button = document.getElementById("lightningTipGetInvoice");

    // Only necessary if it has a child (div with class spinner)
    if (button.children.length !== 0) {
        button.innerHTML = defaultGetInvoice;
    }

}

function resizeInput(element) {
    element.style.height = "auto";
    element.style.height = (element.scrollHeight) + "px";
}
