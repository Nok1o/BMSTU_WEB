try {
  var settings = JSON.parse(localStorage.getItem('settings'))
  var head = document.getElementsByTagName('head')[0]
  var link = document.createElement('link')
  link.rel = 'stylesheet'
  link.type = 'text/css'
  link.href = settings.darkMode ? './style/dark-load.css' : './style/light-load.css'
  link.media = 'all'
  head.appendChild(link)
} catch (e) {
  var head = document.getElementsByTagName('head')[0]
  var link = document.createElement('link')
  link.rel = 'stylesheet'
  link.type = 'text/css'
  link.href = './style/light-load.css'
  link.media = 'all'
  head.appendChild(link)
}
