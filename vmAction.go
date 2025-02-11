package main

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	containertypes "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"

	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

// checkPort checks if a port is available on the given host
func checkPort(host string, port int) bool {
	// Create the address for connection
	address := fmt.Sprintf("%s:%d", host, port)

	// Try to connect to the port within 1 second
	conn, err := net.DialTimeout("tcp", address, time.Second)
	if err != nil {
		return false // Port is closed or unreachable
	}
	defer conn.Close() // Close the connection if the port is open
	return true        // Port is open
}

func contains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

type Pod struct {
	XMLName       xml.Name `xml:"Pod"`           // Root element
	PodName       string   `xml:"PodName"`       // Pod name
	Images        []string `xml:"Images>Image"`  // Array of images
	ExternalImage string   `xml:"ExternalImage"` // Image that is accessible externally
	Metadata      []string `xml:"Metadata>Item"` // Array of metadata items
	InternalPort  int      `xml:"InternalPort"`  // Internal port
}

func VMCreate(pod Pod) error {

	// Sort arrays
	sort.Strings(pod.Metadata)
	sort.Strings(pod.Images)

	// check that the external container is in the list of all images
	if !contains(pod.Images, pod.ExternalImage) {
		return errors.New("ExternalImage is not contained in Image array")
	}
	//TODO: to improve the hashing system. The hash of the image itself should be taken. This will minimize conflict situations in case of use on many hosts
	img := strings.Join(pod.Images, ", ")

	hash := StringToSHA256(fmt.Sprintf("%d,%s,%s,%s,%s", pod.InternalPort, img, strings.Join(pod.Metadata, ", "), pod.PodName, pod.ExternalImage))

	err := addPod(pod.PodName, pod.InternalPort, pod.Images, pod.Metadata, hash, pod.ExternalImage)

	return err

}

func VMgetRunningPods() ([]types.Container, error) {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatalf("Error creating Docker client: %v", err)
	}
	defer cli.Close()

	// Get the list of running containers
	containers, err := cli.ContainerList(ctx, containertypes.ListOptions{})
	if err != nil {
		log.Fatalf("Error listing containers: %v", err)
	}

	return containers, err
}

// The function deletes all running Pods and all associated resources
// Deletion is performed via the network identifier
func VMstopByNetworkName(networkName string) error {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return fmt.Errorf("VMstopByNetworkName>client.NewClientWithOpts: %w", err)
	}
	defer cli.Close()

	filterArgs := filters.NewArgs()
	targetLabelKey := "UniqueID"
	targetLabelValue := networkName
	filterArgs.Add("label", fmt.Sprintf("%s=%s", targetLabelKey, targetLabelValue))

	containers, err := cli.ContainerList(ctx, containertypes.ListOptions{All: true, Filters: filterArgs})
	if err != nil {
		return fmt.Errorf("VMstopByNetworkName>cli.ContainerList: %w", err)
	}

	// Выводим информацию о контейнерах
	for _, container := range containers {

		err := cli.NetworkDisconnect(ctx, networkName, container.ID, true)
		if err != nil {
			return fmt.Errorf("VMstopByNetworkName>cli.NetworkDisconnect: %w", err)
		}

		// Removing a container
		err = cli.ContainerRemove(ctx, container.ID, containertypes.RemoveOptions{
			RemoveVolumes: true,
			//	RemoveLinks:   true,
			Force: true,
		})
		if err != nil {
			return fmt.Errorf("VMstopByNetworkName>cli.ContainerRemove: %w", err)
		}

	}

	// Removing the network
	// No error handling is needed here
	cli.NetworkRemove(ctx, networkName)
	// if err != nil {
	// 	//TOOD:
	// 	//return fmt.Errorf("cli.NetworkRemove: %w", err)
	// }
	return nil
}

// This function starts a Pod, which is identified by a unique hash.
// To allow multiple users to run the same pod at the same time, a unique identifier is added. This can be the user's identifier.
// The uniqueness of the identifier is not checked on the Pilot side. The client calling this function guarantees the uniqueness of the identifier.
// This function generates a random port in the range of 1000 to 9999 and checks it for availability.
// Information about the requested pod is taken from the database. This information is used to configure the Pod.
// If the execution of all procedures is successful, the function will return the port on which the running pod is available.
func VMStart(hash string, UniqueId string) (int, error) {

	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return 0, fmt.Errorf("VMStart>client.NewClientWithOpts error: %s", err.Error())
	}
	defer cli.Close()

	go VMstopOverdue(3)
	//The second step is to stop and delete the containers of the same user
	//TODO: It's a labor-intensive mechanism. It can be improved
	err = VMstopByNetworkName(UniqueId)
	if err != nil {
		return 0, fmt.Errorf("VMStart>VMstopByNetworkName error: %s", err.Error())
	}

	uniquePort := 0

	//	 Getting information on the pod
	podData, err := getPods(hash)
	if err != nil {
		return 0, fmt.Errorf("VMStart>GetPods error: %s", err.Error())
	}

	//
	// Create a virtual network for our Pod
	// Define labels for the network
	labels := map[string]string{
		"uId":  UniqueId,
		"time": fmt.Sprintf("%d", time.Now().Unix()),
		"Hash": hash,
	}

	networkName := UniqueId
	_, err = cli.NetworkCreate(ctx, networkName, types.NetworkCreate{
		Driver: "bridge",
		Labels: labels,
	})

	if err != nil {

		// if strings.Contains(err.Error(), "already exists") {
		// 	err := VMstopByNetworkName(networkName)
		// 	if err != nil {
		// 		return 0, fmt.Errorf("VMstopByNetworkName: %s", err.Error())
		// 	}

		// } else {
		return 0, fmt.Errorf("VMStart>cli.NetworkCreate error: %s", err.Error())
		//}
	}

	// If the virtual network is created, bring up Struchek

	// Network Configuration
	networkConfig := &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			networkName: {
				NetworkID: networkName,
			},
		},
	}

	// Assign environment variables that contain the names of containers for communicating with each other
	// All containers within the same pod see each other
	envVars := []string{}
	for _, img := range podData.Images {
		envVars = append(envVars, fmt.Sprintf("%s=%s-%s", img, img, UniqueId))
	}

	// Go through all containers to run them on the same subnet.
	// One pod has one container that is accessible from the outsid
	for _, img := range podData.Images {

		// Check if the container is external
		// If the container is external we need to forward a random port to it
		var config *container.Config
		var hostConfig *container.HostConfig
		if img == podData.ExternalImage {

			// To minimize the probability of a race, random port generation is implemented here
			for {
				// Generate a unique free port
				rand.Seed(time.Now().UnixNano())
				uniquePort = rand.Intn(9000) + 1000             // random number generation from 1000 to 9999
				portExist := checkPort("127.0.0.1", uniquePort) // Port check

				if !portExist {
					// If the port is free, exit the loop
					break
				}
			}

			// Container configuration
			config = &container.Config{
				Image: img, // Specify the name of the container to run
				Labels: map[string]string{
					"UniqueID": UniqueId,
					"time":     fmt.Sprintf("%d", time.Now().Unix()), // Time is used to track the life of the container. This allows you to limit the lifetime of the container if necessary.
					"port":     fmt.Sprintf("%d", uniquePort),
				},
				Env: envVars,
				ExposedPorts: nat.PortSet{
					"80/tcp": struct{}{},
				},
			}

			// Host configuration with port forwarding
			hostConfig = &container.HostConfig{
				PortBindings: nat.PortMap{
					"80/tcp": []nat.PortBinding{
						{
							HostPort: fmt.Sprintf("%d", uniquePort),
						},
					},
				},
			}

		} else {
			// Container configuration
			config = &container.Config{
				Image: img, // Specify the name of the container to run
				Labels: map[string]string{
					"UniqueID": UniqueId,
					"time":     fmt.Sprintf("%d", time.Now().Unix()), //Time is used to track the life of the container. This allows you to limit the lifetime of the container if necessary.
				},
				Env: envVars,
			}

			// Host configuration without port forwarding
			hostConfig = &container.HostConfig{
				PortBindings: nat.PortMap{},
			}
		}

		//Creating the container
		resp, err := cli.ContainerCreate(ctx, config, hostConfig, networkConfig, nil, fmt.Sprintf("%s-%s", img, UniqueId))
		if err != nil {
			return 0, fmt.Errorf("VMStart>cli.ContainerCreate error: %s", err.Error())
		}

		if err = cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
			return 0, fmt.Errorf("VMStart>cli.ContainerStart error: %s", err.Error())
		} else {
			// fmt.Println("Started container:", resp.ID)
			//fmt.Sprintf("http://%s:%d", "globalIp", 8080), nil
		}

	}

	return uniquePort, nil
}

// TODO: This feature is not implemented in the protocol
// The function imports tar image into the system
// The image file is taken from the /tmp folder
func VMimportImage(image string) error {

	absPath := filepath.Clean(image)
	baseDir := "/tmp/"
	absPath = filepath.Join(baseDir, absPath)

	if !filepath.HasPrefix(absPath, baseDir) {
		return errors.New("VMimportImage>filepath.HasPrefix>Bad file name")
	}

	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return fmt.Errorf("VMimportImage>client.NewClientWithOpts error: %s", err.Error())
	}

	// Opening the image archive
	imageFile, err := os.Open(absPath) // Specify the path to your file
	if err != nil {
		return fmt.Errorf("VMimportImage>os.Open error: %s", err.Error())
	}
	defer imageFile.Close()

	_, err = cli.ImageLoad(context.Background(), imageFile, false)
	if err != nil {
		return fmt.Errorf("VMimportImage>cli.ImageLoad error: %s", err.Error())
	}

	return nil
}

func VMstatus(networkName string) (string, string, error) {

	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return "", "", fmt.Errorf("VMstatus>client.NewClientWithOpts error: %s", err.Error())
	}
	networkInspect, err := cli.NetworkInspect(ctx, networkName, types.NetworkInspectOptions{})
	if err != nil {
		return "", "", fmt.Errorf("VMstatus>cli.NetworkInspect error: %s", err.Error())
	}

	hash := networkInspect.Labels["Hash"]
	portR := "0"
	for _, endpoint := range networkInspect.Containers {

		containerID := endpoint.Name

		containerInspect, err := cli.ContainerInspect(ctx, containerID)
		if err != nil {
			return "", "", fmt.Errorf("VMstatus>cli.ContainerInspect error: %s", err.Error())

		}

		for _, port := range containerInspect.NetworkSettings.Ports {
			for _, binding := range port {
				portR = binding.HostPort
			}
		}
	}
	return portR, hash, nil

}

func VMcheckImageExist(imageName string) (bool, error) {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return false, fmt.Errorf("VMcheckImageExist>client.NewClientWithOpts error: %s", err.Error())
	}

	filter := filters.NewArgs()
	filter.Add("reference", imageName)

	images, err := cli.ImageList(ctx, image.ListOptions{
		Filters: filter,
	})
	if err != nil {
		return false, fmt.Errorf("VMcheckImageExist>cli.ImageList error: %s", err.Error())
	}

	return len(images) > 0, nil

}

func VMstopOverdue(hour int) error {
	// Инициализация клиента Docker
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return err
	}

	// Получаем список сетей
	networks, err := cli.NetworkList(context.Background(), types.NetworkListOptions{})
	if err != nil {
		return err
	}

	for _, network := range networks {

		if timeStr, exists := network.Labels["time"]; exists {

			unixtime, err := strconv.ParseInt(timeStr, 10, 64)
			if err != nil {
				return err
			}
			t := time.Unix(unixtime, 0)
			newTime := t.Add(time.Duration(hour) * time.Hour)
			now := time.Now().Unix()
			if now >= newTime.Unix() {
				VMstopByNetworkName(network.Name)
			}

		}
	}
	return nil
}
