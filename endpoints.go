package main

import (
	"encoding/xml"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/libp2p/go-libp2p/core/network"
)

func unmarshalXML(xmlData []byte, v interface{}) error {
	err := xml.Unmarshal(xmlData, v)
	if err != nil {

		return err
	}
	return nil
}

func marshalXML(v interface{}, s network.Stream) {
	xmlData, err := xml.MarshalIndent(v, "", "  ")
	if err != nil {
		errorXML(err, s)
		return
	}
	// Sending the response back through the stream
	s.Write(xmlData)

	s.Close()
}

func errorXML(err error, s network.Stream) {

	log.Printf("%v", err)

	type Response struct {
		XMLName xml.Name `xml:"Response"`
		Status  int      `xml:"Status"`
	}

	response := Response{
		Status: 400,
	}

	xmlData, _ := xml.MarshalIndent(response, "", "  ")

	// Sending the response back through the stream
	_, _ = s.Write(xmlData)

	s.Close()
}

// Function prints all Pods available on the host
// Input:
// <Start>
//
//	<Action>all</Action>
//
// </Start>
//
// Response:
// <Response>
// <Pod>
// <PodName> </PodName>
// <Hash> </Hash>
// </Pod>
// </Response>
func ListXML(s network.Stream, body Action) {

	type Action struct {
		XMLName xml.Name `xml:"Root"`
		Action  string   `xml:"Root>Action"`
	}
	xmlWithRoot := fmt.Sprintf("<Root>%s</Root>", body.Content)
	var action Action

	err := unmarshalXML([]byte(xmlWithRoot), &action)
	if err != nil {
		errorXML(err, s)
		return
	}

	//Response
	response, err := SQLGetAllPods()
	if err != nil {
		errorXML(err, s)
		return
	}

	s.Write(response)
	s.Close()

}

// End point for starting the Pod
// Input:
// <Start>
//
//		<Hash>The hash that identifies Pod</Hash>
//		<UniqueId>Unique user ID</UniqueId>
//	 <Time>Pod's lifespan</Time>
//
// </Start>
//
// Response:
// <Response>
// <Status></Status> <- This is the processing status of the request.
// <Address></Address>
// </Response>
func RunXML(s network.Stream, body Action) {

	xmlWithRoot := fmt.Sprintf("<Root>%s</Root>", string(body.Content))
	type RunStruct struct {
		XMLName  xml.Name `xml:"Root"`
		Hash     string   `xml:"Hash"`
		UniqueId string   `xml:"UniqueId"`
		Time     string   `xml:"Time"`
	}

	var runXml RunStruct
	err := unmarshalXML([]byte(xmlWithRoot), &runXml)
	if err != nil {
		errorXML(err, s)
		return
	}

	port, err := VMStart(runXml.Hash, runXml.UniqueId, runXml.Time)
	if err != nil {
		errorXML(err, s)
		return
	}

	//Response

	type Response struct {
		XMLName xml.Name `xml:"Response"`
		Address string   `xml:"Address"`
		Status  int      `xml:"Status"`
	}

	response := Response{
		Address: fmt.Sprintf("%s:%d", globalIp, port),
		Status:  200,
	}

	marshalXML(response, s)

}

// The function stops the running pod
// Input:
// <Start>
//
//	<UniqueId>Unique user ID</UniqueId>
//
// </Start>
//
// Response:
// <Response>
// <Status></Status> <- This is the processing status of the request.
// </Response>
func StopXML(s network.Stream, body Action) {

	xmlWithRoot := fmt.Sprintf("<Root>%s</Root>", string(body.Content))
	type RunStruct struct {
		XMLName  xml.Name `xml:"Root"`
		UniqueId string   `xml:"UniqueId"`
	}
	var runXml RunStruct
	err := unmarshalXML([]byte(xmlWithRoot), &runXml)
	if err != nil {
		errorXML(err, s)
		return
	}

	err = VMstopByNetworkName(runXml.UniqueId)
	if err != nil {
		errorXML(err, s)
		return
	}

	//Response

	type Response struct {
		XMLName xml.Name `xml:"Response"`
		Status  int      `xml:"Status"`
	}

	response := Response{
		Status: 200,
	}

	marshalXML(response, s)
}

// Endpoint handler, returns information of the currently running Pod by unique user ID
// Input:
// <Status>
// <UniqueId>Unique user ID</UniqueId>
// </Status>
// Response:
// <Response>
// 	<Status>200</Status> <- This is the processing status of the request.
//  <Hash>Pod ID</Hash>
//  <Port> The port on which Pod is available</Port>
// </Response>

func StatusXML(s network.Stream, body Action) {
	xmlWithRoot := fmt.Sprintf("<Root>%s</Root>", string(body.Content))
	type StatusStruct struct {
		XMLName  xml.Name `xml:"Root"`
		UniqueId string   `xml:"UniqueId"`
	}

	var runXml StatusStruct

	err := unmarshalXML([]byte(xmlWithRoot), &runXml)
	if err != nil {
		errorXML(err, s)
		return
	}

	port, hash, err := VMstatus(runXml.UniqueId)
	if err != nil {
		errorXML(err, s)
		return
	}

	type Response struct {
		XMLName xml.Name `xml:"Response"`
		Status  int      `xml:"Status"`
		Hash    string   `xml:"Hash"`
		Port    string   `xml:"Port"`
	}

	response := Response{
		Status: 200,
		Hash:   hash,
		Port:   port,
	}

	marshalXML(response, s)

}

// End point of printing of already started Pods
// Input:
// <Running>
//
// </Running>
// Response:
// <Response>
//
//		<Status>200</Status> <- This is the processing status of the request.
//		<Running>
//			<User ID>Container name</User ID>
//			<User ID>Container name</User ID>
//	 <Running>
//
// </Response>
func RunningXML(s network.Stream, body Action) {

	rMap := make(map[string][]string)
	containers, err := VMgetRunningPods()
	if err != nil {
		errorXML(err, s)
		return
	}
	for _, container := range containers {
		if len(container.Names) > 0 {
			last := strings.Split(container.Names[0], "-")
			uStr := last[len(last)-1]
			rMap[uStr] = append(rMap[uStr], container.Names[0])
		}
	}

	type Item struct {
		XMLName xml.Name
		Value   string `xml:",chardata"`
	}

	type Running struct {
		XMLName xml.Name `xml:"Running"`
		Items   []Item   `xml:",any"`
	}

	type Response struct {
		XMLName xml.Name `xml:"Response"`
		Status  int      `xml:"Status"`
		Running Running  `xml:"Running"`
	}

	var items []Item
	for key, value := range rMap {

		items = append(items, Item{
			XMLName: xml.Name{Local: key},
			Value:   strings.Join(value, " "),
		})
	}

	// Response

	response := Response{
		Status: 200,
		Running: Running{
			Items: items,
		},
	}

	marshalXML(response, s)

}

func AuthXML(s network.Stream, body Action) {

	//Check user role in the database
	role, _ := SQLcheckRole(s.Conn().RemotePeer().String())
	//Get the full list of user privileges
	perm := CheckPermission(RBAC, role)

	type Response struct {
		XMLName     xml.Name `xml:"Response"`
		Permissions []string `xml:"Permissions>Permission"`
		Status      int      `xml:"Status"`
	}

	response := Response{
		Permissions: perm,
		Status:      200,
	}

	marshalXML(response, s)

}

// The function adds information about the Pod to the database.
// All images must be loaded manually before calling this function.

func AddXML(s network.Stream, body Action) {

	xmlWithRoot := fmt.Sprintf("<Pod>%s</Pod>", string(body.Content))

	var addXml Pod

	err := unmarshalXML([]byte(xmlWithRoot), &addXml)
	if err != nil {
		errorXML(err, s)
		return
	}

	if addXml.InternalPort < 0 || addXml.InternalPort > 1023 {
		errorXML(errors.New("The internal port does not fall within the range 0-1023."), s)
		return
	}

	//

	for _, img := range addXml.Images {
		check, err := VMcheckImageExist(img)
		if err != nil {
			errorXML(err, s)
			return
		}
		if !check {
			errorXML(errors.New(fmt.Sprintf("%s  image not found", img)), s)
			return
		}

	}

	err = VMCreate(addXml)
	if err != nil {
		errorXML(err, s)
		return
	}
	// response
	type Response struct {
		XMLName xml.Name `xml:"Response"`
		Status  int      `xml:"Status"`
	}

	response := Response{
		Status: 200,
	}

	marshalXML(response, s)

}
