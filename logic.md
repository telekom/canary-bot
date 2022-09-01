
# Logic
## Startup
- init memory DB
- starting mesh server (grpc)
  - listen on joinMesh
  - listen on ping
  - listen on nodeDiscovery
  - listen on pushSamples
- starting timerRoutines
  - ping routine
  - sample routine
- starting channelRoutines
  - init newNodeDiscovered channel
  - init timeoutNode channel
- starting API
- (joining mesh)

### Mesh server
1. joinMesh
- add node to mesh
- send newNodeDiscovered signal
2. ping
- send just ping response
3. nodeDiscovery
- add node/s to mesh
- send newNodeDiscovered signal
4. pushSamples
- safe/update samples


### Ping routine
1. pings random x node/s
2. retry x times
3. send timeoutNode signal
4. timeout delay
5. timeout retry x times

- safe RTT as probe

### Sample routine
1. push sample to x random node/s
2. retry x times

### newNodeDiscovered channel
- send known nodes to x random node/s

### timeoutNode channel
- timeout delay
- timeout ping retry x times
- delete node from mesh