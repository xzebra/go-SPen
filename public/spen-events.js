var ws = new WebSocket("ws://192.168.100.9:8080/ws");

var dragging = false;

var connInfo = document.getElementById("conn_info");
var touchArea = document.getElementById("touch_area");

var socketCheck = setInterval(function() {
    if(ws.readyState == ws.OPEN) {
        connInfo.style.display = "none";
        touchArea.style.display = "block";
        ws.send("screen" + "," + window.innerWidth + "," + window.innerHeight)
        clearInterval(socketCheck);
    } else if(ws.readyState == ws.CLOSED) {
        connInfo.innerHTML = "Couldn't connect to the server";
        clearInterval(socketCheck);
    }
}, 500);

document.addEventListener("mousemove", function(e) {
    ws.send(e.clientX + "," + e.clientY);
});

document.addEventListener("touchmove", function(e) {
    let pos = e.targetTouches[0];

    ws.send(Math.floor(pos.clientX) + "," + Math.floor(pos.clientY));
});
