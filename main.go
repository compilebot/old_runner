package main

import (
	"archive/tar"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"

	"github.com/docker/docker/client"
)

func main() {
	code := `package main

import "fmt"

func main() {
	for i := 0; i < 5000; i++ {
		fmt.Println(i)
	}
}`
	_ = code
	New(code)
}

func New(code string) {
	err := writeCodeToFile(code)

	if err != nil {
		panic(err)
	}

	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}

	buildImage(cli)
	x := buildContainer(cli)

	y := getLogs(x, cli)
	fmt.Println(y)

	cleanup(cli, "test")

}

func buildImage(cli *client.Client) {

	config := types.ImageBuildOptions{Tags: []string{"gotest"}}
	createBuildContext()
	buildContext, err := os.Open("./tmp/tar/go.tar")
	if err != nil {
		panic(err)
	}
	defer buildContext.Close()

	br, err := cli.ImageBuild(context.Background(), buildContext, config)
	if err != nil {
		panic(err)
	}
	defer br.Body.Close()
	fmt.Println(br, err)

}

func createBuildContext() {
	file, err := os.Create("/home/jake/apps/gopherpun/runner/tmp/tar/go.tar")
	if err != nil {
		panic(err)
	}
	tw := tar.NewWriter(file)
	defer tw.Close()

	files := map[string][]byte{"Dockerfile": nil, "main.go": nil}
	for k, v := range files {
		a, err := ioutil.ReadFile("/home/jake/apps/gopherpun/runner/tmp/go/" + k)
		v = a
		if err != nil {
			panic(err)
		}
		hdr := &tar.Header{Name: k, Mode: 0600, Size: int64(len(v))}

		if err := tw.WriteHeader(hdr); err != nil {
			panic(err)
		}

		if _, err := tw.Write([]byte(v)); err != nil {
			panic(err)
		}

	}

}

func cleanup(cli *client.Client, id string) error {
	config := types.ContainerRemoveOptions{Force: true}
	ctx := context.Background()

	err := cli.ContainerRemove(ctx, id, config)

	if err != nil {
		return err
	}

	return nil
}

func writeCodeToFile(code string) error {
	f, err := os.OpenFile("/home/jake/apps/gopherpun/runner/tmp/go/main.go", os.O_RDWR, 0777)

	if err != nil {
		return err
	}

	if _, err := f.WriteString(code); err != nil {
		return err
	}

	return nil
}

func buildContainer(cli *client.Client) string {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	config := container.Config{Image: "gotest"}
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
