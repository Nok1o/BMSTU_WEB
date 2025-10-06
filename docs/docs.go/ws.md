# Документация WebSocket API

## Подключение

```http
GET /api/ws
Cookie: session={session_id}
```

## Типы сообщений

### 1. События сообщений

#### Отправка сообщения
**Тип:** `message`

**Запрос:**
```json
{
  "type": "message",
  "payload": {
    "text": "Текст сообщения",
    "chat_id": "3fa85f64-5717-4562-b3fc-2c963f66afa6",
    "receiver_id": "3fa85f64-5717-4562-b3fc-2c963f66afa6",
    "media": ["https://example.com/image1.jpg", "https://example.com/image2.jpg"],
    "audio": ["https://example.com/audio.mp3"],
    "file": ["https://example.com/document.pdf"]
  }
}
```

**Поля запроса:**
- `text` (string) - текстовое содержимое сообщения
- `chat_id` (uuid) - идентификатор чата (обязателен если не указан receiver_id)
- `receiver_id` (uuid) - идентификатор получателя (обязателен если не указан chat_id)
- `media` (string[]) - массив URL медиафайлов
- `audio` (string[]) - URL аудиофайлов
- `file` (string[]) - URL файлов

**Ответ:**
```json
{
  "type": "message",
  "payload": {
    "id": "3fa85f64-5717-4562-b3fc-2c963f66afa6",
    "text": "Текст сообщения",
    "sender_id": "3fa85f64-5717-4562-b3fc-2c963f66afa6",
    "chat_id": "3fa85f64-5717-4562-b3fc-2c963f66afa6",
    "created_at": "2023-10-05T14:30:00Z",
    "user": {
      "id": "3fa85f64-5717-4562-b3fc-2c963f66afa6",
      "username": "username",
      "avatar": "https://example.com/avatar.jpg"
    }
  }
}
```

#### Отметка о прочтении
**Тип:** `message_read`

**Запрос:**
```json
{
  "type": "message_read",
  "payload": {
    "chat_id": "3fa85f64-5717-4562-b3fc-2c963f66afa6",
    "message_id": "3fa85f64-5717-4562-b3fc-2c963f66afa6"
  }
}
```

**Поля запроса:**
- `chat_id` (uuid) - идентификатор чата
- `message_id` (uuid) - идентификатор сообщения

**Ответ:**
```json
{
  "type": "message_read",
  "payload": {
    "chat_id": "3fa85f64-5717-4562-b3fc-2c963f66afa6",
    "message_id": "3fa85f64-5717-4562-b3fc-2c963f66afa6",
    "ts": "2023-10-05T14:30:00Z",
    "sender_id": "3fa85f64-5717-4562-b3fc-2c963f66afa6"
  }
}
```

#### Удаление сообщения
**Тип:** `message_delete`

**Запрос:**
```json
{
  "type": "message_delete",
  "payload": {
    "message_id": "3fa85f64-5717-4562-b3fc-2c963f66afa6"
  }
}
```

**Поля запроса:**
- `message_id` (uuid) - идентификатор сообщения

**Ответ:**
```json
{
  "type": "message_delete",
  "payload": {
    "chat_id": "3fa85f64-5717-4562-b3fc-2c963f66afa6",
    "message_id": "3fa85f64-5717-4562-b3fc-2c963f66afa6"
  }
}
```

#### Удаление чата
**Тип:** `chat_delete`

**Запрос:**
```json
{
  "type": "chat_delete",
  "payload": {
    "chat_id": "3fa85f64-5717-4562-b3fc-2c963f66afa6"
  }
}
```

**Поля запроса:**
- `chat_id` (uuid) - идентификатор чата

**Ответ:**
```json
{
  "type": "chat_delete",
  "payload": {
    "chat_id": "3fa85f64-5717-4562-b3fc-2c963f66afa6"
  }
}
```

### 2. События друзей

#### Получен запрос дружбы
**Тип:** `fr_received`

```json
{
  "type": "fr_received",
  "payload": {
    "id": "3fa85f64-5717-4562-b3fc-2c963f66afa6",
    "username": "username",
    "avatar": "https://example.com/avatar.jpg",
    "last_seen": "2023-10-05T14:30:00Z"
  }
}
```

#### Запрос дружбы принят
**Тип:** `fr_accepted`

```json
{
  "type": "fr_accepted",
  "payload": {
    "id": "3fa85f64-5717-4562-b3fc-2c963f66afa6",
    "username": "username",
    "avatar": "https://example.com/avatar.jpg",
    "last_seen": "2023-10-05T14:30:00Z"
  }
}
```

### 3. События постов

#### Лайк поста
**Тип:** `post_liked`

```json
{
  "type": "post_liked",
  "payload": {
    "post": {
      "id": "3fa85f64-5717-4562-b3fc-2c963f66afa6",
      "content": "Текст поста",
      "created_at": "2023-10-05T14:30:00Z"
    },
    "user": {
      "id": "3fa85f64-5717-4562-b3fc-2c963f66afa6",
      "username": "username",
      "avatar": "https://example.com/avatar.jpg"
    }
  }
}
```

#### Лайк комментария
**Тип:** `comment_liked`

```json
{
  "type": "comment_liked",
  "payload": {
    "comment": {
      "id": "3fa85f64-5717-4562-b3fc-2c963f66afa6",
      "content": "Текст комментария",
      "user_id": "3fa85f64-5717-4562-b3fc-2c963f66afa6"
    },
    "user": {
      "id": "3fa85f64-5717-4562-b3fc-2c963f66afa6",
      "username": "username",
      "avatar": "https://example.com/avatar.jpg"
    }
  }
}
```

#### Комментарий к посту
**Тип:** `post_commented`

```json
{
  "type": "post_commented",
  "payload": {
    "post": {
      "id": "3fa85f64-5717-4562-b3fc-2c963f66afa6",
      "content": "Текст поста"
    },
    "comment": {
      "id": "3fa85f64-5717-4562-b3fc-2c963f66afa6",
      "content": "Текст комментария"
    }
  }
}
```

## Формат ошибок

```json
{
  "error_code": "validation_error",
  "message": "Сообщение не может быть пустым"
}
```