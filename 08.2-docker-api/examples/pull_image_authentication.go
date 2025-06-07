package examples

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

func PullImageAuthentication() {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation()) //собираем клиент с опциями из env переменных
	if err != nil {
		panic(err)
	}

	authConfig := types.AuthConfig{
		Username: "nortsx",
		Password: "slurmpass",
	} //авторизуемся в docker с помощью структуры AuthConfig
	encodedJSON, err := json.Marshal(authConfig) //нам надо получить JSON
	if err != nil {
		panic(err)
	}
	authStr := base64.URLEncoding.EncodeToString(encodedJSON) //и закодировать его в base64

	out, err := cli.ImagePull(ctx, "alpine", types.ImagePullOptions{RegistryAuth: authStr}) //который мы передаем в RegistryAuth
	if err != nil {
		panic(err)
	}

	defer out.Close()
	io.Copy(os.Stdout, out)
}
