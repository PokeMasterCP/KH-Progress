# KH Progress

A Kingdom Hearts achievement journal backed by Steam. Track your completion across the four games in the HD 1.5 + 2.5 ReMIX collection, see what you unlocked most recently, and search or filter every achievement by game and status.

## Configuration

The application requires two environment variables:

| Variable | Description |
| --- | --- |
| `API_KEY` | Your Steam Web API key |
| `STEAM_ID` | Your numeric Steam ID (SteamID64) |

Set them in your shell before running the application:

```sh
export API_KEY="your-steam-api-key"
export STEAM_ID="your-steam-id"
```

## Run with Docker

```sh
docker run --rm -p 8080:8080 \
  -e API_KEY -e STEAM_ID \
  pokemastercp/kh-progress:latest
```

To build and run the image from the source code instead:

```sh
docker build -t kh-progress .
docker run --rm -p 8080:8080 -e API_KEY -e STEAM_ID kh-progress
```

## Run with Go

With Go 1.27.1 or later installed, run from the project directory:

```sh
go run .
```

## Using the journal

Open `http://<container-ip>:8080` if the container's IP is reachable from your browser. With the `-p 8080:8080` mapping above, you can use `http://<docker-host-ip>:8080` instead.

The page is designed for phones first and is organised into four parts:

- **Overview**: a ring with one pane of glass per achievement, grouped by game. Gold panes are unlocked. The centre shows your overall percentage, with your unlocked count and the dates of your first and latest unlocks below. Hover over a pane to see which achievement it is, or tap it to jump to that achievement in the list.
- **By game**: one row per game with its unlocked count, percentage, a bar with a segment per achievement, and how long ago you last unlocked one there. Tap a game to jump to its achievements.
- **Recent unlocks**: your five most recent unlocks, newest first.
- **All achievements**: every achievement with its icon, description, game and unlock date. The filter bar stays at the top of the screen while you scroll: search by name, description or Steam API name (`ACH_###`), show only unlocked or locked achievements, and switch between games. Sort by game order or newest first. Hidden achievements show their name, but Steam doesn't provide their descriptions.

Dates and times are shown in your browser's time zone. The current filters are kept in the page address (for example `?game=kh2&status=locked`), so refreshing or bookmarking the page keeps the same view. The page follows your system's light or dark appearance setting.

The application makes two Steam API calls per page load: one for your achievement status and unlock times, and one for achievement names, descriptions and icons. The header shows when the data was last synced. To get the latest data, refresh the page or select the refresh button beside the sync time.

If Steam doesn't respond, the page explains the problem and offers **Try again**. If it keeps failing, check that `API_KEY` and `STEAM_ID` are correct and that your Steam profile's **Game details** privacy setting is **Public**.

Fonts are loaded from Google Fonts and achievement icons from Steam's CDN. If they can't be reached, the page falls back to system fonts and a placeholder icon.

## Development

Run the tests with:

```sh
go test ./...
```

The page template is `index.html`, which is embedded into the binary at build time. Restart `go run .` after editing it.
