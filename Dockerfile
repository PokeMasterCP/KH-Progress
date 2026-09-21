FROM golang:1.27.1 AS build

WORKDIR /app
COPY go.mod ./
COPY main.go steamAPI.go index.html ./
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /kh-progress .

FROM scratch

# Steam API requests require trusted HTTPS certificates.
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /kh-progress /kh-progress

USER 65532:65532
EXPOSE 8080

# Supply API_KEY and STEAM_ID at runtime.
ENTRYPOINT ["/kh-progress"]
