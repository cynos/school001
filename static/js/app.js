function notify(element, status, msg) {
    var template = `<div class="alert ${status}" role="alert">${msg}</div>`
    $(element).html(template)
    setTimeout(function(){
        $(element).html('')
    }, 2000)
}

function getParameters(paramName){
    var urlParams = new URLSearchParams(window.location.search);
    var getValue  = urlParams.get(paramName);
    return !getValue ? "" : getValue;

}

// get error ajax
$(document).ajaxError(function (event, jqxhr, settings, exception) {
    if (jqxhr.getResponseHeader("Requires-Auth") == "1") window.location.reload();
})