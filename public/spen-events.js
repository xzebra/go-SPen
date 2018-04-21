var dragging = false;
var outX = document.getElementById("outputX");
var outY = document.getElementById("outputY");
var isClicking = document.getElementById("isClicking");

document.addEventListener("mousemove", function(e) {
    outX.innerHTML = e.clientX;
    outY.innerHTML = e.clientY;
});

document.addEventListener("touchmove", function(e) {
    let pos = e.targetTouches[0];

    outX.innerHTML = Math.floor(pos.clientX);
    outY.innerHTML = Math.floor(pos.clientY);
});