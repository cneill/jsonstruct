function clipboardCopy() {
    var output = document.querySelector("#output");
    output.select();
    output.setSelectionRange(0, 9999999);
    navigator.clipboard.writeText(output.value);
}


window.onload = function() {
    var copy = document.querySelector("#copy");

    copy.addEventListener("click", e => {
        clipboardCopy();
    });
}

document.body.addEventListener("htmx:beforeSwap", e => {
    if (e.detail.xhr.status >= 400) {
        e.detail.shouldSwap = false;
        e.detail.isError = true;
    }
});

document.body.addEventListener("htmx:beforeRequest", e => {
    let inputElem = e.detail.elt.querySelector(".input");
    let inputText = inputElem.value;

    try {
        JSON.parse(inputText);
    } catch (err) {
        if (inputText != "") {
            e.preventDefault();
        }
    }
});
