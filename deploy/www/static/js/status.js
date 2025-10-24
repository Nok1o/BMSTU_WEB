// Автообновление статуса каждые 5 секунд
function loadStatus() {
    fetch('/status/raw')
        .then(response => response.text())
        .then(data => {
            document.getElementById('status-content').textContent = data;
        })
        .catch(err => {
            document.getElementById('status-content').textContent = 'Ошибка загрузки статуса';
            console.error('Ошибка:', err);
        });
}

// Загрузить сразу
loadStatus();

// Обновлять каждые 5 секунд
setInterval(loadStatus, 5000);

