This project simulates nodes in a distributed system for implementations of distributed algorithms

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

Here there are 10 processes contending for a shared resource.

The distributed mutex algorithm uses lamport clocks to create a consistent total ordering that allows the processes to synchornize their use of the shared resource.


## Building the Executable

    ```
    go build -o bin/distributed_mutex ./cmd/distributed_mutex
    ```

## 4. Running the Simulation

1.  **Run Process 1**:
    ```
    ./bin/distributed_mutex
    ```
