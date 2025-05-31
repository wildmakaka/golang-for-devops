package examples

import (
	"github.com/docker/docker/client"
	"golang.org/x/net/context"
	"io"
	"os"
)

func GetStats() {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation()) //собираем клиент с опциями из env переменных
	if err != nil {
		panic(err)
	}

	stats, err := cli.ContainerStats(ctx, "304a3af5000d", false) // получаем stream из докера, последний параметр определит будет ли у нас бесконечный stream или одноразовыый запрос
	if err != nil {
		panic(err)
	}
	defer stats.Body.Close() // закрываем Body

	io.Copy(os.Stdout, stats.Body) // копируем все в stdout
}
