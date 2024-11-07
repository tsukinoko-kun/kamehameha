FROM alpine:3 AS journey
RUN apk add --no-cache curl wget jq bash
