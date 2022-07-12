var state = {
	"A": 0,
	"B": 0,
}
var hash = "";

function reloadStatus() {
	getJSON("/view/status", function(err, data) {
		if (err !== null) {
			console.log("Error: " + err);
			return;
		}
		document.getElementById('round').innerHTML = "Kolo " + data.RoundNumber;
		if (data.GlobalMessage) {
			document.getElementById('globalMessage').innerHTML = data.GlobalMessage;
			document.getElementById('globalMessage').style.opacity = 1;
		} else {
			document.getElementById('globalMessage').style.opacity = 0;
		}

		var countdownObj = document.getElementById('countdown');
		if (data.CountdownActive) {
			countdownObj.style.opacity = 1;
			startTimer(data.CountdownSeconds, countdownObj);
		} else {
			countdownObj.style.opacity = 0;
		}

		state = data.GlobalState;
		hash = data.Hash;

		setGauge("A", state["A"]);
		setReef("A", state["A"]);
		setGauge("B", state["B"]);
		setReef("B", state["B"]);

		console.log(data);
	})
}

function checkHash() {
	var newHash = httpGet("view/getHash");
	if (newHash != hash) {
		reloadStatus()
	}
}
setInterval(checkHash, 1000);
checkHash();

////////////////////////////////////////////////////////////////////////////////
// Gauge ///////////////////////////////////////////////////////////////////////

// min/max for gauge needle
const needleMin = -100;
const needleMax = 200;
var gaugeState = {
	"A": 0,
	"B": 0,
}

// transforms value to 0..1 by min and max
function calcValue(value, min, max) {
	var valueCrop = Math.max(Math.min(value, max), min);
	return (valueCrop - min) / (max - min);
}

function setGauge(name, value) {
	var needle = document.getElementById("needle" + name);
	var counter = document.getElementById("counter" + name);

	var duration = 3000;
	var start = gaugeState["A"];

	var needlePos = (1 - calcValue(value, needleMin, needleMax)) * 180
	// console.log(needlePos);
	needle.style.transform = "rotate(" + needlePos + "deg)";

	$({countValue: start}).animate(
		{countValue: value},
		{	duration: duration, /* time for animation in milliseconds */
			step: function (value) { /* fired every "frame" */
				value = Math.floor(value);
				gaugeState[name] = value;
				counter.innerHTML = value;
			}
		}
	);
}

////////////////////////////////////////////////////////////////////////////////
// Reef ////////////////////////////////////////////////////////////////////////

// min/max for visualisation
const reefMin = -100;
const reefMax = 150;

function setReef(name, value) {
	var reefImgs = $("#reef" + name + " img");
	var reefBubbles = $("#reef" + name + " .bubble");

	var v = calcValue(value, reefMin, reefMax);
	var saturate = 10 + 90*v;
	var hueRotate = 90 * (1-v);
	var brightness = 50 + 50*v;

	reefImgs.css("filter", "saturate(" + saturate + "%) hue-rotate(-" + hueRotate + "deg) brightness(" + brightness + "%)");
	reefBubbles.css("border-color", "hsl(120, 100%, " + (25 + v*75) + "%)");
}

////////////////////////////////////////////////////////////////////////////////
// Fish ////////////////////////////////////////////////////////////////////////

function randInt(from, to) {
	return from + Math.floor(Math.random() * (to - from));
}

function addFish(name, alive) {
	const e = document.createElement('div');
	var className = "fish";

	if (Math.random() > 0.5) {
		className = className + " right";
	}

	if (Math.random() > 0.5) {
		className = className + " large";
	}
	e.className = className;

	e.style = "top: " + randInt(0, 80) + "%";

	var fish = "fish" + randInt(1, 3) + ".png";
	if (!alive) {
		fish = "fish_dead.png";
	}
	e.innerHTML = '<img src="/static/images/' + fish + '">';

	// insert fish... and autoremove it after 30s to prevent infinite number of divs
	$(e).insertBefore("#reef" + name + " .front");
	setTimeout(function() {
		e.remove();
	}, 30000)
}

function spawnLoop() {
	for (const name of ["A", "B"]) {
		console.log(name);
		var s = state[name];
		var alive = true;
		if (s <= 0) {
			alive = false;
		} else if (s < 100 && randInt(0, 100) > s) {
			alive = false;
		}
		addFish(name, alive);
	}
}

setInterval(spawnLoop, 5000);
