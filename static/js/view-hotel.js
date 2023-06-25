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
			/* Parse global message */
			var matches = data.GlobalMessage.match(/.*bojler (.) ochladil o (\d+°.).*/);
			if (matches) {
				var arrow = matches[1] == "A" ? '→' : '←';
				var change = matches[2];
				document.getElementById('middle-gauge').innerHTML = "<span>" + arrow + "</span><br>" + change;
			} else {
				document.getElementById('middle-gauge').innerHTML = "?";
			}
		} else {
			document.getElementById('middle-gauge').innerHTML = "";
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

////////////////////////////////////////////////////////////////////////////////
// Thermometer /////////////////////////////////////////////////////////////////

// transforms value to 0..1 by min and max
function calcValue(value, min, max) {
	var valueCrop = Math.max(Math.min(value, max), min);
	return (valueCrop - min) / (max - min);
}

function convertToRGB (hex) {
	var color = [];
	hex = (hex.charAt(0) == '#') ? hex.substring(1, 7) : hex;
	color[0] = parseInt(hex.substring (0, 2), 16);
	color[1] = parseInt(hex.substring (2, 4), 16);
	color[2] = parseInt(hex.substring (4, 6), 16);
	return color;
}

function convertToHex (rgb) {
	return '#' + Math.round(rgb[0]).toString(16) + Math.round(rgb[1]).toString(16) + Math.round(rgb[2]).toString(16);
}

const temperatureConfig = {
	minTemp: -100,
	maxTemp: 200,
	unit: "°F",
	colorStart: convertToRGB("#e44427"),
	colorEnd: convertToRGB("#3dcadf"),
};

function temperatureAlpha(value) {
	return (value - temperatureConfig.minTemp) / (temperatureConfig.maxTemp - temperatureConfig.minTemp)
}

function setState(name, prevValue, value) {
	var alpha = temperatureAlpha(value);
	var color = [];
	color[0] = temperatureConfig.colorStart[0] * alpha + (1 - alpha) * temperatureConfig.colorEnd[0];
	color[1] = temperatureConfig.colorStart[1] * alpha + (1 - alpha) * temperatureConfig.colorEnd[1];
	color[2] = temperatureConfig.colorStart[2] * alpha + (1 - alpha) * temperatureConfig.colorEnd[2];

	var temperature = document.getElementById("temperature" + name);
	temperature.style.height = alpha * 100 + "%";
	temperature.dataset.value = value + temperatureConfig.unit;

	var bojler = document.getElementById("bojler" + name);
	var hexColor = convertToHex(color);
	bojler.style.backgroundColor = hexColor;

	document.documentElement.style.setProperty('--very-transparent' + name, hexColor + "66");
	document.documentElement.style.setProperty('--medium-transparent' + name, hexColor + "cc");

	if (prevValue < value) {
		document.getElementById("water" + name).style.opacity = 0;

		document.getElementById("fire" + name).style.opacity = 1;
		setTimeout(function() {
			document.getElementById("fire" + name).style.opacity = 0;
		}, 7_000);

		new Audio('static/sounds/fire.mp3').play();

	} else if (prevValue > value) {
		document.getElementById("water" + name).style.opacity = 1;
		setTimeout(function() {
			document.getElementById("water" + name).style.opacity = 0;
		}, 7_000);

		new Audio('static/sounds/shower.mp3').play();

		document.getElementById("fire" + name).style.opacity = 0;
	}

	var els = bojler.getElementsByClassName("bubble");
	if (value <= 32) {
		Array.from(els).forEach(element => {
			element.classList.add("freeze");
		});
	} else {
		Array.from(els).forEach(element => {
			element.classList.remove("freeze");
		});
	}
}
