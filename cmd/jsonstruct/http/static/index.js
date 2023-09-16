function clipboardCopy() {
    var output = document.querySelector("#output");
    output.select();
    output.setSelectionRange(0, 9999999);
    console.log(output.value);
    navigator.clipboard.writeText(output.value);
}


window.onload = function() {
    var copy = document.querySelector("#copy");

    copy.addEventListener("click", e => {
        clipboardCopy();
    });
}
