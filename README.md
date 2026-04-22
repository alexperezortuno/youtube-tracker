# youtube-tracker

## Get Channel ID

```bash
(source .env && curl \
  "https://youtube.googleapis.com/youtube/v3/channels?part=id&forUsername=${CHANNEL_NAME}&key=${YOUTUBE_API_KEY}" \
  --header "Accept: application/json" \
  --compressed)"
```

```bash
(source .env && echo "curl https://www.googleapis.com/youtube/v3/channels?part=id&forUsername=${CHANNEL_NAME}&key=${YOUTUBE_API_KEY} --header 'Accept: application/json' --compressed")
```

---

## Get Live Streams

```bash
(source .env && curl \
  "https://www.googleapis.com/youtube/v3/search?part=snippet&q=${VIDEO_ID}&type=channel&key=${YOUTUBE_API_KEY}" \
  --header "Accept: application/json" \
  --compressed)
```

---

## Get Live Streams

```bash
(source .env && curl \
  "https://www.googleapis.com/youtube/v3/search?part=snippet&channelId=${CHANNEL_ID}&eventType=live&type=video&key=${YOUTUBE_API_KEY}" \
  --header "Accept: application/json" \
  --compressed)
```

---

