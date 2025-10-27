import requests
import random
import time
import threading
import uuid
from concurrent.futures import ThreadPoolExecutor

class ApiSpammer:
    def __init__(self, base_url="http://localhost:8090"):
        self.base_url = base_url
        self.session = requests.Session()
        self.users = []
        self.sessions = []

    def generate_user_data(self):
        first_names = ["John", "Jane", "Alex", "Maria", "Mike", "Sarah", "David", "Emma", "Chris", "Lisa"]
        last_names = ["Smith", "Johnson", "Brown", "Davis", "Wilson", "Miller", "Taylor", "Anderson", "Thomas", "Jackson"]

        username = f"user_{random.randint(1000, 9999)}"
        firstname = random.choice(first_names)
        lastname = random.choice(last_names)

        return {
            "username": username,
            "password": "SecurePass123!",
            "firstname": firstname,
            "lastname": lastname,
            "birth_date": f"199{random.randint(0, 9)}-{random.randint(1, 12):02d}-{random.randint(1, 28):02d}",
            "sex": random.randint(0, 1)
        }

    def register_user(self, prefix=""):
        try:
            user_data = self.generate_user_data()

            # Иногда создаем дублирующих пользователей
            if random.random() < 0.2 and self.users:
                user_data = random.choice(self.users)

            response = self.session.post(
                f"{self.base_url}{prefix}/api/v2/users",
                json=user_data,
                headers={"accept": "application/json", "Content-Type": "application/json"},
                timeout=5
            )

            if response.status_code == 201:
                print(f"✅ {prefix} User registered: {user_data['username']}")
                self.users.append(user_data)
                return user_data
            else:
                print(f"❌ {prefix} Registration failed ({response.status_code}): {response.text}")
                return None

        except Exception as e:
            print(f"{prefix} Registration error: {e}")
            return None

    def login_user(self, prefix=""):
        try:
            # Иногда пробуем несуществующих пользователей
            if random.random() < 0.3:
                username = f"nonexistent_{random.randint(10000, 99999)}"
                password = "wrongpassword"
            elif self.users:
                user = random.choice(self.users)
                username = user["username"]
                password = user["password"]
            else:
                return None

            login_data = {
                "username": username,
                "password": password
            }

            response = self.session.post(
                f"{self.base_url}{prefix}/api/v2/sessions",
                json=login_data,
                headers={"accept": "application/json", "Content-Type": "application/json"},
                timeout=5
            )

            if response.status_code == 201:
                session_data = response.json()
                print(f"✅ {prefix} Login successful: {username}")
                self.sessions.append(session_data["id"])
                return session_data["id"]
            else:
                print(f"❌ {prefix} Login failed ({response.status_code}): {username}")
                return None

        except Exception as e:
            print(f"{prefix} Login error: {e}")
            return None

    def search_users(self, session_id=None, prefix=""):
        try:
            headers = {"accept": "application/json"}
            if session_id:
                headers["Cookie"] = f"session={session_id}"

            search_terms = ["john", "alex", "maria", "tech", "test", "user", ""]
            term = random.choice(search_terms)

            response = self.session.get(
                f"{self.base_url}{prefix}/api/v2/users",
                params={"to_search": term, "count": random.randint(5, 20)},
                headers=headers,
                timeout=5
            )

            if response.status_code == 200:
                print(f"✅ {prefix} Search successful: '{term}'")
            else:
                print(f"❌ {prefix} Search failed ({response.status_code}): '{term}'")

        except Exception as e:
            print(f"{prefix} Search error: {e}")

    def create_community(self, session_id, prefix=""):
        try:
            if not session_id:
                return

            communities = [
                {"nickname": "techlovers", "name": "Tech Enthusiasts", "desc": "Сообщество для обсуждения технологий"},
                {"nickname": "gamers", "name": "Gaming Community", "desc": "Все о видеоиграх"},
                {"nickname": "musicfans", "name": "Music Lovers", "desc": "Обсуждаем музыку"},
                {"nickname": "bookclub", "name": "Book Club", "desc": "Читаем и обсуждаем книги"},
                {"nickname": "travelers", "name": "Travel Community", "desc": "Путешествия по миру"}
            ]

            community = random.choice(communities)

            data = {
                "nickname": community["nickname"],
                "name": community["name"],
                "description": community["desc"],
                "avatar": "",
                "cover": ""
            }

            response = self.session.post(
                f"{self.base_url}{prefix}/api/v2/communities",
                files=data,
                headers={"Cookie": f"session={session_id}"},
                timeout=5
            )

            if response.status_code == 201:
                print(f"✅ {prefix} Community created: {community['name']}")
            else:
                print(f"❌ {prefix} Community creation failed ({response.status_code}): {community['name']}")

        except Exception as e:
            print(f"{prefix} Community error: {e}")

    def create_post(self, session_id, prefix=""):
        try:
            if not session_id:
                return

            posts_texts = [
                "Отличный пост! Мне очень понравилось.",
                "Сегодня замечательный день!",
                "Поделюсь интересной находкой...",
                "Что вы думаете об этом?",
                "Новое обновление вышло!",
                "Интересные мысли по поводу..."
            ]

            post_data = {
                "text": random.choice(posts_texts),
                "media": [str(uuid.uuid4())] if random.random() > 0.7 else [],
                "files": [str(uuid.uuid4())] if random.random() > 0.8 else [],
                "audio": [str(uuid.uuid4())] if random.random() > 0.9 else []
            }

            response = self.session.post(
                f"{self.base_url}{prefix}/api/v2/posts",
                json=post_data,
                headers={
                    "accept": "application/json",
                    "Content-Type": "application/json",
                    "Cookie": f"session={session_id}"
                },
                timeout=5
            )

            if response.status_code == 200:
                print(f"✅ {prefix} Post created")
            else:
                print(f"❌ {prefix} Post creation failed ({response.status_code})")

        except Exception as e:
            print(f"{prefix} Post error: {e}")

    def get_feed(self, session_id=None, prefix=""):
        try:
            headers = {"accept": "application/json"}
            if session_id:
                headers["Cookie"] = f"session={session_id}"

            feed_types = ["recommendations", "friends", "subscriptions", "unknown_type"]
            feed_type = random.choice(feed_types)

            response = self.session.get(
                f"{self.base_url}{prefix}/api/v2/posts",
                params={
                    "count": random.randint(5, 20),
                    "type": feed_type
                },
                headers=headers,
                timeout=5
            )

            if response.status_code == 200:
                print(f"✅ {prefix} Feed loaded: {feed_type}")
            else:
                print(f"❌ {prefix} Feed failed ({response.status_code}): {feed_type}")

        except Exception as e:
            print(f"{prefix} Feed error: {e}")

    def run_spam_cycle(self):
        actions = [
            (self.register_user, 0.3),
            (self.login_user, 0.4),
            (self.search_users, 0.6),
            (self.create_community, 0.1),
            (self.create_post, 0.2),
            (self.get_feed, 0.7),
        ]

        prefixes = ["", "/mirror"]

        # Выбираем случайные действия на основе вероятности
        for action, probability in actions:
            if random.random() < probability:
                prefix = random.choice(prefixes)

                # Для действий требующих сессии, передаем случайную сессию если есть
                if action in [self.create_community, self.create_post]:
                    session_id = random.choice(self.sessions) if self.sessions else None
                    threading.Thread(target=action, args=(session_id, prefix)).start()
                elif action in [self.search_users, self.get_feed]:
                    session_id = random.choice(self.sessions) if self.sessions and random.random() > 0.3 else None
                    threading.Thread(target=action, args=(session_id, prefix)).start()
                else:
                    threading.Thread(target=action, args=(prefix,)).start()

                time.sleep(random.uniform(0.1, 0.5))

def main():
    spammer = ApiSpammer()

    print("Starting API spammer...")
    print("Targets: http://localhost:8090/api/v2 and http://localhost:8090/mirror/api/v2")
    print("Press Ctrl+C to stop\n")

    try:
        while True:
            spammer.run_spam_cycle()
            time.sleep(random.uniform(0.5, 2.0))
    except KeyboardInterrupt:
        print("\nStopping spammer...")

if __name__ == "__main__":
    main()