var spenWS = new WebSocket("ws://192.168.100.9:8080/spen");
var fingerWS = new WebSocket("ws://192.168.100.9:8080/finger");

var connInfo = document.getElementById("conn_info");
var touchArea = document.getElementById("touch_area");

var ongoingTouches = new Array;
var canvas = document.getElementById("canvas");

canvas.width = window.innerWidth;
canvas.height = window.innerHeight;

var fingers = new Array(2);

var socketCheck = setInterval(function () {
    if (spenWS.readyState == spenWS.OPEN && fingerWS.readyState == fingerWS.OPEN) {
        connInfo.style.display = "none";
        touchArea.style.display = "block";
        spenWS.send("screen" + "," + window.innerWidth + "," + window.innerHeight);
        clearInterval(socketCheck);
    } else if (spenWS.readyState == spenWS.CLOSED || fingerWS.readyState == fingerWS.CLOSED) {
        connInfo.innerHTML = "Couldn't connect to the server";
        clearInterval(socketCheck);
    }
}, 500);


function ongoingTouchIndexById(idToFind) {
    for (var i = 0; i < ongoingTouches.length; i++) {
        var id = ongoingTouches[i].identifier;

        if (id == idToFind) {
            return i;
        }
    }
    return -1; // not found
}

document.addEventListener("mousemove", function (e) {
    spenWS.send(e.clientX + "," + e.clientY);
});

function handleStart(evt) {
    evt.preventDefault();
    var touches = evt.changedTouches;

    for (var i = 0; i < touches.length; i++) {
        ongoingTouches.push(touches[i]);
        if (touches[i].radiusX === 0) { //using SPen
            spenWS.send("pressing");
            spenWS.send(touches[i].clientX + "," + touches[i].clientY);
        } else { // finger touches
            // store the finger touches and its starting positions
            let id = touches[i].identifier;
            if(id < 2) {
                if(fingers[id] === null) {
                    fingers[id] = {
                        'xDown': touches[i].clientX,
                        'yDown': touches[i].clientY,
                        'swipe': null
                    };
                }
            } else {
                // we can break the loop as if it is a finger touching
                // we can't be using the spen
                break;
            }
        }
    }
}

function handleMove(evt) {
    evt.preventDefault();
    var touches = evt.changedTouches;

    for (var i = 0; i < touches.length; i++) {
        var idx = ongoingTouchIndexById(touches[i].identifier);

        if (touches[i].radiusX === 0) { // using SPen
            spenWS.send(ongoingTouches[idx].clientX + "," + ongoingTouches[idx].clientY);
            spenWS.send(touches[i].clientX + "," + touches[i].clientY);
        } else { // finger has moved
            let id = touches[i].identifier;
            let xDiff = fingers[id].xDown - touches[i].clientX;
            let yDiff = fingers[id].yDown - touches[i].clientY;

            // We have to invert it because I use the phone in landscape
            // without rotating it (so I don't have the chrome bar wasting
            // that much space)
            if(Math.abs(xDiff) > Math.abs(yDiff)) { // horizontal -> vertical
                if(xDiff > 0) { // down swipe
                    fingerWS.send("swipe," + fingers[id].xDown + "," + fingers[id].yDown + ",3," + id);
                } else if (xDiff < 0){ //up swipe
                    fingerWS.send("swipe," + fingers[id].xDown + "," + fingers[id].yDown + ",1," + id);
                }           
            } else if (Math.abs(xDiff) < Math.abs(yDiff)){ // vertical -> horizontal
                if(yDiff > 0) { // left swipe
                    fingerWS.send("swipe," + fingers[id].xDown + "," + fingers[id].yDown + ",4," + id);
                } else if (yDiff < 0) { // right swipe
                    fingerWS.send("swipe," + fingers[id].xDown + "," + fingers[id].yDown + ",2," + id);
                }
            }
        }
        ongoingTouches.splice(idx, 1, touches[i]);  // swap in the new touch record
    }
}

function handleEnd(evt) {
    evt.preventDefault();
    var touches = evt.changedTouches;

    for (var i = 0; i < touches.length; i++) {
        var idx = ongoingTouchIndexById(touches[i].identifier);

        if (touches[i].radiusX === 0) { //using SPen
            spenWS.send(ongoingTouches[idx].clientX + "," + ongoingTouches[idx].clientY);
            spenWS.send(touches[i].clientX + "," + touches[i].clientY);
            spenWS.send("stoppressing");
        } else { //Finger
            //fingerWS.send(ongoingTouches[idx].clientX + "," + ongoingTouches[idx].clientY);
            //fingerWS.send(touches[i].clientX + "," + touches[i].clientY);
            //fingerWS.send("stoppressing");
            let id = touches[i].identifier;
            if(id < 2) {
                fingerWS.send("stop," + id);
                fingers[id] = null;
            }
        }
        ongoingTouches.splice(i, 1);  // remove it; we're done
    }
}

function handleCancel(evt) {
    evt.preventDefault();
    var touches = evt.changedTouches;

    for (var i = 0; i < touches.length; i++) {
        ongoingTouches.splice(i, 1);  // remove it; we're done
        if (touches[i].radiusX === 0) { //using SPen
            spenWS.send("stoppressing");
        } else { //Finger
            let id = touches[i].identifier;
            if(id < 2) {
                fingerWS.send("stop," + id);
                fingers[id] = null;
            }
        }
    }
}

function startup() {
    canvas.addEventListener("touchstart", handleStart, false);
    canvas.addEventListener("touchend", handleEnd, false);
    canvas.addEventListener("touchcancel", handleCancel, false);
    canvas.addEventListener("touchleave", handleEnd, false);
    canvas.addEventListener("touchmove", handleMove, false);
}
