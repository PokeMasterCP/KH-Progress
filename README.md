# KH Progress

A Kingdom Hearts achievement journal backed by Steam.

## Run with Docker

```sh
docker build -t kh-progress .
docker run --rm -p 8080:8080 -e API_KEY -e STEAM_ID kh-progress
```

Set `API_KEY` (your Steam Web API key) and `STEAM_ID` in your shell first.
Open http://localhost:8080. Achievement data is fetched when the server starts.

## Publish to Docker Hub

In the GitHub repository's **Settings → Secrets and variables → Actions**, configure:

| Type | Name | Value |
| --- | --- | --- |
| Variable | `DOCKERHUB_USERNAME` | Docker Hub account used to publish |
| Secret | `DOCKERHUB_TOKEN` | Docker Hub access token with permission to push |
| Variable (optional) | `DOCKERHUB_NAMESPACE` | Organization owning the image; defaults to the username |

Create a `kh-progress` repository in that Docker Hub namespace with your preferred visibility.
Pushes to `main` run tests and publish `latest` and `sha-<commit>` image tags.
The workflow can also be run manually from the Actions tab.

`API_KEY` and `STEAM_ID` are runtime settings; they are not needed to build or publish the image.
