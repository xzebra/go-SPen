var spenWS = new WebSocket("ws://192.168.100.9:8080/spen");
var fingerWS = new WebSocket("ws://192.168.100.9:8080/finger");

var connInfo = document.getElementById("conn_info");
var touchArea = document.getElementById("touch_area");

var ongoingTouches = new Array;
var canvas = document.getElementById("canvas");

canvas.width = window.innerWidth;
canvas.height = window.innerHeight;

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


function ongoingTouchIndexById(idToFind) {
  for (var i=0; i<ongoingTouches.length; i++) {
    var id = ongoingTouches[i].identifier;
    
    if (id == idToFind) {
      return i;
    }
  }
  return -1; // not found
}

document.addEventListener("mousemove", function(e) {
    spenWS.send(e.clientX + "," + e.clientY);
});

function handleStart(evt) {
  evt.preventDefault();
  var touches = evt.changedTouches;

  spenWS.send("pressing");
        
  for (var i=0; i<touches.length; i++) {
    ongoingTouches.push(touches[i]);
    spenWS.send(touches[i].clientX + "," + touches[i].clientY);
  }
}

function handleMove(evt) {
    evt.preventDefault();
    var touches = evt.changedTouches;
  
    for (var i=0; i<touches.length; i++) {
      var idx = ongoingTouchIndexById(touches[i].identifier);

      spenWS.send(ongoingTouches[idx].clientX + "," + ongoingTouches[idx].clientY);
      spenWS.send(touches[i].clientX + "," + touches[i].clientY);
      ongoingTouches.splice(idx, 1, touches[i]);  // swap in the new touch record
    }
  }

function handleEnd(evt) {
    evt.preventDefault();
    var touches = evt.changedTouches;

    spenWS.send("stoppressing");

    for (var i=0; i<touches.length; i++) {
        var idx = ongoingTouchIndexById(touches[i].identifier);

        spenWS.send(ongoingTouches[idx].clientX + "," + ongoingTouches[idx].clientY);
        spenWS.send(touches[i].clientX + "," + touches[i].clientY);
        ongoingTouches.splice(i, 1);  // remove it; we're done
    }
}

function handleCancel(evt) {
    evt.preventDefault();
    var touches = evt.changedTouches;
    
    spenWS.send("stoppressing");
  
    for (var i=0; i<touches.length; i++) {
      ongoingTouches.splice(i, 1);  // remove it; we're done
    }
}

function startup() {
    canvas.addEventListener("touchstart", handleStart, false);
    canvas.addEventListener("touchend", handleEnd, false);
    canvas.addEventListener("touchcancel", handleCancel, false);
    canvas.addEventListener("touchleave", handleEnd, false);
    canvas.addEventListener("touchmove", handleMove, false);
  }
