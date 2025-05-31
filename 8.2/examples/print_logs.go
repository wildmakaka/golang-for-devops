package examples

import (
	"context"
	"io"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

//печатаем логи из контейнера
func PrintLogs(container string) {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation()) //собираем клиент с опциями из env переменных
	if err != nil {
		panic(err)
	}

	options := types.ContainerLogsOptions{ShowStdout: true} //устанавливаем настройки для логирования контейнера
	out, err := cli.ContainerLogs(ctx, container, options)  //здесь печатаем логи
	if err != nil {
		panic(err)
	}

	io.Copy(os.Stdout, out) //просто копируем все вместе в stdout, в примерах с run_container будет показано как разделить stdout и stderr
}
