# Сonductor
Conduror is a tool for remote management of docker containers. Condustor provides a mechanism to search for a container on several hosts and authenticated access to interact with containers 

## Contents
- [Conductor Capabilities](#conductor-capabilities)
- [Installation](#installation)
- [Add User](#add-user)

## Conductor Capabilities
Conductor provides the ability to remotely start and stop Docker containers. Several containers can be combined into one isolated network. Containers organized in one subnet are named Pods. Conductor automatically selects a free port on the host to open access from the outside. When starting a Pod, you can set its lifetime in hours. After this time, the Pod will be shut down and removed from the system.

Conductor has differentiated access rights. Currently, two roles with different access levels are supported.

User with administrator rights:
- Can start and stop Pods
- Can print all available Pods on the host
- Can view the status of a specific Pod
- Can view a list of all running Pods
- Can add Pods to the host
  
Common User:
- Can start and stop Pods
- Can print all available Pods on the host
- Can view the status of a specific Pod

 The Conductor-CLI tool has been developed for remote interaction with Conductor.

## Installation

Since Conductor is written in golang it can be run on most systems (Linux, Windows, Mac OS). To get started with Conductor, you will need to install Docker. The installation process may vary depending on your system. Check the official documentation for relevant information on the Docker installation process for your system.

After installing Docker, you can clone this repository and build the project.
```bash
git clone https://github.com/robocop4/Conductor.git
cd Conductor
go mod tidy
go build
```

To initially configure Conductor, run the main file. This creates a local database with the Conductor host configuration. The database will store your private key. Keep it secret to avoid compromising Conductor. A CID will also be generated at the first startup. The CID is a unique string by which Conductor-CLI can find the Conductor host. If there are several Conductor hosts in the network, the CID value must be the same for all of them. For this purpose, you can manually change the CID in the locale database (file conductor.db stored in project folder). 

The following snippet from the terminal demonstrates the launch of Conductor Host:

```bash
./main
My id:  QmYZSkbAA6VByCRDdJAQJ2kZLtAzkWHzENyygaocvVHAwu
My address:  [/ip4/127.0.0.1/tcp/41537 /ip4/192.168.88.196/tcp/41537]
My CID: 06Opjgjf06qLdzN
```

## Add User

To add a user for remote communication, you need to run Conductor with the following arguments:

```bash
// Add / Remove administrator
./main --admin --add <ID>
./main --admin --remove <ID>

// Add / Remove user
./main --user --add <ID>
./main --user --remove <ID>
```
ID is the user's identity generated on the basis of a private key. Each user must have its own private key. Within the same network there cannot be two simultaneous users with the same ID. Keep the user's private key secret. This key is used to authorize the user. To get the user's ID, run Conductor-CLI, which will generate a private key (if it is the first time the CLI has been started) and return the client's ID to the console:

```bash
go run . --cid <CID of Conductor Host>
My id:  QmfT81zosWxyHP5RkXnrecAtnLY1Y7ZZ1Yx1XCWsTfmSPD

```

Once a user is added, they can interact with the host.

