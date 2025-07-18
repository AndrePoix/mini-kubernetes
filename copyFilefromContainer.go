package main

import (
    "archive/tar"
    "bytes"
    "context"
    "io"
    "os"
    "time"

    "github.com/docker/docker/client"
)

func copyFileFromContainer(containerID, containerPath, localPath string, cli *client.Client, parentContext context.Context) error {

    ctx, cancel := context.WithTimeout(parentContext, 10*time.Second)
    defer cancel()
    reader, _, err := cli.CopyFromContainer(ctx, containerID, containerPath)
    if err != nil {
        return err
    }
    defer reader.Close()

    tr := tar.NewReader(reader)

    for {
        hdr, err := tr.Next()
        if err == io.EOF {
            break
        }
        if err != nil {
            return err
        }

        if hdr.Typeflag == tar.TypeReg {
            buf := new(bytes.Buffer)
            if _, err := io.Copy(buf, tr); err != nil {
                return err
            }

            return os.WriteFile(localPath, buf.Bytes(), 0644)
        }
    }

    return nil
}
