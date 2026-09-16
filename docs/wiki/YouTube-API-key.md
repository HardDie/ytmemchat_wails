# YouTube API key

Google calls this an **API key**. People often say “token”. In ytmemchat it is the **YouTube API key** field on [Configuration](Configuration).

It is a password-like string from [Google Cloud](https://console.cloud.google.com/) that lets ytmemchat call YouTube’s official Data API. It is **not** your Google account password, and it is not a login as the channel owner.

Leave the field empty to read the public live chat page (no quota). Paste a key to use [YouTube Data API v3](https://developers.google.com/youtube/v3) only. A wrong or disabled key does **not** fall back to the no-key client; Home shows that the key is invalid and you must clear or fix it.

Without a key, YouTube may hide messages it marks as spam or otherwise “bad”. Those lines never reach ytmemchat. Details are on [Getting Started](Getting-Started#which-messages-you-will-see).

The key is stored in the local settings file (mode `0600`). It is never logged.

## How to get a key

1. Sign in to Google Cloud with a [Google Account](https://accounts.google.com/).
2. Create a project and enable **YouTube Data API v3**: [API Library](https://console.cloud.google.com/apis/library/youtube.googleapis.com). Google’s walkthrough is [YouTube Data API overview](https://developers.google.com/youtube/v3/getting-started).
3. Open [Credentials](https://console.cloud.google.com/apis/credentials) → **Create credentials** → **API key**. Copy the key into Configuration and Save.

You can restrict the key to YouTube Data API v3 so it cannot call other Google APIs. See [API key restrictions](https://cloud.google.com/docs/authentication/api-keys#api_key_restrictions).

## Default quota (why a long stream may run out)

A new Google Cloud project gets a **small daily budget**. Google currently gives **10,000 units per day** for most YouTube Data API calls (quota resets at midnight Pacific Time). Costs are listed in the [quota calculator](https://developers.google.com/youtube/v3/determine_quota_cost). Reading live chat (`liveChatMessages.list`) costs **1 unit per request**.

ytmemchat asks for new messages about every **5 seconds** while Start is running (or a bit faster/slower if YouTube sends a polling interval).

| | Amount |
|---|---|
| Requests per minute | 60 ÷ 5 = **12** |
| Units per hour | 12 × 60 = **720** |
| Hours on 10,000 units | 10,000 ÷ 720 ≈ **14 hours** of continuous Start |

So the default key is enough for about a **typical long stream**, not a full day of 24-hour chat with the official API. If YouTube asks to poll more often, the budget runs out sooner. Home then shows that quota was exceeded; wait for the daily reset or request more quota from the [YouTube Data API overview](https://developers.google.com/youtube/v3/getting-started#quota-usage). Chat without a key does not use this budget.

The Data API does **not** return leftover units, and asking Google Cloud for them needs extra credentials (not this API key). Home shows **spent units this stream** from the v3 requests ytmemchat already makes (this process). That count does not spend extra quota. It resets at **midnight Pacific Time** and when you **change the stream ID** (Save, Start, or Find latest that picks a different video). It can be lower than the Cloud Console if another tool shares the project, or if you already used the key before launching the app today.

## Find latest still fits the default key

**Find latest** only looks up a video/channel. That is one cheap `videos.list` (1 unit). If that video is not live, it may also call `search.list` (at most twice). Search has its own default cap of **100 calls per day**, which is still plenty for picking a new stream ID.

You can use the default key **just to get the current stream ID**, even if you do not want to spend the 10,000 units on live chat. Paste the key on Configuration, click Find latest, Save. If you then **Start** with the key still filled, chat uses the official API and the daily units. Leave the key empty to Start without quota.
