;(function () {
  var domain = '.yworks.com'
  if (location.hostname.indexOf(domain) === location.hostname.length - domain.length) {
    function initMatomo() {
      var pageTracker = window.Matomo.getTracker('//trk.yworks.com/matomo.php', '1')
      pageTracker.disableCookies()
      pageTracker.trackPageView()
      pageTracker.enableLinkTracking()
      window.matomoPageTracker = pageTracker

      var appTracker = window.Matomo.getTracker('//trk.yworks.com/matomo.php', '3')
      appTracker.disableCookies()
      appTracker.trackPageView()
      appTracker.enableLinkTracking()
      appTracker.trackEvent('yEdLive2017', 'Status', 'Loading')
      window.matomoAppTracker = appTracker
    }

    if (window.Matomo) {
      initMatomo()
    } else {
      document.getElementById('matomo-script').addEventListener('load', initMatomo)
    }
  }
})()
