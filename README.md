


<p align="center"><img  src="https://raw.githubusercontent.com/robocop4/Conductor/refs/heads/main/logo.png"  height="200"/></p>
<h1 align="center">Сonductor</h1>


Conduror is a tool for remote management of docker containers. Condustor provides a mechanism to search for a container on several hosts and authenticated access to interact with containers 

## Contents
- [Conductor Capabilities](#conductor-capabilities)
- [Installation](#installation)
- [Add User](#add-user)
- [Establishing a Connection](#establishing-a-connection)
- [Creating And Managing Pods](#creating-and-managing-pods)
- [Communications Between Containers Within a Single Pod](#communications-between-containers-within-a-single-pod)

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

 The [Conductor-CLI](https://github.com/robocop4/Conductor_CLI) tool has been developed for remote interaction with Conductor.

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
// Add | Remove administrator
./main --admin --add <ID>
./main --admin --remove <ID>

// Add | Remove user
./main --user --add <ID>
./main --user --remove <ID>
```
ID is the user's identity generated on the basis of a private key. Each user must have its own private key. Within the same network there cannot be two simultaneous users with the same ID. Keep the user's private key secret. This key is used to authorize the user. To get the user's ID, run Conductor-CLI, which will generate a private key (if it is the first time the CLI has been started) and return the client's ID to the console:

```bash
go run . --cid <CID of Conductor Host>
My id:  QmfT81zosWxyHP5RkXnrecAtnLY1Y7ZZ1Yx1XCWsTfmSPD

```

Once a user is added, they can interact with the host.

## Establishing a Connection

Use [Conductor-CLI](https://github.com/robocop4/Conductor_CLI) for remote interaction with Conductor. To establish a connection, run the CLI with the --cid <unique identifier> switch and wait for the connection to be established, this may take a few minutes. During this time Conductor will discover all hosts with the specified identifier. Use the providers command to print out a list of all detected hosts as shown in the following terminal snippet:

```bash
./cli --cid 06Opjgjf06qLdzN
My id:  QmfT81zosWxyHP5RkXnrecAtnLY1Y7ZZ1Yx1XCWsTfmSPD
My address:  [/ip4/127.0.0.1/tcp/34963 /ip4/192.168.88.196/tcp/34963]
>providers
0 {QmYZSkbAA6VByCRDdJAQJ2kZLtAzkWHzENyygaocvVHAwu: [/ip4/<ip>/tcp/41537]}
```

Use the coand `use <int>` to select the host you want to communicate with. After selecting the host, you will receive a response from Conductor that will list the allowed commands for your user as shown in the following terminal snippet:

```bash
>providers
0 {QmYZSkbAA6VByCRDdJAQJ2kZLtAzkWHzENyygaocvVHAwu: [/ip4/178.165.68.120/tcp/41537]}
>use 0
Received response: <Response>
  <Permissions>
    <Permission>Add</Permission>
    <Permission>Auth</Permission>
    <Permission>Start</Permission>
    <Permission>Stop</Permission>
    <Permission>List</Permission>
    <Permission>Status</Permission>
    <Permission>Running</Permission>
  </Permissions>
  <Status>200</Status>
</Response>
QmYZSkbAA6VByCRDdJAQJ2kZLtAzkWHzENyygaocvVHAwu>
```

If you only see `<Permission>Auth</Permission>` permissions then you have not added your user and cannot interact with the host.

## Creating And Managing Pods


You must have administrator privileges to create a pod. The following command will create a Pod on the system:
```bash
QmYZSkbAA6VByCRDdJAQJ2kZLtAzkWHzENyygaocvVHAwu>add <Pod Name> <Port> <Img1,Img2> <Main IMG> <Metadata,Metadata>
Received response: <Response>
  <Status>200</Status>
</Response>
```

- `<Pod Name>` is the name of the Pod being created. Within a single host, this value must be unique. 
- `<Port>` is the port that will be forwarded from the Pod's virtual network to the outside. 
- `<Img1,Img2>` is a comma separated list of Docker images that will be launched when the  Pod starts. These images must be added to the system in advance via the Docker CLI. 
- `<Main IMG>` is the image that will look outward. The name of this image must be in the list from the previous agrument.
- `<Metadata,Metadata>` is any comma separated data that you want to add to the Pod. This can be used to comment on the Pod. 

A response with status `200` means that the command was successful and you can now view the list of Pods in the system via the `list` command:

```bash
QmYZSkbAA6VByCRDdJAQJ2kZLtAzkWHzENyygaocvVHAwu>list
Received response: <Response>
  <Pod>
    <PodName>Rododedron</PodName>
    <Hash>b4786b837c88411c9b3bb09275e0ad28768538f7b50172d5483673e9f6369b88</Hash>
  </Pod>
  <Pod>
    <PodName>PodTest</PodName>
    <Hash>c977ea9d35cc19738ab1230335e86920d5f1f597fbf19bac74db92d596add66c</Hash>
  </Pod>
</Response>
```

Pods are identified in the system through hash values. This value must be unique within a single host. Use the `run` command to start the Pod as shown in the following snippet:

```bash
QmYZSkbAA6VByCRDdJAQJ2kZLtAzkWHzENyygaocvVHAwu>run c977ea9d35cc19738ab1230335e86920d5f1f597fbf19bac74db92d596add66c AnyString 1
Received response: <Response>
  <Address>IP:9669</Address>
  <Status>200</Status>
</Response>
```

In this example, a Pod with the identifier `c977ea9d35cc19738ab1230335e86920d5f1f597fbf19bac74db92d596add66c` was started and the identifier `AnyString` was assigned to it.The lifetime of the Pod is one hour.
Only one Pod with the 'AnyString' identifier can be running on a single host. If you repeat the above command, the old Pod will be stopped and deleted and a new Pod will be started instead. To check the status of a pod by its ID, use the `status` command as shown in the following example:

```bash
QmYZSkbAA6VByCRDdJAQJ2kZLtAzkWHzENyygaocvVHAwu>status AnyString
Received response: <Response>
  <Status>200</Status>
  <Hash>c977ea9d35cc19738ab1230335e86920d5f1f597fbf19bac74db92d596add66c</Hash>
  <Port>9669</Port>
</Response>
```

To stop the Pod, you can wait until one hour has elapsed or use the `stop` command. To view all running Pods, use the `running` command:

```bash
QmYZSkbAA6VByCRDdJAQJ2kZLtAzkWHzENyygaocvVHAwu>running
Received response: <Response>
  <Status>200</Status>
  <Running>
    <AnyString2>/test2-AnyString2 /test-AnyString2</AnyString2>
    <AnyString>/test2-AnyString /test-AnyString</AnyString>
  </Running>
</Response>
QmYZSkbAA6VByCRDdJAQJ2kZLtAzkWHzENyygaocvVHAwu>stop AnyString
Received response: <Response>
  <Status>200</Status>
</Response>`
```

## Communications Between Containers Within a Single Pod

All containers within a single Pod are bounded by a virtual network and can communicate with each other. As an example, suppose that Pod contains two containers and we need to send an HTTP request from container `test` to container `test2`. It is enough to use the name of the second container as url as shown in the following fragment from the terminal:

```bash
root@496e01a61974:/# curl $test2
<!DOCTYPE html>
<html>
<head>
<title>Welcome to nginx!</title>
<style>
html { color-scheme: light dark; }
body { width: 35em; margin: 0 auto;
font-family: Tahoma, Verdana, Arial, sans-serif; }
</style>
</head>
<body>
<h1>Welcome to nginx!</h1>
<p>If you see this page, the nginx web server is successfully installed and
working. Further configuration is required.</p>

<p>For online documentation and support please refer to
<a href="http://nginx.org/">nginx.org</a>.<br/>
Commercial support is available at
<a href="http://nginx.com/">nginx.com</a>.</p>

<p><em>Thank you for using nginx.</em></p>
</body>
</html>
```

The `test2` in this case is the name of the image.
