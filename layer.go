package main

import (
	"archive/tar"
	"io"
	"path/filepath"
	"sort"

	"github.com/docker/docker/pkg/tarsum"
)

func PackLayerSum(w io.Writer, g Graph, paths []string) (n int64, sum string, err error) {
	r := toReader(func(w io.Writer) error {
		return PackLayer(w, g, paths)
	})
	s, err := tarsum.NewTarSum(r, true, tarsum.Version1)
	if err != nil {
		return -1, "", err
	}
	n, err = io.Copy(w, s)
	if err == nil {
		sum = s.Sum(nil)
	}
	return
}

func PackLayer(w io.Writer, g Graph, paths []string) error {
	sort.Strings(paths)

	tarWriter := tar.NewWriter(w)
	for _, path := range paths {
		if err := filepath.Walk(path, PackTarForestWalkFunc(tarWriter, path)); err != nil {
			return err
		}
	}

	var closure Graph
	for _, path := range paths {
		closure = g.Closure(closure, path)
	}

	for _, path := range closure.Paths() {
		if err := filepath.Walk(path, PackTarWalkFunc(tarWriter, "/")); err != nil {
			return err
		}
	}

	return tarWriter.Close()
}
