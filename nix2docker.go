package main

import (
	"archive/tar"
	"compress/gzip"
	"encoding/hex"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
)

func init() {
	log.SetFlags(0)
	log.SetOutput(os.Stderr)
}

func main() {
	var config struct {
		DockerConfig map[string]interface{}
		Paths        []string
		Graphs       []string
		Repository   string
	}

	if err := readJSONFile(os.Getenv("configPath"), &config); err != nil {
		log.Fatal(err)
	}

	outHash, _, ok := parseNixPath(os.Getenv("out"))
	if !ok {
		log.Fatalf("couldn't parse $out")
	}

	graph := make(Graph)
	for _, file := range config.Graphs {
		_, err := ParseGraphFile(graph, file)
		if err != nil {
			log.Fatalf("ParseGraphFile: %s: %s", file, err)
		}
	}

	outDir := "out"
	imageID := hex.EncodeToString(outHash[:])
	imageDir := filepath.Join(outDir, imageID)
	if err := os.MkdirAll(imageDir, 0755); err != nil {
		log.Fatal(err)
	}

	layer, err := os.OpenFile(filepath.Join(imageDir, "layer.tar"), os.O_RDWR|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer layer.Close()

	tarSize, tarSum, err := PackLayerSum(layer, graph, config.Paths)
	if err != nil {
		log.Fatalf("PackImageSum: %s", err)
	}

	manifest := map[string]interface{}{
		"config":       config.DockerConfig,
		"architecture": runtime.GOARCH,
		"os":           runtime.GOOS,
		"created":      "1970-01-01T00:00:01Z",
		"id":           imageID,
		"checksum":     tarSum,
		"Size":         tarSize,
	}
	if err := writeJSONFile(filepath.Join(imageDir, "json"), manifest, 0644); err != nil {
		log.Fatalf("couldn't write manifest: %s", err)
	}
	if err := ioutil.WriteFile(filepath.Join(imageDir, "VERSION"), []byte("1.0"), 0644); err != nil {
		log.Fatalf("couldn't write VERSION: %s", err)
	}

	repos := map[string]map[string]string{
		config.Repository: {imageID: imageID},
	}
	if err := writeJSONFile(filepath.Join(outDir, "repositories"), repos, 0644); err != nil {
		log.Fatalf("couldn't write repositories: %s", err)
	}

	f, err := os.OpenFile(os.Getenv("out"), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	defer f.Close()
	w := gzip.NewWriter(f)
	t := tar.NewWriter(w)

	err = filepath.Walk(outDir, PackTarWalkFunc(t, outDir))
	if err == nil {
		err = t.Close()
	}
	if err == nil {
		err = w.Close()
	}
	if err != nil {
		log.Fatalf("couldn't write output: %s", err)
	}

	if err := ioutil.WriteFile(os.Getenv("tag"), []byte(config.Repository+":"+imageID), 0644); err != nil {
		log.Fatalf("couldn't write tag: %s", err)
	}
}
