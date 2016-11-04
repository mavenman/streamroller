function generateEntry(id, source, user, message, time) {
  return '<li id="' + id + '" class="message left appeared ' + source + '"> \
    <div class="avatar"><img src="/img/' + source + '.png" /></div> \
    <div class="text_wrapper"> \
      <div class="text"><span class="user">' + user + '</span>: ' + message + '</div> \
      <div class="timestamp">' + time + '</div> \
      </div> \
  </li>';
}

const ws = new WebSocket('ws://' + window.location.host + '/socket');
ws.onopen = function() {
  $("#connection").attr('class', 'connected');
  $("#connection").text('Connected');
};
ws.onclose = function() {
  $("#connection").attr('class', 'disconnected');
  $("#connection").text('Disconnected');
};
ws.onmessage = function(msg) {
  var data = JSON.parse(msg.data);
  var date = moment();
  var id = 'msg_' + date.unix();
  $('#messages').prepend(generateEntry(id, data.source, data.user, data.message, date.format('hh:mma')))
  // TODO Revist this.
  // setTimeout(function() {
  //   $('#' + id).addClass('appeared')
  // }, 10)
};
