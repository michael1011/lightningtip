function resizeInput(element) {
    element.style.height = "auto";
    element.style.height = (element.scrollHeight) + "px";
}

// To prohibit multiple requests at the same time
var running = false;

var invoice;
var qrCode;

var defaultGetInvoice;

// TODO: listen to eventsource and show tank you when invoice settled
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

                                invoice = json.Invoice;

                                var wrapper = document.getElementById("lightningTip");

                                wrapper.innerHTML = "<a>Your invoice</a>";
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
                            showErrorMessage("Failed to reach backend");
                        }

                    }

                };

                // TODO: proper url handling window.location.protocol + window.location.hostname + ":8081/getinvoice"
                request.open("POST", "http://localhost:8081/getinvoice", true);
                request.send(data);

                var button = document.getElementById("lightningTipGetInvoice");

                button.style.height = button.clientHeight + "px";
                button.style.width = button.clientWidth + "px";

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

function starTimer(duration, element) {
    showTimer(duration, element);

    var interval = setInterval(function () {
        if (duration > 0) {
            duration--;

            showTimer(duration, element);

        } else {
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