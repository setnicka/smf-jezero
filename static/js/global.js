function httpGet(theUrl) {
	var xmlHttp = new XMLHttpRequest();
	xmlHttp.open("GET", theUrl, false); // false for synchronous request
	xmlHttp.send(null);
	return xmlHttp.responseText;
}

function getJSON(url, callback) {
	var xhr = new XMLHttpRequest();
	xhr.open('GET', url, true);
	xhr.responseType = 'json';
	xhr.onload = function() {
		var status = xhr.status;
		if (status === 200) {
			callback(null, xhr.response);
		} else {
			callback(status, xhr.response);
		}
	};
	xhr.send();
};

// source: https://stackoverflow.com/a/20618517
function startTimer(duration, display, callback=null) {
	var start = Date.now(),
	    diff,
	    minutes,
	    seconds;
	var timer;
	function tick() {
		// get the number of seconds that have elapsed since
		// startTimer() was called
		diff = duration - (((Date.now() - start) / 1000) | 0);


		if (diff < 0) {
			diff = 0;
		}

		// does the same job as parseInt truncates the float
		minutes = (diff / 60) | 0;
		seconds = (diff % 60) | 0;

		minutes = minutes < 10 ? "0" + minutes : minutes;
		seconds = seconds < 10 ? "0" + seconds : seconds;

		display.textContent = minutes + ":" + seconds;

		if (diff <= 0) {
			clearInterval(timer);
			if (callback) {
				callback();
			}
		}
	};
	// we don't want to wait a full second before the timer starts
	tick();
	timer = setInterval(tick, 1000);
}
