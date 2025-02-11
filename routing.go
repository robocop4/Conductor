package main

import (
	"encoding/xml"
	"strings"

	"github.com/libp2p/go-libp2p/core/network"
)

// The port on which Pod is available
type Router struct {
	routes map[string]func(s network.Stream, body Action)
}

func NewRouter() *Router {
	return &Router{routes: make(map[string]func(s network.Stream, body Action))}
}

// Route handler registration
func (r *Router) HandleFunc(route string, handler func(s network.Stream, body Action)) {
	r.routes[route] = handler
}

// Handler for threads
func streamHandler(router *Router) func(s network.Stream) {
	return func(s network.Stream) {

		buf := make([]byte, 1024)
		n, err := s.Read(buf)
		if err != nil {
			errorXML(err, s)
			return
		}

		var root xml.Name
		decoder := xml.NewDecoder(strings.NewReader(string(buf[:n])))

		// Getting a token for XML parsing
		token, err := decoder.Token()
		if err != nil {
			errorXML(err, s)
			return
		}

		// The initial token must be the beginning of an XML element
		switch t := token.(type) {
		case xml.StartElement:
			root = t.Name
		}

		//  XML to Action structure
		var action Action
		err = xml.Unmarshal(buf[:n], &action)
		if err != nil {
			errorXML(err, s)
			return
		}

		// permission check
		role, err := SQLcheckRole(s.Conn().RemotePeer().String())
		if err != nil {
			errorXML(err, s)
			return
		}

		perm := ChackRole(RBAC[root.Local], role)
		// the user has no rights to call the function
		if !perm {

			type Response struct {
				XMLName string `xml:"Response"`
				Status  int    `xml:"Status"`
			}

			resp := Response{
				XMLName: "Response",
				Status:  500,
			}

			output, _ := xml.MarshalIndent(resp, "", "  ")

			// Sending the response back through the stream
			s.Write(output)

			// Closing the stream
			s.Close()
			return
		}

		// Routing the request depending on the root element
		if handler, ok := router.routes[root.Local]; ok {
			handler(s, action) // Call the handler for this route
		} else {
			errorXML(err, s)
		}
	}
}
