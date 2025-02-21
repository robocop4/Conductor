package main

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/ipfs/go-cid"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/net/connmgr"
	"github.com/multiformats/go-multiaddr"
)

var globalIp string

var db *sql.DB

func StringToSHA256(input string) string {
	hash := sha256.New()
	hash.Write([]byte(input))
	hashBytes := hash.Sum(nil)
	hashString := hex.EncodeToString(hashBytes)
	return hashString
}

func getGlobalIP() string {
	type Response struct {
		IP string `json:"ip"`
	}
	res, err := http.Get("https://ifconfig.co")
	if err != nil {
		fmt.Println("getGlobalIP()>ifconfig.co error:", err.Error())
		res, err := http.Get("https://api.ipify.org/?format=json")
		if err != nil {
			fmt.Println("getGlobalIP()>api.ipify.org error:", err.Error())
			return ""
		}
		resBody, err2 := ioutil.ReadAll(res.Body)
		if err2 != nil {
			fmt.Println("getGlobalIP>ioutil.ReadAll error:", err.Error())
			return ""
		}
		var response Response
		err = json.Unmarshal(resBody, &response)
		if err != nil {
			fmt.Println("getGlobalIP>json.Unmarshal error:", err.Error())
			return ""
		}

		return response.IP

	}

	resBody, err2 := ioutil.ReadAll(res.Body)
	if err2 != nil {
		fmt.Printf("client: could not read response body: %s\n", err)
		return ""
	}
	return strings.TrimSpace(string(resBody))
}

// Функция для создания нового Стручка на этой ноде.
// Функция принимает XML файл с описанием Стручка
// Пример XML разметки
// <Pod>
//     <PodName>example-pod</PodName> // Имя по которому будет доступен POD. Должно быть уникальным.
//     <Images> //Массив образов связанных с этим Стручком
//         <Image>image1:latest</Image>
//         <Image>image2:v1</Image>
//         <Image>image3:v2</Image>
//     </Images>
//     <InternalPort>80</InternalPort> //Внутренний порт по умолчанию 80
//     <Metadata>
//		<Item>Text</Item>
//      <Item>Text</Item>
//	   </Metadata>
// </Pod>

// Connection manager to limit connections
var connMgr, _ = connmgr.NewConnManager(1, 2, connmgr.WithGracePeriod(time.Minute))

// Extra options for libp2p
var Libp2pOptionsExtra = []libp2p.Option{
	libp2p.NATPortMap(),
	libp2p.ConnectionManager(connMgr),
	//libp2p.EnableAutoRelay(),
	libp2p.EnableNATService(),
}

func handleConnection(net network.Network, conn network.Conn) {

	// Here you can reject the connection based on the blacklist

}

type Action struct {
	Content string `xml:",innerxml"`
}

func main() {

	// err1 := VMstopOverdue(1)
	// fmt.Println(err1)

	// return
	_, err := initDB()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	var adminFlag bool
	var userFlag bool
	var addFlag string
	var removeFlag string
	var listFlag bool

	flag.BoolVar(&adminFlag, "admin", false, "Administrator operation.")
	flag.BoolVar(&userFlag, "user", false, "User operation.")
	flag.StringVar(&addFlag, "add", "", "Add user.")
	flag.StringVar(&removeFlag, "remove", "", "Delete user.")
	flag.BoolVar(&listFlag, "list", false, "List of all users in the system.")
	//port := flag.Int("port", 0, "Change port.")

	flag.Parse()

	if adminFlag != false {

		//Do an action on the administrator
		//Add an administrator
		if addFlag != "" {
			fmt.Println(addFlag)
			err := addUser(1, addFlag)
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Println(fmt.Sprintf("%s has been granted admin privileges.", addFlag))
			return
			//Delete the administrator
		} else if removeFlag != "" {

			err := deleteUser(1, addFlag)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			fmt.Println(fmt.Sprintf("%s removed from the list of administrators.", addFlag))
			return

		} else {

			fmt.Println("Example of use:")
			fmt.Println("--admin --add 5c3fdb68680711a2f5f143d2a0a0f27ccfe51194cc349bdb2f2d5e705c7f2a8c")
			fmt.Println("--admin --remove 5c3fdb68680711a2f5f143d2a0a0f27ccfe51194cc349bdb2f2d5e705c7f2a8c")
		}

	}

	if userFlag != false {
		//Do an action with a user user
		//Add user
		if addFlag != "" {
			err := addUser(2, addFlag)
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Println(fmt.Sprintf("%s has been granted user privileges.", addFlag))
			return
			//Deleting a user
		} else if removeFlag != "" {
			err := deleteUser(2, addFlag)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			fmt.Println(fmt.Sprintf("%s removed from the list of users.", addFlag))
			return

		} else {
			fmt.Println("Example of use:")
			fmt.Println("--user --add 5c3fdb68680711a2f5f143d2a0a0f27ccfe51194cc349bdb2f2d5e705c7f2a8c")
			fmt.Println("--user --remove 5c3fdb68680711a2f5f143d2a0a0f27ccfe51194cc349bdb2f2d5e705c7f2a8c")
		}
	}

	if listFlag != false {
		users, err := listUsers()
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		for _, item := range users {
			fmt.Println(item)
		}
		return

	}

	// if *port != 0 {
	// 	//TODO

	// }

	globalIp = getGlobalIP()

	RBACinit()

	ctx := context.Background()
	//privKey, _ := LoadKeyFromFile()
	portRecord, dhtRecord, privKey, err := SQLgetSettings()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	listen, err := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", portRecord))
	if err != nil {
		fmt.Println(err.Error())
	}

	// Initialize our host
	h, mydht, _, err := SetupLibp2p(
		ctx,
		privKey,
		nil,
		[]multiaddr.Multiaddr{listen},
		nil,
		Libp2pOptionsExtra...,
	)
	if err != nil {
		fmt.Println(err.Error())
		return

	}
	defer h.Close()

	fmt.Println("My id: ", h.ID().String())
	fmt.Println("My address: ", h.Addrs())
	fmt.Println("My CID:", dhtRecord)

	h.Network().Notify(&network.NotifyBundle{
		ConnectedF: handleConnection,
	})

	router := NewRouter()

	// Регистрируем обработчики для маршрутов
	router.HandleFunc("Auth", AuthXML)
	router.HandleFunc("List", ListXML)
	router.HandleFunc("Start", RunXML)
	router.HandleFunc("Stop", StopXML)
	router.HandleFunc("Status", StatusXML)
	router.HandleFunc("Running", RunningXML)
	router.HandleFunc("Add", AddXML)
	h.SetStreamHandler("/conductor/0.0.1", streamHandler(router))

	// Connect to a known host
	bootstrapHost, _ := multiaddr.NewMultiaddr("/ip4/104.131.131.82/tcp/4001/p2p/QmaCpDMGvV2BGHeYERUEnRQAwe3N8SzbUtfsmvsqQLuvuJ")
	peerinfo, _ := peer.AddrInfoFromP2pAddr(bootstrapHost)
	err = h.Connect(ctx, *peerinfo)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	time.Sleep(5 * time.Second)

	provideCid := cid.NewCidV1(cid.Raw, []byte(dhtRecord))

	// Provide a value in the DHT
	if err := mydht.Provide(ctx, provideCid, true); err != nil {
		log.Fatalf("Failed to provide value: %v", err)
		return
	}

	fmt.Println("Ready")
	var wg sync.WaitGroup
	wg.Add(1) // Add 1 goroutine to wait
	wg.Wait() // Block here until the WaitGroup is done

}
