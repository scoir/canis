package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

func main() {
	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	//images, err := cli.ImageList(ctx, types.ImageListOptions{})
	//if err != nil {
	//	panic(err)
	//}
	//
	//for _, image := range images {
	//	fmt.Printf("%s\t%s\n", image.ID, image.RepoTags)
	//}

	args := filters.NewArgs()
	args.Add("name", "canis_steward")
	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{
		Filters: args,
	})

	for _, c := range containers {
		if c.State != "running" {
			err := cli.ContainerRemove(ctx, c.ID, types.ContainerRemoveOptions{})
			if err != nil {
				log.Printf("error removing container %s\n", c.Names[0])
			} else {
				log.Printf("successfully removed container %s\n", c.Names[0])
			}
		} else {
			fmt.Printf("%s %s %s\n", c.ID[:10], c.Image, c.State)
		}
	}

	out, err := cli.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{ShowStdout: true})
	if err != nil {
		panic(err)
	}

	_, err = stdcopy.StdCopy(os.Stdout, os.Stderr, out)
	if err != nil {
		panic(err)
	}
}
