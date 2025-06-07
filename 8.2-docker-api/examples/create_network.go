package examples

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"io"
	"os"
)

// Пример создания сети
func CreateNetwork() {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation()) //собираем клиент с опциями из env переменных
	if err != nil {
		panic(err)
	}

	imageName := "postgres"

	out, err := cli.ImagePull(ctx, imageName, types.ImagePullOptions{}) //качаем image с именем postgres
	if err != nil {
		panic(err)
	}
	defer out.Close()       //закрываем ответ от sdk в конце функции
	io.Copy(os.Stdout, out) //копируем полный вывод в stdout для пользователя

	resp, err := cli.ContainerCreate(ctx, &container.Config{ //создаем контейнер
		Hostname:   "postgresC",
		Domainname: "postgresC",
		Image:      imageName,
		Env: []string{
			"POSTGRES_PASSWORD=example",
		},
	}, nil, nil, nil, "postgresC")
	if err != nil {
		panic(err)
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	} //запускаем контейнер
	fmt.Println(resp.ID)
	networkCreate, err := cli.NetworkCreate(ctx, "mynet", types.NetworkCreate{Attachable: true}) //создаем сеть
	if err != nil {
		panic(err)
	}

	fmt.Println("Network created with ID:" + networkCreate.ID)

	err = cli.NetworkConnect(ctx, networkCreate.ID, resp.ID, &network.EndpointSettings{}) // и добавляем наш контейнер к ней
	if err != nil {
		panic(err)
	}
}
