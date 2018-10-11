// Edit this variable if you are not running LightningTip on the same domain or IP address as your webserver or not on port 8081
// Don't forget the "/" at the end!
var requestUrl = window.location.protocol + "//" + window.location.hostname + "/tiprest/";

// To prohibit multiple requests at the same time
var requestPending = false;

var invoice;
var qrCode;

var defaultGetInvoice;

// Data capacities for QR codes with mode byte and error correction level L (7%)
// Shortest invoice: 194 characters
// Longest invoice: 1223 characters (as far as I know)
var qrCodeDataCapacities = [
    {"typeNumber": 9, "capacity": 230},
    {"typeNumber": 10, "capacity": 271},
    {"typeNumber": 11, "capacity": 321},
    {"typeNumber": 12, "capacity": 367},
    {"typeNumber": 13, "capacity": 425},
    {"typeNumber": 14, "capacity": 458},
    {"typeNumber": 15, "capacity": 520},
    {"typeNumber": 16, "capacity": 586},
    {"typeNumber": 17, "capacity": 644},
    {"typeNumber": 18, "capacity": 718},
    {"typeNumber": 19, "capacity": 792},
    {"typeNumber": 20, "capacity": 858},
    {"typeNumber": 21, "capacity": 929},
    {"typeNumber": 22, "capacity": 1003},
    {"typeNumber": 23, "capacity": 1091},
    {"typeNumber": 24, "capacity": 1171},
    {"typeNumber": 25, "capacity": 1273}
];

// TODO: solve this without JavaScript
// Fixes weird bug which moved the button up one pixel when its content was changed
window.onload = function () {
    var button = document.getElementById("lightningTipGetInvoice");

    button.style.height = (button.clientHeight + 1) + "px";
    button.style.width = (button.clientWidth + 1) + "px";
};

// TODO: show invoice even if JavaScript is disabled
// TODO: fix scaling on phones
// TODO: show price in dollar?
function getInvoice() {
    if (!requestPending) {
        requestPending = true;

        var tipValue = document.getElementById("lightningTipAmount");

        if (tipValue.value !== "") {
            if (!isNaN(tipValue.value)) {
                var data = JSON.stringify({"Amount": parseInt(tipValue.value), "Message": document.getElementById("lightningTipMessage").innerText});

                var request = new XMLHttpRequest();

                request.onreadystatechange = function () {
                    if (request.readyState === 4) {
                        try {
                            var json = JSON.parse(request.responseText);

                            if (request.status === 200) {
                                console.log("Got invoice: " + json.Invoice);
                                console.log("Invoice expires in: " + json.Expiry);
                                console.log("Got rHash of invoice: " + json.RHash);

                                console.log("Starting listening for invoice to get settled");

                                listenInvoiceSettled(json.RHash, json.Picture);

                                invoice = json.Invoice;

                                // Update UI
                                var wrapper = document.getElementById("lightningTip");

                                wrapper.innerHTML = "<a>Your tip request</a>";
                                wrapper.innerHTML += "<input type='text' class='lightningTipInput' id='lightningTipInvoice' onclick='copyInvoiceToClipboard()' value='" + invoice + "' readonly>";
                                wrapper.innerHTML += "<div id='lightningTipQR'></div>";

                                wrapper.innerHTML += "<div id='lightningTipTools'>" +
                                    "<button class='lightningTipButton' id='lightningTipCopy' onclick='copyInvoiceToClipboard()'>Copy</button>" +
                                    "<button class='lightningTipButton' id='lightningTipOpen'>Open</button>" +
                                    "<a id='lightningTipExpiry'></a>" +
                                    "</div>";

                                starTimer(json.Expiry, document.getElementById("lightningTipExpiry"));

                                // Fixes bug which caused the content of #lightningTipTools to be visually outside of #lightningTip
                                document.getElementById("lightningTipTools").style.height = document.getElementById("lightningTipCopy").clientHeight + "px";

                                document.getElementById("lightningTipOpen").onclick = function () {
                                    location.href = "lightning:" + json.Invoice;
                                };

                                showQRCode();

                            } else {
                                showErrorMessage(json.Error);
                            }

                        } catch (exception) {
                            console.error(exception);

                            showErrorMessage("Failed to reach backend");
                        }

                        requestPending = false;
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

function listenInvoiceSettled(rHash, picture) {
    try {
        var eventSrc = new EventSource(requestUrl + "eventsource");

        eventSrc.onmessage = function (event) {
            if (event.data === rHash) {
                console.log("Invoice settled");

                eventSrc.close();

                showThankYouScreen(picture);
            }

        };

    } catch (e) {
        console.error(e);
        console.warn("Your browser does not support EventSource. Sending a request to the server every two second to check if the invoice settled");

        var interval = setInterval(function () {
            if (!requestPending) {
                requestPending = true;

                var request = new XMLHttpRequest();

                request.onreadystatechange = function () {
                    if (request.readyState === 4) {
                        if (request.status === 200) {
                            var json = JSON.parse(request.responseText);

                            if (json.Settled) {
                                console.log("Invoice settled");

                                clearInterval(interval);

                                showThankYouScreen(json.Picture);
                            }

                        }

                        requestPending = false;
                    }

                };

                request.open("POST", requestUrl + "invoicesettled", true);
                request.send(JSON.stringify({"RHash": rHash}));
            }

        }, 2000);

    }

}

function showThankYouScreen(picture) {
    var wrapper = document.getElementById("lightningTip");

    wrapper.innerHTML = "<p id=\"lightningTipLogo\">⚡</p>";
    wrapper.innerHTML += "<a id='lightningTipFinished'>Thank you for your tip!</a>";
    if (picture !== "") {
      wrapper.innerHTML += "<a href=" + picture + "><img height=150 src=" + picture + "></a>";
    }
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

    createQRCode();

    element.innerHTML = qrCode;

    var size = document.getElementById("lightningTipInvoice").clientWidth + "px";

    var qrElement = element.children[0];

    qrElement.style.height = size;
    qrElement.style.width = size;
}

function createQRCode() {
    var invoiceLength = invoice.length;

    // Just in case an invoice bigger than expected gets created
    var typeNumber = 26;

    for (var i = 0; i < qrCodeDataCapacities.length; i++) {
        var dataCapacity = qrCodeDataCapacities[i];

        if (invoiceLength < dataCapacity.capacity) {
            typeNumber = dataCapacity.typeNumber;

            break;
        }

    }

    console.log("Creating QR code with type number: " + typeNumber);

    var qr = qrcode(typeNumber, "L");

    qr.addData(invoice);
    qr.make();

    qrCode = qr.createImgTag(6, 6);
}

function copyInvoiceToClipboard() {
    var element = document.getElementById("lightningTipInvoice");

    element.select();

    document.execCommand('copy');

    console.log("Copied invoice to clipboard");
}

function showErrorMessage(message) {
    requestPending = false;

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

function divRestorePlaceholder(element) {
    // <br> and <div><br></div> mean that there is no user input
    if (element.innerHTML === "<br>" || element.innerHTML === "<div><br></div>") {
        element.innerHTML = "";
    }
}
