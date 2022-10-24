Goal: a simple chat room application that supports only private message

#### Private Message Server and Client

# <ins>Instructions</ins>
1. Clone the privateChat repository to your machine
2. Open as many terminals as you want clients, and one additional terminal to run the server
3. Start by running the server, providing a port number for the server
    - `go run main.go SERVER {port}`
    - Example: `go run main.go SERVER 8080`
4. Launch as many clients as you want, providing command-line arguments in the form {host}:{port} {username}, where the port is the port previously provided to launch the server. If running locally, the {host} argument, will always be 'localhost'. 
    - `go run main.go CLIENT {host}:{port} {username}`
    - Example 1 `go run main.go CLIENT localhost:8080 lewis`
    - Example 2 `go run main.go CLIENT localhost:8080 john`
5. To send a message to another client connected to the server, access the command line for a client and input a message in the form {to} {from} {message}
    - Example `>> lewis john Hello Lewis!`
        - In this example, John sends a message to lewis saying "Hello Lewis!"
        - Lewis will receive a message that displays, 
            -  `From:    john 
                Content: Hello Lewis!`
6. To shut down the server and all clients, access the <b>server's</b> command line, and input 'EXIT'. Inputting 'EXIT' will close all active TCP connections and shut down the server.

#### <ins>Documentation</ins>

# Product Requirements
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

# Program Flow
    The entry point to the program is main.go, where users can run either a server or a client by specifying which they would like to run as the first command-line argument. They then provide additional command-line arguments depending on whether they are running a client or a server. 
    If SERVER is specified, main.go calls the Server(port string) function in pkg/server/server.go, which starts by logging the start time of the server and listening for TCP connections on the provided port. It then instantiates a channel for incoming messages from connected clients, and creates the router table which keeps a log of active connections. Then, Server(port string) starts a thread, stopChatroom(*r Router), which allows for command-line input to exit the server and sever all client connections if the 'EXIT' message is entered. Then, Server(port string) initiates an infinite loop, which accepts incoming TCP connections and calls the handleConnection(c net.Conn, r Router) function ... <b> NOT DONE, TALK ABOUT handleConnection() </b>
    If running a client, main.go, calls the Client(address, username) function in pkg/client/client.go, which logs the connection time of the client and presents information regarding to the command-line parameters the client was started with. It then creates a TCP connection to the server, and sends an initializing message to the server, so the server learns about the client's existence and can add it to the router table. Each client creates two threads, one for receiving messages, receiveMessages(), and one for catching an EXIT message, catchSignalInterrupt(). The Client(address, username) function then starts a loop that reads the command line for input, encoding messages and sending them to the server when there is a command-line input. 


# Design Choices 
- <b>Server</b>
    - <b>Structs</b>
        - <b>Message</b>
            - Contains a "From" string, a "To" string, and a "Message" string, to align with product requirements. This struct holds the data from a client's command line input. Message structs are used in the "incoming" and "outgoing" channels in main.go, which are passed to various functions that need access to command-line input.
        - <b>ClientConnection</b>
            - Contains a "address" string for the client's local address, a "c" net.conn connection for the respective client, and a gob encoder and gob decoder for that client. A ClientConnection struct is created by the server when a client connects to the server, it stores the requisite decoder and encoder to send and receive messages to/from the client from another client.
        - <b>Router</b>
            - Contains the "incoming" channel of Message defined in main, and a map of ClientConnection structs. The Router struct contains all ClientConnection structs created when a new client connects to the server. This struct serves as a central database for the server, it maps the "address" of a ClientConnection to the username provided when the client is initialized. It has a list of all clients and information specific to each client, so that sending messages from the incoming channel is as simple as matching the "To" field of the incoming message to an "address" contained within the map of ClientConnections. 
    - <b>Functions</b>
        - `stopChatroom(ch chan string)` - If the 'EXIT' message is received on the server command line, close connection to all clients and exit the program.
        - `receiveMessages(c net.Conn, router Router, enc *gob.Encoder, dec *gob.Decoder)` - Receive messages from a net.conn TCP connection, decodes the incoming "Message" struct using gob, creates a new ClientConnection struct which stores the client's connection info and data, then sends a pointer to the decoded message to router.dispatch.
        - `sendMessages(c net.Conn, outgoing chan Message)` - Takes a "Message" from the "outgoing" channel then checks which client the message should be sent to. If there's a matching client, the message is encoded using gob and is sent over the TCP connection to the intended client.
        - `handleConnection(c net.Conn, signal chan string, router Router)` - For each new client, adds a new ClientConnection struct to the map of ClientConnections in router. Then creates a Message struct, and instantiates the receiveMessages thread with the 
        - `(r Router) dispatch(m Message, c net.Conn)` - Takes a pointer to a Message struct and a net.Conn from the receiveMessages function. The function checks if a message is sent to another client that is currently online, by cross referencing the username in the provided message with the table of usernames within Router map of ClientConnectoins. 
        - `Server(port string)` - <b>TODO when program flow has been finalized</b>
    - The server is both the controller and the database in the architecture of our program. More specifically, the Router struct is the model, containing all the data and organization necessary for clients to communicate with one another. The functions within server.go control the flow of messages from one client to another, without non-participating clients being aware of any message sent between two participating clients. 

- <b>Client</b>
    - 


