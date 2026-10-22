# Fallback: run DSA Quest anywhere Docker runs (Vercel container runtime,
# Render, Fly, or your own machine). The frontend is already built into web/dist.
FROM golang:1.24 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o /dsaquest .

FROM gcr.io/distroless/static-debian12
COPY --from=build /dsaquest /dsaquest
ENV PORT=8080
EXPOSE 8080
ENTRYPOINT ["/dsaquest"]
