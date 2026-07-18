(function () {
  function setText(id, text) {
    var el = document.getElementById(id);
    if (el) {
      el.textContent = text;
      if (el.title !== undefined) el.title = text;
    }
  }

  function setWidth(id, pct) {
    var el = document.getElementById(id);
    if (el) el.style.width = pct + '%';
  }

  function setIconClass(id, cls) {
    var el = document.getElementById(id);
    if (el) {
      el.className = el.className.replace(/icon-\[[^\]]+\]/g, '').trim() + ' ' + cls;
    }
  }

  function handleSystem(name, data) {
    var d = JSON.parse(data);
    setText('value-' + name, d.value);
    setWidth('bar-' + name, d.percentage);
  }

  function handleWeather(name, data) {
    if (name === 'current') {
      var d = JSON.parse(data);
      setIconClass('weather-icon-0', d.icon);
      setText('day-name-0', d.name);
      setText('temp', d.more.current_temperature);
      setText('max-temp-0', d.temperature_max);
      setText('min-temp-0', d.temperature_min);
      setText('humidity', d.more.humidity);
      setText('sunrise', d.more.sunrise);
      setText('apparent', d.more.apparent_temperature);
      setText('sunset', d.more.sunset);
      return;
    }
    if (name === 'forecast') {
      var days = JSON.parse(data);
      for (var i = 1; i < days.length; i++) {
        var d = days[i];
        setIconClass('weather-icon-' + i, d.icon);
        setText('day-name-' + i, d.name);
        setText('max-temp-' + i, d.temperature_max);
        setText('min-temp-' + i, d.temperature_min);
      }
    }
  }

  var handlers = { system: handleSystem, weather: handleWeather };

  function connect(url, stream) {
    var es = new EventSource(url);
    es.onmessage = function (e) {
      var handler = handlers[stream];
      if (handler) {
        try { handler('message', e.data); } catch (err) {}
      }
    };
    es.addEventListener('cpu', function (e) { handleSystem('cpu', e.data); });
    es.addEventListener('ram', function (e) { handleSystem('ram', e.data); });
    es.addEventListener('disk', function (e) { handleSystem('disk', e.data); });
    es.addEventListener('current', function (e) { handleWeather('current', e.data); });
    es.addEventListener('forecast', function (e) { handleWeather('forecast', e.data); });
  }

  function init() {
    document.querySelectorAll('[data-sse]').forEach(function (r) {
      var url = r.getAttribute('data-sse');
      var stream = url.split('stream=')[1];
      connect(url, stream);
    });
  }

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
  } else {
    init();
  }
})();