function resizeInput(element) {
    element.style.height = "auto";
    element.style.height = (element.scrollHeight) + "px";
}


function getInvoice() {
    var tipValue = document.getElementById("lightningTipAmount");

    if (tipValue.value !== "") {
        if (!isNaN(tipValue.value)) {
            var data = JSON.stringify({"Amount": parseInt(tipValue.value), "Message": document.getElementById("lightningTipMessage").value});

            var request = new XMLHttpRequest();

            request.onreadystatechange = function () {
                if (request.readyState === 4) {
                    var json = JSON.parse(request.responseText);

                    if (request.status === 200) {
                        console.log("Got invoice: " + json.Invoice);
                        console.log("Invoice expires in: " + json.Expiry);

                        var wrapper = document.getElementById("lightningTip");

                        // TODO: timer until expiry
                        wrapper.innerHTML = "<a>Your invoice</a>";
                        wrapper.innerHTML += "<textarea class='lightningTipInput' id='lightningTipInvoice' readonly>" + json.Invoice + "</textarea>";

                        resizeInput(document.getElementById("lightningTipInvoice"))

                    } else {
                        showErrorMessage(json.Error);
                    }

                }

            };

            // TODO: proper url handling window.location.protocol + window.location.hostname + ":8081/getinvoice"
            request.open("POST", "http://localhost:8081/getinvoice", true);
            request.send(data);

            var button = document.getElementById("lightningTipGetInvoice");

            button.style.height = button.clientHeight + "px";
            button.style.width = button.clientWidth + "px";

            button.innerHTML = "<div class='spinner'></div>";

        } else {
            showErrorMessage("Tip amount must be a number");
        }

    } else {
        showErrorMessage("No tip amount set");
    }

}

function showErrorMessage(message) {
    console.error(message);

    var error = document.getElementById("lightningTipError");

    error.parentElement.style.marginTop = "0.5em";

    error.innerHTML = message;
}