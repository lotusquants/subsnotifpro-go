# Google Cloud Pub/Sub Format Examples

## Overview
Google Cloud Pub/Sub can deliver messages in two formats depending on how the push subscription is configured:

1. **Wrapped Format** (default and most common)
2. **Unwrapped Format** (when push subscription is configured to unwrap messages)

## Wrapped Format Example

This is the default format when you create a push subscription without specifying unwrapping:

```json
{
  "message": {
    "data": "eyJ2ZXJzaW9uIjoiMS4wIiwicGFja2FnZU5hbWUiOiJjb20uZXhhbXBsZS5hcHAiLCJldmVudFRpbWVNaWxsaXMiOjE1MDMzNDk1NjYxNjgsInN1YnNjcmlwdGlvbk5vdGlmaWNhdGlvbiI6eyJ2ZXJzaW9uIjoiMS4wIiwibm90aWZpY2F0aW9uVHlwZSI6NCwicHVyY2hhc2VUb2tlbiI6IlBVUkNIQVNFX1RPS0VOIiwic3Vic2NyaXB0aW9uSWQiOiJwcmVtaXVtX21vbnRobHkifX0=",
    "messageId": "136969346945",
    "publishTime": "2021-02-26T19:13:55.749Z",
    "attributes": {
      "key": "value"
    }
  },
  "subscription": "projects/myproject/subscriptions/mysubscription"
}
```

The `data` field contains the base64-encoded RTDN payload.

## Unwrapped Format Example

This format is used when the push subscription is configured to unwrap messages:

```json
{
  "version": "1.0",
  "packageName": "com.example.app",
  "eventTimeMillis": 1503349566168,
  "subscriptionNotification": {
    "version": "1.0",
    "notificationType": 4,
    "purchaseToken": "PURCHASE_TOKEN",
    "subscriptionId": "premium_monthly"
  }
}
```

In this format, the RTDN payload is delivered directly without base64 encoding or wrapping.

## Push Subscription Configuration

### Wrapped Format (Default)
```bash
gcloud pubsub subscriptions create my-subscription \
    --topic=my-topic \
    --push-endpoint=https://myapp.com/webhook
```

### Unwrapped Format
```bash
gcloud pubsub subscriptions create my-subscription \
    --topic=my-topic \
    --push-endpoint=https://myapp.com/webhook \
    --push-no-wrapper
```

Or using the REST API:
```json
{
  "name": "projects/myproject/subscriptions/my-subscription",
  "topic": "projects/myproject/topics/my-topic",
  "pushConfig": {
    "pushEndpoint": "https://myapp.com/webhook",
    "noWrapper": true
  }
}
```

## Implementation Support

Our RTDN handler supports both formats automatically:

1. **Wrapped Format Detection**: If the request contains a `message` field with a `data` field, it's treated as wrapped format
2. **Unwrapped Format Fallback**: If wrapped format parsing fails, it falls back to unwrapped format
3. **Logging**: The handler logs which format is being processed for debugging

## Recommendations

- **Use Wrapped Format** for most cases as it's the default and provides additional metadata
- **Use Unwrapped Format** only if you need to minimize payload size or have specific processing requirements
- **Handle Both Formats** in your webhook handler to support different subscription configurations
