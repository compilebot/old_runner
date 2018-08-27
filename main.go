package main

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"

	"github.com/docker/docker/client"
)

func main() {
	New()
}

func New() {
	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}

	//x := getLogs(containers[1].ID, cli)
	x := buildContainer(cli)

	y := getLogs(x, cli)
	fmt.Println(y)

}

func buildContainer(cli *client.Client) string {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	config := container.Config{Cmd: []string{"echo", "hello", "world"}, Image: "alpine"}
	hostconfig := container.HostConfig{}
	netconfig := network.NetworkingConfig{}

	c, err := cli.ContainerCreate(ctx, &config, &hostconfig, &netconfig, "test")
	if err != nil {
		panic(err)
	}
	if err := cli.ContainerStart(ctx, c.ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}

	return c.ID
}

func getLogs(id string, cli *client.Client) string {
	var b strings.Builder

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	reader, err := cli.ContainerLogs(ctx, id, types.ContainerLogsOptions{ShowStdout: true, ShowStderr: true})
	if err != nil {
		fmt.Println(err)
	}

	_, err = io.Copy(&b, reader)
	if err != nil && err != io.EOF {
		fmt.Println(err)
	}

	return b.String()
}
