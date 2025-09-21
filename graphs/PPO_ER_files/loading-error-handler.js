/**
 * This file will be included in the index.html as early as possible to handle errors which occur before the primary
 * error handler is loaded.
 *
 * *****************************************************************************************************************
 * IMPORTANT: This file will not be babeled in the current build process. Therefore this needs to be written in ES5.
 * *****************************************************************************************************************
 */

;(function () {
  // try to increase the stacktrace limit
  try {
    Error.stackTraceLimit = 50
  } catch (e) {
    /* do nothing if it didn't work */
  }

  window.onerror = function (message, url, lineNumber, columnNumber, error) {
    var eventId = Math.random().toString(16).slice(2)
    var payload = {
      event_id: eventId,
      platform: 'javascript',
      message: message,
      sdk: {
        name: 'sentry.javascript.browser',
        packages: [
          {
            name: 'npm:@sentry/browser',
            version: '4.1.1'
          }
        ],
        version: '4.1.1'
      },
      release: '7.3.2',
      level: 'warning',
      request: {
        url: url,
        headers: {
          'User-Agent': window.navigator.userAgent
        }
      },
      exception: {
        mechanism: {
          handled: false,
          type: 'generic'
        },
        values: [
          {
            type: 'Error',
            value: 'LoadingError: ' + message
          }
        ]
      }
    }

    if (error) {
      payload.extra = {
        stack: error.stack
      }
    }

    var xhr = new XMLHttpRequest()
    xhr.open(
      'POST',
      'https://reporting.yworks.com/api/2/store/?sentry_key=d151434c21f143c78b0c5e75968212fc&sentry_version=7',
      true
    )
    xhr.setRequestHeader('Content-Type', 'application/json')
    xhr.send(JSON.stringify(payload))
  }
})()
