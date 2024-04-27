ARG GO_IMAGE="golang:1.22.2"
ARG POSTGRES_IMAGE="postgres:alpine"

# Step 1: Modules caching
FROM ${GO_IMAGE} as modules
ARG GOPRIVATE_USER="__token__"
ARG GOPRIVATE_PAT=""
ARG GOPRIVATE=""
RUN apt update && apt install --reinstall ca-certificates
# Configure private repo
RUN echo -e machine "${GOPRIVATE}\nlogin ${GOPRIVATE_USER}\npassword ${GOPRIVATE_PAT}\n" > ~/.netrc
COPY go.mod go.sum /modules/
WORKDIR /modules
RUN go mod download

# Step 2: Builder
FROM ${GO_IMAGE} as builder

ARG APP_NAME="app"
ARG GOOS="linux"
ARG GOARCH="amd64"

COPY --from=modules /go/pkg /go/pkg
COPY . /${APP_NAME}
WORKDIR /${APP_NAME}
RUN CGO_ENABLED=0 GOOS=${GOOS} GOARCH=${GOARCH} \
    go build -o /bin/${APP_NAME} -ldflags="-w -s" ./cmd/${APP_NAME}

# Step 3: Final
FROM ${POSTGRES_IMAGE}

ARG APP_NAME="app"
COPY --from=builder /bin/${APP_NAME} /app

CMD ["/app"]
