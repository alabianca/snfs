# SNFS
With SNFS you can create a peer to peer network which enables you to share content with your peers
in a fully peer to peer fashion. SNFS uses the Kademlia protocol as a routing mechanism.
SNFS currently consists of a backend server and a frontend cli. The cli and server communicate
using a REST api on localhost. I plan on adding a desktop UI in the future.

## Dependencies
- gokad: The Kademlia DHT Implementation https://github.com/alabianca/gokad
- kadnet: The Kademlia Node Implementation https://github.com/alabianca/kadnet

## Installation
I plan on improving this, but currently you must clone this repo along the 2 dependencies.
Once cloned use `make` to install the cli and and server in your `GOBIN`.
You know these two applications are installed if you can run `snfs` and `snfsd` in your command line.

## Environment Variables
There are a few environment variables that can be set for the backend server

| Env                         | Description                                | Default |
|-----------------------------|:------------------------------------------:|--------:|
|SNFS_CLIENT_CONNECTIVITY_PORT|The port to which client apps connect (cli) | 4200    |
|SNFS_DISCOVERY_PORT          |The port that is discoverable by other Nodes| 5050    |
|SNFS_FS_PORT                 |Content is published at this port           |         |


## Usage
Once you have set the above dependencies you can start the server using `snfsd`.
Use the cli to become discoverable by other nodes. Type `snfs up <your_name>`.
This will spin up the Kademlia Node and will actively start responding to Kademlia RPCs.
To stop you can always do `snfs down` and you will stop responding to RPCs.

Once you are up and running you have 2 choices. 
1. Start your own network
2. Join and existing network

##### Option 1. Start your own network
You basically started your own network once you used the `snfs up` command. Other nodes may wish
to join your network using [Option 2](#option-2-join-an-existing-network)

##### Option 2. Join an existing network
To join an existing network your node must be up and running. You also need the IP and Port of
a `bootstrap` node. The Port will be the `SNFS_DISCOVERY_PORT` of the bootstrap node.
To bootstrap use the `snfs bootstrap <port> <ip>` command. It may take a while depending on the size
of the network you are joining. At this point you are discovering the closest nodes to you while
simultaneously announcing your presence in the network.
Once bootstrapped, your routing table should be filled with nodes close to you. Check by running `snfs kad status`.


Once you are part of a network you can store files in the network using the `snfs share <context>` command.
The `<context>` argument is either the file/folder name or `.` if you want to share the content of your current working directory.

The backend server will return a hash. The hash acts like a url. If another node wants to access your shared content, it has to use that hash
in the `snfs clone <hash>` command. 

## Limitations
The currently largest limitation is that it only works within a local network due to the fact that
most personal computers sit behind a NAT. I plan to get around that by using some sort of UDP and TCP 
hole punching method.

## Todos
- [ ] Serialize/Deserialize RoutingTable on exit/startup
- [ ] Go beyond a local network using some sort of NAT hole punching for UDP and TCP
- [ ] When `snfs down` announce it to the network
- [ ] Properly handle full KBuckets according to the Kademlia Spec
- [ ] Desktop UI
- [ ] Periodic republishing according to the Kademlia Spec
- [ ] Bring in Travis CI

