# youtube-tracker

# youtube-tracker

## Get Channel ID

```bash
curl \
  'https://youtube.googleapis.com/youtube/v3/channels?part=id&forUsername=CHANNEL&key=${YOUTUBE_API_KEY}' \
  --header 'Accept: application/json' \
  --compressed
```

---

## Get Live Streams

```bash
curl \
  'https://www.googleapis.com/youtube/v3/search?part=snippet&q=CHANNEL&type=channel&key=${YOUTUBE_API_KEY}' \
  --header 'Accept: application/json' \
  --compressed
```

---

## Get Live Streams

```bash
curl \
  'https://www.googleapis.com/youtube/v3/search?part=snippet&channelId=CHANNEL_ID&eventType=live&type=video&key=${YOUTUBE_API_KEY} \
  --header 'Accept: application/json' \
  --compressed
```

---

