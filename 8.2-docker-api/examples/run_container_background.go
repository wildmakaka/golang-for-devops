package examples

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

//запускаем контейнер и выходим
func RunContainerBackground() {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation()) //собираем клиент с опциями из env переменных
	if err != nil {
		panic(err)
	}

	imageName := "bfirsh/reticulate-splines"

	out, err := cli.ImagePull(ctx, imageName, types.ImagePullOptions{}) //тянем изображение которое запускает бесконечный процесс
	if err != nil {
		panic(err)
	}
	defer out.Close()
	io.Copy(os.Stdout, out)

	resp, err := cli.ContainerCreate(ctx, &container.Config{ //создаем контейнер с минимумом настроек, команду НЕ переопределяем
		Image: imageName,
	}, nil, nil, nil, "")
	if err != nil {
		panic(err)
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	} //запускаем и можно выходить, контейнер будет спокойно крутиться на машине

	fmt.Println(resp.ID)
}
