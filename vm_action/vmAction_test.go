package vm_action

import (
	"context"
	"encoding/xml"
	"fmt"
	vmSQL "main/sql"
	"os"
	"testing"

	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
)

func importTar(path string) {

	tarFilePath := path + "/hello_world/hello.tar"

	// Открываем tar файл
	tarFile, err := os.Open(tarFilePath)
	if err != nil {
		fmt.Printf("Ошибка при открытии tar файла: %v\n", err)
		return
	}
	defer tarFile.Close()

	// Создаем Docker клиент
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithVersion("1.41"))
	if err != nil {
		fmt.Printf("Ошибка при создании Docker клиента: %v\n", err)
		return
	}

	// Загружаем образ из tar файла
	// Используем archive.Overwrite для перезаписи существующих образов
	_, err = cli.ImageLoad(context.Background(), tarFile, true)
	if err != nil {
		fmt.Printf("Ошибка при загрузке образа: %v\n", err)
		return
	}

	fmt.Println("Образ успешно импортирован.")
}

func removeImage() {

	// Specify the name or ID of the image you want to delete
	imageName := "hello:latest" // Например, "my_image:latest"

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithVersion("1.41"))
	if err != nil {
		fmt.Printf("Ошибка при создании Docker клиента: %v\n", err)
		return
	}

	// Deleting the image
	removeOptions := image.RemoveOptions{
		Force: true,
	}

	_, err = cli.ImageRemove(context.Background(), imageName, removeOptions)
	if err != nil {
		fmt.Printf("Ошибка при удалении образа: %v\n", err)
		return
	}

	fmt.Println("Образ успешно удален.")

}

type Pod_t struct {
	PodName string `xml:"PodName"`
	Hash    string `xml:"Hash"`
}

type Response_t struct {
	XMLName xml.Name `xml:"Response"`
	Pods    []Pod_t  `xml:"Pod"`
}

func TestVMStart_Normal(t *testing.T) {
	dir, _ := os.Getwd()

	db, err := vmSQL.SQLgetDB()
	if err != nil {
		t.Fatal("[FAIL] vmSQL.SQLgetDB got:", err)
	}

	defer db.Close()
	// Importing the test image
	importTar(dir)

	//Adding to the database
	err = VMCreate(db, Pod{PodName: "Tests", Images: []string{"hello"}, ExternalImage: "hello", Metadata: []string{"lol"}, InternalPort: 80})
	if err != nil {
		t.Errorf("[ERROR] VMCreate got: %s", err.Error())
	} else {
		t.Logf("[OK] VMCreate")
	}

	err = VMCreate(db, Pod{PodName: "Tests", Images: []string{"hello"}, ExternalImage: "hello", Metadata: []string{"lol"}, InternalPort: 80})
	if err == nil {
		t.Errorf("[FAIL] VMCreate got: %s", "There must be an error because two submissions with the same values cannot exist within the same host")
	} else {
		t.Logf("[OK] VMCreate got: %s", err.Error())
	}

	var resp Response_t

	data, err := vmSQL.SQLGetAllPods(db)
	if err != nil {
		t.Errorf("[FAIL] vmSQL.SQLGetAllPods got: %s", err.Error())
	} else {
		t.Logf("[OK] vmSQL.SQLGetAllPods")
	}

	err = xml.Unmarshal(data, &resp)
	if err != nil {
		t.Errorf("[FAIL] xml.Unmarshal got: %s", err.Error())
	} else {
		t.Logf("[OK] %s", resp.Pods[0].Hash)
	}

	_, err = VMStart(db, "badHash", "user123", "1")
	if err == nil {
		t.Errorf("expected error due to bad hash (getPods fail)")
	} else {
		t.Logf("[OK] %s", err.Error())
	}

	port, err := VMStart(db, resp.Pods[0].Hash, "user123", "1")
	if err != nil {
		t.Errorf("[FAIL] VMStart got: %s", err.Error())
	} else {
		t.Logf("[OK] port %d", port)
	}

	port2, err := VMStart(db, resp.Pods[0].Hash, "user123", "1")
	if err != nil {
		t.Errorf("[ERROR] VMStart got: %s", err.Error())
	} else {
		if port != port2 {
			t.Logf("[OK] port1 %d, port2 %d", port, port2)
		} else {
			t.Errorf("[ERROR] port1 %d, port2 %d", port, port2)
		}
	}

	running, err := VMgetRunningPods()
	if err != nil {
		t.Errorf("[FAIL] VMgetRunningPods: %s", err.Error())
	} else {
		for _, container := range running {
			t.Logf("ID1 : %s", container.ID)
		}
	}

	// err = VMstopByNetworkName("user123")
	// if err != nil {
	// 	t.Errorf("[ERROR] VMstopByNetworkName got: %s", err.Error())
	// }

	// running, err = VMgetRunningPods()
	// if err != nil {
	// 	t.Errorf("[FAIL] VMgetRunningPods: %s", err.Error())
	// } else {

	// 	if len(running) == 0 {
	// 		t.Logf("No container")
	// 	} else {
	// 		t.Errorf("[ERROR] Containers not deleted")
	// 	}
	// 	// for _, container := range running {
	// 	// 	t.Logf("ID2 : %s", container.ID)
	// 	// }
	// }

	// err = vmSQL.SQLdeletePod(db, resp.Pods[0].Hash)
	// if err != nil {
	// 	t.Errorf("[FAIL] vmSQL.SQLdeletePod( got: %s", err.Error())
	// } else {
	// 	t.Logf("[OK] vmSQL.SQLdeletePod")
	// }
	// removeImage()
	t.Errorf("[OK]")
}
