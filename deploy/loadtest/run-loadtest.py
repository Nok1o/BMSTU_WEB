#!/usr/bin/env python3

import os
import sys
import time
import random
import string
import subprocess
from datetime import datetime
from typing import Dict, List
import requests


# -------------------------------
# КОНФИГУРАЦИЯ
# -------------------------------

NGINX_LOG_PATH = "/opt/homebrew/var/log/nginx/access.log"
BASE_URL = "http://localhost:8090"
TEST_ID = f"lt-{int(time.time())}"

CYCLES_GET = 500
CYCLES_POST = 200

REPORT_DIR = "./reports"
os.makedirs(REPORT_DIR, exist_ok=True)
REPORT_PATH = os.path.join(REPORT_DIR, f"report-{TEST_ID}.md")

HEADERS = {
    "accept": "application/json",
    "Content-Type": "application/json",
    "X-Test-ID": TEST_ID
}


# -------------------------------
# ВСПОМОГАТЕЛЬНЫЕ ФУНКЦИИ
# -------------------------------

def log(message: str, level: str = "INFO"):
    timestamp = datetime.now().strftime("%H:%M:%S")
    print(f"[{timestamp}] {level:5} | {message}")


def generate_username(prefix: str = "user") -> str:
    suffix = ''.join(random.choices(string.ascii_lowercase + string.digits, k=6))
    return f"{prefix}_{suffix}"


def clear_nginx_log():
    log("Очищаем лог Nginx...")
    try:
        subprocess.run(["sh", "-c", f"truncate -s 0 {NGINX_LOG_PATH}"], check=True)
        time.sleep(1)
    except subprocess.CalledProcessError as e:
        log(f"Ошибка при очистке лога: {e}", "ERROR")
        sys.exit(1)


def read_logs() -> List[str]:
    try:
        with open(NGINX_LOG_PATH, 'r') as f:
            return [line.strip() for line in f if TEST_ID in line]
    except Exception as e:
        log(f"Ошибка чтения лога: {e}", "ERROR")
        return []


def analyze_upstream(logs: List[str]) -> Dict[str, int]:
    stats = {"8080": 0, "8070": 0, "8060": 0}
    for line in logs:
        if "127.0.0.1:8080" in line or "[::1]:8080" in line:
            stats["8080"] += 1
        if "127.0.0.1:8070" in line:
            stats["8070"] += 1
        if "127.0.0.1:8060" in line:
            stats["8060"] += 1
    return stats



def make_post(url: str, data: dict, expected_status: int = 200) -> int:
    try:
        # log(f"POST {url}")
        response = requests.post(url, headers=HEADERS, json=data, timeout=10)
        if response.status_code == expected_status:
            #log(f"Успешно: {response.status_code}")
            return response.status_code
        else:
            log(f"Ошибка: {response.status_code} {response.text}", "ERROR")
            return response.status_code
    except Exception as e:
        log(f"Исключение: {e}", "ERROR")
        return 500


def make_get(url: str, expected_status: int = 200) -> int:
    try:
        #log(f"GET {url}")
        response = requests.get(url, headers=HEADERS, timeout=10)
        if response.status_code == expected_status:
            #log(f"Успешно: {response.status_code}")
            return response.status_code
        else:
            log(f"Ошибка: {response.status_code}", "ERROR")
            return response.status_code
    except Exception as e:
        log(f"Исключение: {e}", "ERROR")
        return 500


# -------------------------------
# ГЕНЕРАЦИЯ ОТЧЁТА
# -------------------------------

def generate_report(test_results: Dict):
    log(f"Генерация отчёта: {REPORT_PATH}")

    with open(REPORT_PATH, "w") as f:
        f.write("# Автогенерируемый отчёт о нагрузочном тестировании\n\n")
        f.write(f"**Дата:** {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}\n")
        f.write(f"**Тест-ID:** `{TEST_ID}`\n")
        f.write(f"**Nginx:** brew (macOS)\n\n")

        # --- Тест 1: Регистрация и логин ---
        f.write("## Тест 1: Регистрация и аутентификация\n")
        f.write(f"- Username: `{test_results['username']}`\n")
        f.write(f"- Регистрация: `{test_results['signup_status']}`\n")
        f.write(f"- Логин: `{test_results['login_status']}`\n")

        # --- Тест 2: GET-запросы — балансировка ---
        f.write("\n## Тест 2: GET-запросы — балансировка 2:1:1\n")

        stats = test_results["get_stats"]
        total = sum(stats.values())

        f.write("| Порт | Запросов | Доля |\n")
        f.write("|------|----------|------|\n")
        for port, count in sorted(stats.items()):
            percent = (count / total * 100) if total > 0 else 0
            f.write(f"| {port} | {count} | {percent:.1f}% |\n")

        # Визуализация
        f.write("\n#### Распределение\n```\n")
        max_width = 50
        for port, count in sorted(stats.items()):
            bar_width = int((count / total) * max_width) if total > 0 else 0
            bar = "█" * bar_width
            f.write(f"{port}: {bar} {count}\n")
        f.write("```\n")

        # --- Тест 3: POST-запросы — только на 8080 ---
        f.write("\n## Тест 3: POST — только на 8080\n")
        post_stats = test_results["post_stats"]
        f.write("| Порт | POST |\n")
        f.write("|------|------|\n")
        f.write(f"| 8080 | {post_stats['8080']} |\n")
        f.write(f"| 8070/8060 | {post_stats['other']} |\n")

        if post_stats['other'] == 0:
            f.write("\nВсе POST-запросы ушли на 8080.\n")
        else:
            f.write("\nОШИБКА: POST на ro-инстансы!\n")

        # --- Тест 4: Прямой POST на ro ---
        f.write("\n## Тест 4: Прямой POST на ro (8070)\n")
        f.write("```bash\n")
        f.write("curl -X 'POST' 'http://localhost:8070/api/v2/users' \\\n")
        f.write("  -H 'Content-Type: application/json' \\\n")
        f.write("  -d '{\"username\":\"test\",\"password\":\"pass\",...}'\n")
        f.write("```\n\n")

        f.write(f"**Код ответа:** `{test_results['ro_post_status']}`\n")

        if test_results["ro_post_status"] == 403:
            f.write("**Результат:** Ожидаемый ответ 403.\n")
        else:
            f.write("**Результат:** Ожидаемый код: `403`, получен: `{}`\n".format(test_results['ro_post_status']))

        if test_results.get("ro_post_body"):
            f.write("\n**Тело ответа:**\n")
            f.write("```\n")
            f.write(f"{test_results['ro_post_body']}\n")
            f.write("```\n")

    log(f"Отчёт сохранён: {REPORT_PATH}")


# -------------------------------
# ОСНОВНОЙ СЦЕНАРИЙ
# -------------------------------

def main():
    log("Запуск нагрузочного тестирования")

    # Очищаем лог только один раз — в начале
    #clear_nginx_log()

    results = {
        "username": "",
        "signup_status": "",
        "login_status": "",
        "get_stats": {"8080": 0, "8070": 0, "8060": 0},
        "post_stats": {"8080": 0, "other": 0},
        "ro_post_status": 0
    }

    # --- 1. Регистрация ---
    username = generate_username("lt_user")
    results["username"] = username

    signup_data = {
        "username": username,
        "password": "SecurePass123!",
        "firstname": "Test",
        "lastname": "User",
        "birth_date": "1990-01-01",
        "sex": 1
    }

    status = make_post(f"{BASE_URL}/api/v2/users", signup_data, 201)
    results["signup_status"] = "201 Created" if status == 201 else str(status)

    # --- 2. Успешный логин ---
    login_data = {
        "username": username,
        "password": "SecurePass123!"
    }

    status = make_post(f"{BASE_URL}/api/v2/sessions", login_data, 201)
    results["login_status"] = "201 Created" if status == 201 else str(status)

    # --- 3. GET-запросы — балансировка ---
    log(f"Запуск {CYCLES_GET} GET-запросов к /api/v2/health...")
    for i in range(CYCLES_GET):
        make_get(f"{BASE_URL}/api/v2/health")

    logs = read_logs()
    get_logs = [line for line in logs if "GET" in line and "health" in line]
    results["get_stats"] = analyze_upstream(get_logs)

    # --- 4. POST-запросы — только на 8080 ---
    log(f"Запуск {CYCLES_POST} POST-запросов к /api/v2/sessions (логин)...")
    for i in range(CYCLES_POST):
        new_username = generate_username("login_test")
        signup_data = {
            "username": new_username,
            "password": "SecurePass123!",
            "firstname": "Login",
            "lastname": "Test",
            "birth_date": "1990-01-01",
            "sex": 1
        }
        # Регистрация
        make_post(f"{BASE_URL}/api/v2/users", signup_data, 201)
        # Логин
        login_data = {
            "username": new_username,
            "password": "SecurePass123!"
        }
        make_post(f"{BASE_URL}/api/v2/sessions", login_data, 201)

    logs = read_logs()
    post_logs = [line for line in logs if "POST" in line and "sessions" in line]
    post_stats = analyze_upstream(post_logs)
    results["post_stats"] = {
        "8080": post_stats["8080"],
        "other": post_stats["8070"] + post_stats["8060"]
    }

    # --- 5. Прямой POST на ro ---
    log("Проверка: прямой POST /api/v2/users на ro-инстанс (8070)...")
    try:
        response = requests.post(
            "http://localhost:8070/api/v2/users",
            json={
                "username": "blockeduser" + str(int(time.time()))[:4],
                "password": "SecurePass123!",
                "firstname": "Blocked",
                "lastname": "Test",
                "birth_date": "1990-01-01",
                "sex": 1
            },
            headers={"Content-Type": "application/json"},
            timeout=5
        )
        results["ro_post_status"] = response.status_code
        results["ro_post_body"] = response.text.strip()
    except Exception as e:
        log(f"Ошибка при запросе к ro: {e}", "ERROR")
        results["ro_post_status"] = 500
        results["ro_post_body"] = str(e)

    # --- Генерация отчёта ---
    generate_report(results)

    log("Тестирование завершено")


if __name__ == "__main__":
    main()
