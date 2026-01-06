# Application Configuration


## Telegram Bot

| Setting               | Variable                               | Default | Example                                   | Description                        |
|-----------------------|----------------------------------------|---------|-------------------------------------------|------------------------------------|
| Bot token             | CONFIGURATION_BOT_API_TOKEN            | —       | 123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11 | Telegram Bot API token             |
| Drop pending updates  | CONFIGURATION_BOT_DROP_PENDING_UPDATES | true    | true                                      | If true, skips accumulated updates |
| Request timeout (sec) | CONFIGURATION_BOT_REQUEST_TIMEOUT_SEC  | 10      | 15                                        | Timeout for Telegram API requests  |

## Webhook

| Setting        | Variable                               | Default | Example             | Description                     |
|----------------|----------------------------------------|---------|---------------------|---------------------------------|
| Base URL       | CONFIGURATION_BOT_WEBHOOK_URL          | —       | https://example.com | Base URL for the webhook        |
| Path           | CONFIGURATION_BOT_WEBHOOK_PATH         | /bot    | /webhook            | Webhook endpoint path           |
| Listen address | CONFIGURATION_BOT_WEBHOOK_LISTEN_ADDR  | :8080   | 0.0.0.0:8080        | Local server address and port   |
| Secret token   | CONFIGURATION_BOT_WEBHOOK_SECRET_TOKEN | —       | super-secret-token  | Secret for webhook verification |

## Google API

| Setting  | Variable                     | Default | Example                    | Description                                                                          |
|----------|------------------------------|---------|----------------------------|--------------------------------------------------------------------------------------|
| API keys | CONFIGURATION_GOOGLE_API_KEY | —       | key1,key2,key3             | Comma-separated keys; first key is used for videos, remaining keys rotate for search |

## Redis (cache)

Cache environment variable prefix: `CONFIGURATION_CACHER_`.

| Setting        | Variable                            | Default          | Example        | Description    |
|----------------|-------------------------------------|------------------|----------------|----------------|
| Redis address  | CONFIGURATION_CACHER_REDIS_ADDR     | localhost:6379   | redis:6379     | Redis address  |
| Redis username | CONFIGURATION_CACHER_REDIS_USERNAME | app              | bot            | Redis username |
| Redis password | CONFIGURATION_CACHER_REDIS_PASSWORD | local-redis-pass | my-strong-pass | Redis password |
