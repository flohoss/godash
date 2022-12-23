// webSocket
const WsType = { Weather: 0, System: 1 };
const wsUrl = window.location.origin.replace("http", "ws") + "/ws";
let timeOut = 1;
connect();

// weather elements
const weatherIcon = document.getElementById("weatherIcon");
const weatherTemp = document.getElementById("weatherTemp");
const weatherDescription = document.getElementById("weatherDescription");
const weatherHumidity = document.getElementById("weatherHumidity");
const weatherSunrise = document.getElementById("weatherSunrise");
const weatherSunset = document.getElementById("weatherSunset");

// system elements
const systemCpuPercentage = document.getElementById("systemCpuPercentage");
const systemRamPercentage = document.getElementById("systemRamPercentage");
const systemRamValue = document.getElementById("systemRamValue");
const systemDiskPercentage = document.getElementById("systemDiskPercentage");
const systemDiskValue = document.getElementById("systemDiskValue");
const systemUptimePercentage = document.getElementById("systemUptimePercentage");
const uptimeDays = document.getElementById("uptimeDays");
const uptimeHours = document.getElementById("uptimeHours");
const uptimeMinutes = document.getElementById("uptimeMinutes");
const uptimeSeconds = document.getElementById("uptimeSeconds");

function connect() {
  let ws = new WebSocket(wsUrl);
  ws.onopen = () => {
    console.log("WebSocket is open.");
    timeOut = 1;
  };
  ws.onmessage = (event) => handleMessage(JSON.parse(event.data));
  ws.onerror = () => ws.close();
  ws.onclose = () => {
    console.log("WebSocket is closed. Reconnect will be attempted in " + timeOut + " second.");
    setTimeout(() => connect(), timeOut * 1000);
    timeOut += 1;
  };
}

function handleMessage(parsed) {
  if (parsed.ws_type === WsType.Weather) replaceWeather(parsed.message);
  else if (parsed.ws_type === WsType.System) replaceSystem(parsed.message);
}

function replaceWeather(parsed) {
  weatherIcon.setAttribute("xlink:href", "#" + parsed.icon);
  weatherTemp.innerText = parsed.temp;
  weatherDescription.innerText = parsed.description;
  weatherHumidity.innerText = parsed.humidity + "%";
  weatherSunrise.innerText = parsed.sunrise;
  weatherSunset.innerText = parsed.sunset;
}

function replaceSystem(parsed) {
  systemCpuPercentage.style = "width:" + parsed.cpu + "%";
  systemRamPercentage.style = "width:" + parsed.ram.percentage + "%";
  systemRamValue.innerText = parsed.ram.value;
  systemDiskPercentage.style = "width:" + parsed.disk.percentage + "%";
  systemDiskValue.innerText = parsed.disk.value;
  systemUptimePercentage.style = "width:" + parsed.uptime.percentage + "%";
  uptimeDays.style = "--value:" + parsed.uptime.days;
  uptimeHours.style = "--value:" + parsed.uptime.hours;
  uptimeMinutes.style = "--value:" + parsed.uptime.minutes;
  uptimeSeconds.style = "--value:" + parsed.uptime.seconds;
}
