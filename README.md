# StreamSite

Live streaming platform for dance studios.

## Overview

This site lets studios broadcast live classes to parents watching from home or on in-studio displays. Each studio gets isolated rooms with their own stream keys, viewer access control, and real-time status. Managed through a single admin interface.

## Tech Stack

- **Backend**: Go (auth, multi-tenancy, stream lifecycle, scheduling)
- **Frontend**: TypeScript / Preact
- **Streaming**: SRS (RTMP → HLS) + FFmpeg (RTSP transcoding)
- **Database**: BoltDB via vbolt
- **Auth**: JWT + OAuth2

## Getting Started

```bash
make dev        # start SRS + app (local)
make build      # compile frontend + Go binary
make deploy     # build and deploy to VPS
make test       # run Go tests
make typecheck  # TypeScript type checking
```
