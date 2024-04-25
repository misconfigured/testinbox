function goBack() {
    var currentDomain = window.location.hostname;
    var referrerDomain = new URL(document.referrer).hostname;
    if (referrerDomain === currentDomain && window.history.length > 1) {
        window.history.back();
    } else {
        window.location.href = '/';
    }
}
