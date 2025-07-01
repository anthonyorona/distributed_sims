## Overview

This project simulates distributed system nodes to aid in learning distributed algorithms.

### Distributed Mutex Algorithm using Lamport Clocks

```
--- Current System State ---
PID: 0, Clock: 81, Queue Length: 9, State: Requested
PID: 1, Clock: 64, Queue Length: 10, State: Requested
PID: 2, Clock: 90, Queue Length: 9, State: Free
PID: 3, Clock: 66, Queue Length: 10, State: Requested
PID: 4, Clock: 87, Queue Length: 10, State: Requested
PID: 5, Clock: 75, Queue Length: 10, State: Requested
PID: 6, Clock: 90, Queue Length: 18, State: Holding
PID: 7, Clock: 81, Queue Length: 9, State: Requested
PID: 8, Clock: 84, Queue Length: 9, State: Requested
PID: 9, Clock: 87, Queue Length: 9, State: Requested
----------------------------
```

In this first simulation there are N (here 10) contending for a shared resource.

The distributed mutex algorithm uses lamport clocks to create a consistent total ordering.

In other words, the processes all see the same total ordering without being constrained by the need to synchronize centralized state.

This allows processes to coordinate their use of shared resources, on of many problems requiring coordination in a DS.

## Building the Executable

    ```
    go build -o bin/distributed_mutex ./cmd/distributed_mutex
    ```

## 4. Running the Simulation

1.  **Run Process 1**:
    ```
    ./bin/distributed_mutex
    ```
