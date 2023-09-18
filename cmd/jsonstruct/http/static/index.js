function clipboardCopy() {
    let outputElem = document.querySelector("#output");

    if (outputElem) {
        navigator.clipboard.writeText(outputElem.innerText);
    }
}


window.onload = function() {
    var copy = document.querySelector("#copy");

    copy.addEventListener("click", e => {
        clipboardCopy();
    });
}

document.body.addEventListener("htmx:beforeRequest", e => {
    let inputElem = e.detail.elt.querySelector(".input");
    if (!inputElem) {
        return;
    }

    try {
        JSON.parse(inputElem.value);
    } catch (err) {
        if (inputElem.value != "") {
            e.preventDefault();
        }
    }
});

document.body.addEventListener("htmx:beforeSwap", e => {
    if (e.detail.xhr.status >= 400) {
        e.detail.shouldSwap = false;
        e.detail.isError = true;
    }
});

document.body.addEventListener("htmx:afterSwap", e => {
    let codeElem = e.detail.elt.querySelector(".output");

    if (codeElem) {
        Prism.highlightElement(codeElem);
    }
});
