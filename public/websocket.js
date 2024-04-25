var protocolPrefix = (window.location.protocol === 'https:') ? 'wss:' : 'ws:';
var socket = new WebSocket(protocolPrefix + '//' + window.location.host + '/ws');

document.addEventListener('DOMContentLoaded', function () {
    socket.onopen = function() {
        console.log("WebSocket connection established");
    };

    socket.onmessage = function(event) {
        if (event.data === "heartbeat") {
            console.log("Heartbeat received");
            return;
        }
        try {
            var email = JSON.parse(event.data);
            updateEmailList(email);
        } catch (e) {
            console.error("Error parsing email data:", e);
        }

    };

    socket.onerror = function(error) {
        console.log("WebSocket Error: " + error);
    };
});

function updateEmailList(email) {
    var emailTableBody = document.getElementById("emailList").getElementsByTagName("tbody")[0];
    if (emailTableBody.rows.length >= 25) {
        emailTableBody.deleteRow(emailTableBody.rows.length - 1);
    }
    
    var formattedDate = new Date(email.receivedAt).toLocaleString(); 
    console.error("formattedDate" + formattedDate);

    var newRow = emailTableBody.insertRow(0);
    newRow.style.cursor = "pointer";
    newRow.onclick = function() { window.location = '/inbox/' + email.id; };

    var cellRecipient = newRow.insertCell(0);
    cellRecipient.className = "px-6 py-4 whitespace-nowrap";
    cellRecipient.innerHTML = `<div class="text-sm text-gray-900">${email.recipient}</div>`;

    var cellSender = newRow.insertCell(1);
    cellSender.className = "px-6 py-4 whitespace-nowrap";
    cellSender.innerHTML = `<div class="text-sm text-gray-900">${email.sender}</div>`;

    var cellSubject = newRow.insertCell(2);
    cellSubject.className = "px-6 py-4 whitespace-nowrap";
    cellSubject.innerHTML = `<div class="text-sm text-gray-900">${email.subject}</div>`;

    var cellReceivedAt = newRow.insertCell(3);
    cellReceivedAt.className = "px-6 py-4 whitespace-nowrap";
    cellReceivedAt.innerHTML = `<div class="text-sm text-gray-500">${email.receivedAt}</div>`;

}
