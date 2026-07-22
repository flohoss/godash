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

  function handleSystemValue(metric, data) {
    try { setText('value-' + metric, JSON.parse(data)); } catch (e) {}
  }

  function handleSystemPercentage(metric, data) {
    try { setWidth('bar-' + metric, JSON.parse(data)); } catch (e) {}
  }

  function updateHourCard(i, h) {
    setText('hour-time-' + i, h.time);
    setIconClass('hour-icon-' + i, h.icon);
    setText('hour-temp-' + i, h.temperature);
    setText('hour-wind-' + i, h.wind_speed || '');
    setText('hour-precip-' + i, h.precip_prob || '');
  }

  function rebuildHourly(container, hours) {
    container.innerHTML = '';
    for (var i = 0; i < hours.length; i++) {
      var h = hours[i];
      var card = document.createElement('div');
      var responsive = i >= 6 ? 'hidden 2xl:flex' : i >= 4 ? 'hidden xl:flex' : i >= 2 ? 'hidden lg:flex' : '';
      card.className = 'flex w-16 shrink-0 flex-col items-end gap-1 px-2 py-1.5 ' + responsive;
      card.innerHTML =
        '<div class="flex flex-col items-end gap-0.5">' +
          '<div id="hour-time-' + i + '" class="text-secondary text-xs">' + h.time + '</div>' +
          '<div id="hour-icon-' + i + '" class="icon size-8 shrink-0 ' + h.icon + '"></div>' +
          '<div id="hour-temp-' + i + '" class="font-semibold text-sm">' + h.temperature + '</div>' +
        '</div>' +
        '<div class="flex flex-col items-end gap-0.5 text-secondary text-xs w-full">' +
          '<div class="flex items-center justify-center gap-1 whitespace-nowrap">' +
            '<span class="icon-[carbon--windy] size-3.5 shrink-0"></span>' +
            '<div id="hour-wind-' + i + '" class="text-secondary whitespace-nowrap">' + (h.wind_speed || '') + '</div>' +
          '</div>' +
          '<div class="flex items-center justify-center gap-1">' +
            '<span class="icon-[carbon--rain] size-3.5 shrink-0"></span>' +
            '<div id="hour-precip-' + i + '" class="text-secondary">' + (h.precip_prob || '') + '</div>' +
          '</div>' +
        '</div>';
      container.appendChild(card);
    }
  }

  function handleWeather(name, data) {
    try {
      if (name === 'current') {
        var d = JSON.parse(data);
        setIconClass('weather-icon-0', d.icon);
        setText('day-name-0', d.name);
        setText('temp', d.more.current_temperature);
        setText('max-temp-0', d.temperature_max);
        setText('min-temp-0', d.temperature_min);
        setText('humidity', d.more.humidity);
        setText('wind', (d.more.wind_speed || ''));
        setText('sunrise', d.more.sunrise);
        setText('apparent', d.more.apparent_temperature);
        setText('sunset', d.more.sunset);
        return;
      }
      if (name === 'hourly') {
        var hours = JSON.parse(data);
        var container = document.getElementById('hourly');
        if (!container) return;
        var existing = container.children;
        if (existing.length !== hours.length) {
          rebuildHourly(container, hours);
        } else {
          for (var i = 0; i < hours.length; i++) {
            updateHourCard(i, hours[i]);
          }
        }
      }
    } catch (e) {}
  }

  function connect(url) {
    var es = new EventSource(url);
    es.addEventListener('cpu-value', function (e) { handleSystemValue('cpu', e.data); });
    es.addEventListener('cpu-percentage', function (e) { handleSystemPercentage('cpu', e.data); });
    es.addEventListener('ram-value', function (e) { handleSystemValue('ram', e.data); });
    es.addEventListener('ram-percentage', function (e) { handleSystemPercentage('ram', e.data); });
    es.addEventListener('disk-value', function (e) { handleSystemValue('disk', e.data); });
    es.addEventListener('disk-percentage', function (e) { handleSystemPercentage('disk', e.data); });
    es.addEventListener('current', function (e) { handleWeather('current', e.data); });
    es.addEventListener('hourly', function (e) { handleWeather('hourly', e.data); });
  }

  function init() {
    document.querySelectorAll('[data-sse]').forEach(function (r) {
      connect(r.getAttribute('data-sse'));
    });
  }

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
  } else {
    init();
  }
})();