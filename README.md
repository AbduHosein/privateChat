Goal: a simple chat room application that supports only private messages

# Private Message Server and Client

## Instructions
1. Clone the privateChat repository to your machine
2. Open as many terminals as you want clients, and one additional terminal to run the server
3. Start by running the server, providing a port number for the server
    - `go run main.go SERVER {port}`
    - Example: `go run main.go SERVER 8080`
4. Launch as many clients as you want, providing command-line arguments in the form {host}:{port} {username}, where the port is the port previously provided to launch the server. If running locally, the {host} argument, will always be 'localhost'. 
    - `go run main.go CLIENT {host}:{port} {username}`
    - Example 1 `go run main.go CLIENT localhost:8080 lewis`
    - Example 2 `go run main.go CLIENT localhost:8080 john`
5. To send a message to another client connected to the server, access the command line for a client and input a message in the form {to} {message}. NOTE: {from} will be the connected user by default.
    - Example `>> lewis john Hello Lewis!`
        - In this example, john sends a message to lewis saying "Hello Lewis!"
        - Lewis will receive a message that displays, 
            ```
            -  From:    john  

                Content: Hello Lewis!
            ```
6. To shut down the server and all clients, access the <b>server's</b> command line, and input 'EXIT'. Inputting 'EXIT' will close all active TCP connections and shut down the server.

# <ins>Documentation</ins>

## Product Requirements
- Server/Client architecture. One centralized server handles all connections and sending messages between clients. 
    - <b>Client</b>
        - Input: a host address, a port number, and a username
        - Client reads messages/commands from the user via the command line. 
        - Client sends messages to the server
        - Upon receiving the intended message from the server, display on the screen.
        - Upon receiving the termination signal from the server, the client should exit the program.
        - Upon receiving the “EXIT” command from the user, the client program exits.
    - <b>Server</b>
        - Input: a port number
        - Create one thread to handle the communication with each client.
        - Use channel to communicate between the client thread and the main server thread.
        - Upon receiving a message from a client C, check the To field, and deliver the message to the To-client; if the To-client is not connected, send C an error message.
        - The server should have a separate thread that waits for an “EXIT” command from the user via the command line (of the server program). Upon receiving such a command, the server exits the program and sends the termination signal to all the connected clients.

## Program Flow
The entry point to the program is main.go, where users can run either a server or a client by specifying which they would like to run as the first command-line argument. They then provide additional command-line arguments depending on whether they are running a client or a server. 

If SERVER is specified, main.go calls the `Server` function in pkg/server/server.go, which starts by logging the start time of the server and listening for TCP connections on the provided port. It then instantiates a channel for incoming messages from connected clients, and creates the router table which keeps a log of active connections. Then, `Server` starts a thread, `stopChatroom`, which allows for command-line input to exit the server and sever all client connections if the 'EXIT' message is entered. If the 'EXIT' message is recieved, `stopChatroom` calls `dispatchMulti`, which sends the exit signal to all connected clients. Then, `Server` initiates an infinite loop, which accepts incoming TCP connections and calls the `handleConnection` function.

The `handleconnection` function creates an encoder and decoder for each new ClientConnection, because each encoder and decoder takes the specific net.Conn for each client. The function then takes it's parameters, the newly created encoder and decoder, and creates a new ClientConnection struct. It then creates an empty Message struct, which takes data from the gob decoder. The username that's now in the message struct gets matched ClientConnection variable, they're added to the router table, with the username as the key, and the ClientConnection as the value. Finally, `handleConnection` starts a thread with the `receiveMessages` function.

The `receiveMessages` function reads data from the provided net.Conn, creating a new Message struct when there is command-line input recieved from a client. If the message is intended for the server, for example the instantiation message, or when a client disconnects, it deletes the corresponding client from the router table. If the message is a message from one client to another, `receiveMessages` calls `router.dispatch(*message)`.
`dispatch` checks if a message is sent to another client that is currently online, by cross referencing the username in the provided message with the list of keys in the Router map of `Username[ClientConnections]`. If the destination client is not online, `dispatch` writes an error message to the command-line of the sending client. If the destination client exists, the function encodes the message with the sending client's gob encoder, and sends it to the recieving client's decoder. 

If CLIENT is specified, main.go, calls the `Client(address, username)` function in pkg/client/client.go, which logs the connection time of the client and presents information regarding to the command-line parameters the client was started with. It then creates a TCP connection to the server, and sends an initializing message to the server, so the server learns about the client's existence and can add it to the router table. Each client creates two threads, one for receiving messages, `receiveMessages`, and one for catching an EXIT message, `catchSignalInterrupt`. The `Client` function then starts a loop that reads the command line for input, encoding messages and sending them to the server when there is a command-line input. 


## Design Choices 
- <b>Server</b>
    - <b>Structs</b>
        - <b>Message</b>
            - Contains a "From" string, a "To" string, and a "Content" string, to align with product requirements. This struct holds the data from a client's command line input.
        - <b>ClientConnection</b>
            - Contains a "address" string for the client's local address, a "c" net.conn connection for the respective client, and a gob encoder and gob decoder for that client. A ClientConnection struct is created by the server when a client connects to the server, it stores the requisite decoder and encoder to send and receive messages to/from the client from another client.
        - <b>Router</b>
            - Contains the "incoming" channel of Message defined in main, and a map of ClientConnection structs. The Router struct contains all ClientConnection structs created when a new client connects to the server. This struct serves as a central database for the server, it maps the "address" of a ClientConnection to the username provided when the client is initialized. It has a list of all clients and information specific to each client, so that sending messages from the incoming channel is as simple as matching the "To" field of the incoming message to an "address" contained within the map of ClientConnections. 
    - <b>Functions</b>
        - `stopChatroom(ch chan string)` - If the 'EXIT' message is received on the server command line, close connection to all clients and exit the program.
        - `receiveMessages(c net.Conn, router Router, enc *gob.Encoder, dec *gob.Decoder)` - Receive messages from a net.conn TCP connection, decodes the incoming "Message" struct using gob, creates a new ClientConnection struct which stores the client's connection info and data, then sends a pointer to the decoded message to router.dispatch.
        - `sendMessages(c net.Conn, outgoing chan Message)` - Takes a "Message" from the "outgoing" channel then checks which client the message should be sent to. If there's a matching client, the message is encoded using gob and is sent over the TCP connection to the intended client.
        - `handleConnection(c net.Conn, signal chan string, router Router)` - For each new client, adds a new ClientConnection struct to the map of ClientConnections in router. Then creates a Message struct, and instantiates the receiveMessages thread with same data as the ClientConnection variable. 
        - `(r Router) dispatch(m Message, c net.Conn)` - Takes a pointer to a Message struct and a net.Conn from the receiveMessages function. The function checks if a message is sent to another client that is currently online, by cross referencing the username in the provided message with the table of usernames within Router map of ClientConnections. 
        - `(r Router) dispatchMulti(content string)` - Loops over all usernames and ClientConnections in the router table, sends a message to all of them, specifically the exit message.
        - `Server(port string)` - The server is both the controller and the database in the architecture of our program. More specifically, the Router struct is the model, containing all the data and organization necessary for clients to communicate with one another. The functions within server.go control the flow of messages from one client to another, without non-participating clients being aware of any message sent between two participating clients. The `Server(port string)` function is the entry point into server.go, all other functions in server.go are subsequent calls from the initialization of the server and the threads it starts. 

- <b>Client</b>
    - <b>Structs</b>
        - <b>Message</b>
            - Contains a "From" string, a "To" string, and a "Content" string, to align with product requirements. This struct holds the data from a client's command line input.
    - <b>Functions</b>
        - `recieveMessages(c net.Conn)` - Launched as a thread that decodes messages received from the server and puts that information into a local message to be displayed on the command line. Also has the responsibility of shutting down clients if the server sends a 'SIGINT' message when it it closed by Ctrl+C. 
        - `readCommandLine(enc *gob.Encoder, username string)` - Starts an infinite loop that reads input from the command line. If the input is 'STOP', the client sends a message to the server to remove it from the Router table, then ends the function and the input loop. Otherwise, the client reads the command-line input, then sends it to the server to be passed to the destination client. Also handles some error checking.
        - `catchSignalInterrupt(c net.Conn, username string)` - Launched as a thread in `Client`, creates a channel that blocks until an os.Interrupt signal is recieved, if the signal is recieved then the channel unblocks and sends the message to the server which removes it from the Router table, then ends the function and closes the input loop.
        - `Client(address, username string)` - Client is the entry point to client.go, which is called in main.go if the initial command-line parameters specify launching a client. It creates a single TCP connection per-client to connect to the server, then starts the threads for `receiveMessages` and `catchSignalInterrupt`. It sends an initializing message to the server, so that the server can add this client's data to the Router table. It creates a gob encoder to read the input from the command line, then starts the command-line input loop to send and receive messages. 


## Justification

We had to create multiple gob encoders and decoders, one per ClientConnection struct in server.go. This is because gob.Encoder(net.Conn) takes a net.Conn, which is specific to each client connection. 
