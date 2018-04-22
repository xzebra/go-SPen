var spenWS = new WebSocket("ws://192.168.100.9:8080/spen");
var fingerWS = new WebSocket("ws://192.168.100.9:8080/finger");

var dragging = false;

var connInfo = document.getElementById("conn_info");
var touchArea = document.getElementById("touch_area");

var socketCheck = setInterval(function() {
    if(spenWS.readyState == spenWS.OPEN && fingerWS.readyState == fingerWS.OPEN) {
        connInfo.style.display = "none";
        touchArea.style.display = "block";
        spenWS.send("screen" + "," + window.innerWidth + "," + window.innerHeight);
        clearInterval(socketCheck);
    } else if(spenWS.readyState == spenWS.CLOSED || fingerWS.readyState == fingerWS.CLOSED) {
        connInfo.innerHTML = "Couldn't connect to the server";
        clearInterval(socketCheck);
    }
}, 500);

document.addEventListener("mousemove", function(e) {
    // SPen touch radius is always 0 while finger touches
    // are greater than 0
    if(e.radiusX === 0) { // Using the spen   
        spenWS.send(e.clientX + "," + e.clientY);
    } else { // Touching with the finger        
        fingerWS.send(e.clientX + "," + e.clientY);
    }
});

document.addEventListener("touchmove", function(e) {
    let touch = e.targetTouches[0];
    // SPen touch radius is always 0 while finger touches
    // are greater than 0
    if(touch.radiusX === 0) { // Using the spen
        spenWS.send(Math.floor(touch.clientX) + "," + Math.floor(touch.clientY));
    } else { // Touching with the finger
        fingerWS.send(Math.floor(touch.clientX) + "," + Math.floor(touch.clientY));
    }
});
