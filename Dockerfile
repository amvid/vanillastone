# Dev image: hot reload via wgo. Source mounted at runtime (see compose).
FROM golang:1.26-alpine AS dev

WORKDIR /app

# wgo = file watcher, reruns on save. https://github.com/bokwoon95/wgo
RUN go install github.com/bokwoon95/wgo@latest

# Cache deps first. go.sum optional (none yet) — glob avoids COPY failure.
COPY go.mod go.sum* ./
RUN go mod download

COPY . .

EXPOSE 8080

# Watch ./ , rerun server on change.
CMD ["wgo", "run", "./cmd/server"]


# Web build: compile the React client to web/static so the Go binary embeds it.
FROM node:24-alpine AS web-build
WORKDIR /web
RUN corepack enable
COPY web/package.json web/pnpm-lock.yaml* ./
RUN pnpm install --frozen-lockfile
COPY web/ .
RUN pnpm build

# Prod image: static binary, no toolchain. Used later, not for dev loop.
FROM golang:1.26-alpine AS build
WORKDIR /app
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
# Overwrite the committed placeholder static/ with the built client.
COPY --from=web-build /web/static ./web/static
RUN CGO_ENABLED=0 go build -o /vanillastone ./cmd/server

FROM gcr.io/distroless/static-debian12 AS prod
COPY --from=build /vanillastone /vanillastone
EXPOSE 8080
ENTRYPOINT ["/vanillastone"]
