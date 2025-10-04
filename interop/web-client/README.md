# MOQT Web Client

SolidJS-based MOQT (Media over QUIC Transport) Web client implementation

## Setup

### Prerequisites

1. **WebTransport Support**: Chrome/Edge with experimental features enabled
2. **HTTPS Required**: Even for localhost development
3. **MOQT Library**: Local TypeScript library at `../`

### Development Server

```bash
# Install dependencies
npm install

# Start development server
npm run dev
```

## Project Structure

- `/` - Home page (navigation)
- `/subscribe` - MOQT Subscribe functionality
- `/publish` - MOQT Publish functionality

## MOQT Library Integration

### Current Status

The project is configured to use the local MOQT TypeScript library located at `interop/web/`. However, there are currently module resolution issues that need to be resolved.

### Implementation Plan

1. **Fix Module Resolution**
   - Resolve package.json exports configuration
   - Ensure proper TypeScript declaration files
   - Test import/export functionality

2. **Subscribe Implementation**
   ```typescript
   import { Client } from '@okutanidaichi/moqt';

   const client = new Client();
   const session = await client.dial("https://localhost:4469/broadcast");
   const announceReader = await session.acceptAnnounce("/");

   // Process announcements and subscribe to tracks
   ```

3. **Publish Implementation**
   ```typescript
   import { Client, TrackMux } from '@okutanidaichi/moqt';

   const client = new Client();
   const session = await client.dial("https://localhost:4469/publish");

   // Set up publication handling
   ```

### Available MOQT Library APIs

From `@okutanidaichi/moqt`:

- `Client` - Main MOQT client class
- `Session` - MOQT session management
- `TrackMux` - Track multiplexing
- `AnnouncementReader` - Announcement stream handling
- `Publication` - Publication management
- `BroadcastPath` - Path utilities
- `TrackPrefix` - Track prefix utilities

### Current Implementation

For now, the pages use WebTransport API directly and simulate MOQT operations until the module resolution is fixed.

## Testing

1. **Start MOQT Server** (separate terminal):
   ```bash
   cd interop/server
   go run main.go
   ```

2. **Start Web Client**:
   ```bash
   npm run dev
   # Access http://localhost:3004 (port may vary)
   ```

3. **Test WebTransport**:
   - Navigate to `/subscribe` or `/publish`
   - Check browser console for WebTransport availability
   - Verify simulated functionality

## Development Notes

- **Module Resolution Issue**: Currently investigating package.json exports configuration
- **WebTransport**: Requires experimental browser features
- **HTTPS**: Required even for localhost development
- **Error Handling**: Comprehensive error logging in console

## Next Steps

1. Fix `@okutanidaichi/moqt` import resolution
2. Implement actual MOQT protocol communication
3. Connect with Go server for integration testing
4. Add proper error handling and UI feedback

## References

- [WebTransport API](https://developer.mozilla.org/en-US/docs/Web/API/WebTransport)
- [MOQT Draft Specification](https://datatracker.ietf.org/doc/draft-ietf-moq-transport/)
- Go version implementation: `interop/client/main.go`

The page will reload if you make edits.<br>

### `npm run build`

Builds the app for production to the `dist` folder.<br>
It correctly bundles Solid in production mode and optimizes the build for the best performance.

The build is minified and the filenames include the hashes.<br>
Your app is ready to be deployed!

## Deployment

You can deploy the `dist` folder to any static host provider (netlify, surge, now, etc.)

## This project was created with the [Solid CLI](https://github.com/solidjs-community/solid-cli)
