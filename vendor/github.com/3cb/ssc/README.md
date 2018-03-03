# Simple Socket Controller
Built on top of gorilla toolkit's websocket package, this library allows caller to create and control a pool of websockets

NOTE: This package is still under development.

To import the package into your application:

```bash
import "github.com/3cb/ssc"
```

To install the package on your system:

```bash
go get github.com/3cb/ssc
```
## Pool Layout
Each WebSocket has its own readSocket and writeSocket goroutine but there is only one instance of each control goroutine per Pool:


![Diagram](https://images2.imgbox.com/ed/94/2MkYE7Np_o.png?download=true)

## Example Usage

First create an instance of `ssc.Pool` by calling `ssc.NewPool()` which takes a configuration object as a parameter:
```go
sockets := []string{
    "wss://api.example.com/ws1",
    "wss://api.example.com/ws2",
    "wss://api.example.com/ws3",
}

pool := ssc.NewPool(sockets, time.Second*45)
err := pool.Start()
if err != nil {
    log.Fatalf("Error starting new Pool.")
}
```
The above example starts a pool of websocket server connections.  In order to start an empty pool which is ready for client websockets to connect, send an empty slice for the first parameter:
```go

pool, err := ssc.NewPool([]string{}, time.Second*45])
err := pool.Start()
if err != nil {
    log.Fatalf("Error starting new Pool.")
}
```
To add individual client connections call `pool.AddClientSocket()` from a route handler:

```go
func main() {
	r := mux.NewRouter()	// gorilla mux package

	pool := ssc.NewPool([]string{}, time.Second*30)
	err := pool.Start()
	if err != nil {
	    log.Fatalf("Error starting new Pool.")
	}

	upgrader := &websocket.Upgrader{}	// gorilla websocket package
	r.Handle("/". wsHandler(pool, upgrader))

	log.Fatal(http.ListenAndServe(":3000", r))
}

func wsHandler(pool *ssc.Pool, upgrader *websocket.Upgrader) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := pool.AddClientSocket("clientID", upgrader, w, r)
		if err != nil {
			// handle error
		}
		// do other stuff
	})
}
```
Caller application can also add individual server connections with `pool.AddServerSocket`:
```go
err := pool.AddServerSocket(URL)
if err != nil {
	// handle error
}
```
Data is moved in instances of the Message type.
```go
type Message struct {
	// ID is the raw URL string for server connections and is set by caller application for client connections.
	// The read goroutine sets the ID equal to the ID of the Socket before sending it to the pool to be sent to the outbound channel.
	// Caller app must be careful to change the ID on the Message if it needs to send it to a specific socket in a Pool of client sockets
	ID      string

	// Type are defined in the websocket protocol: https://tools.ietf.org/html/rfc6455
	// They are also described here: http://www.gorillatoolkit.org/pkg/websocket#constants
	Type    int

	// Payload is any message serialized into a slice of bytes
	Payload []byte
}
```
For writing to sockets from the caller application there are a few options:
```go
// Write sends message to specific socket based on the id string
// returns an error if Message.ID parameter is empty or not in stack
err := pool.Write(&Message{ID: "client1", Type: websocket.TextMessage, Payload: []byte("Hello!")})
if err != nil {
	// handle error
}

//WriteAll sends message to all sockets in `Pool` regardless of the ID string in Message
pool.WriteAll(&Message{Type: websocket.TextMessage, Payload: []byte("Hello!")})

// Another way to send a Message to all connected websockets is to send directly into the Pool.Inbound channel
pool.Inbound <- &Message{Type: websocket.TextMessage, Payload: []byte("Hello!")}
```
ssc provides a coupld of convenience methods as well:
```go
// Count returns the number of connections active in the `Pool`
count := pool.Count()

// List returns a slice of ID strings for the connections active in the pool
connections := pool.List()
```
Lastly, if the caller application has no more need for the `Pool` it can use:
```go
pool.Stop()
```
The above will cleanup all running goroutines.  This method will shutdown all read/write goroutines, wait for confirmation, and then shut down all control goroutines that are running.



Usage for server WebSockets here: https://github.com/3cb/gemini_clone/tree/go_stream

Usage for client WebSockets here: https://github.com/3cb/seattle911 and here: https://github.com/3cb/melbourne_parking

## Work Left To Do

-> MORE TESTS?