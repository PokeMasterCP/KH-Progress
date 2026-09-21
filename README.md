# KH Progress

A Kingdom Hearts achievement journal backed by Steam. View your unlocked and locked achievements, filter by game, and track your completion percentage across the HD 1.5 + 2.5 ReMIX collection.

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

Select **All games** or an individual game to update the achievement list and completion percentage.

The application makes two Steam API calls per page load, one for your achievement status and one for achievement names and icons. To get the latest data simply refresh the page.
