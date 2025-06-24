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

		if (data.PollutionTo) {
			var arrow = data.PollutionTo == "A" ? '→' : '←';
			// document.getElementById('middle-gauge').innerHTML = "<span>" + arrow + "</span><br>" + data.PollutionAmount;
		} else {
			// document.getElementById('middle-gauge').innerHTML = "?";
		}

		var gm = document.getElementById('global-message')
		if (data.GlobalMessage) {
			gm.style.opacity = 1.0;
			gm.innerHTML = data.GlobalMessage;
		} else {
			gm.style.opacity = 0.0;
		}

		var countdownObj = document.getElementById('countdown');
		if (data.CountdownActive) {
			countdownObj.style.opacity = 1;
			startTimer(data.CountdownSeconds, countdownObj);
		} else {
			countdownObj.style.opacity = 0;
		}

		prevState = state
		state = data.GlobalState;
		hash = data.Hash;

		setState("A", prevState["A"], state["A"]);
		setState("B", prevState["B"], state["B"]);

		// console.log(data);
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

// min/max for visualisation
const cityMin = -60;
const cityMax = 150;

// transforms value to 0..1 by min and max
function calcValue(value, min, max) {
	var valueCrop = Math.max(Math.min(value, max), min);
	return (valueCrop - min) / (max - min);
}

var saturationMin = 0.25;
var saturationMax = 1.75;
var brightnessMin = 0.75;
var brightnessMax = 1.25;

function setState(name, prevValue, value) {
	var state = document.getElementById("state" + name);
	var img = document.getElementById("city" + name);

	var saturation = saturationMin + calcValue(value, cityMin, cityMax) * (saturationMax - saturationMin);
	var brightness = brightnessMin + calcValue(value, cityMin, cityMax) * (brightnessMax - brightnessMin);
	img.style.filter = `saturate(${saturation}) brightness(${brightness})`;

	var duration = 3000; /* time for animation in milliseconds */

	$({countValue: prevValue}).animate(
		{countValue: value},
		{	duration: duration,
			step: function (value) { /* fired every "frame" */
				value = Math.floor(value);
				state.innerHTML = value;
			}
		}
	);
}
