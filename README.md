# Collaborative Whiteboard

A real-time collaborative whiteboard backend built in **Go using WebSockets**.  
The server maintains a shared drawing state and synchronizes pixel updates across all connected clients with low latency.

---

## Backend Overview

The backend is implemented as a **WebSocket server with a central broadcast hub**.  
It maintains the global state of the whiteboard and propagates drawing updates to all connected clients.

The board is represented as a **640 × 480 grid**, where each cell corresponds to a pixel value.

When a client draws, the update is sent to the server, which updates the global grid and broadcasts the change to all clients.

---

## Concurrency Model

The system uses Go’s **goroutines and channels** to handle concurrent connections efficiently.

Each connected client has two goroutines:

### Read Goroutine
Responsible for receiving drawing updates from the client.

1. Reads a WebSocket message.
2. Parses the pixel update.
3. Sends the update to the **broadcast channel**.

### Write Goroutine
Responsible for sending updates from the server to the client.

1. Listens on the client's outbound channel.
2. Serializes updates.
3. Writes them to the WebSocket connection.

This separation ensures that **reading and writing operations do not block each other**.

---

## Broadcast Hub

The broadcast hub acts as the **central coordinator**.

- Receive pixel updates from clients
- Update the shared grid
- Broadcast the update to all active clients

The hub is implemented as a single goroutine to ensure that **grid updates occur sequentially**, avoiding race conditions.
