package main

import (
	"encoding/json"
	"fmt"
	"github.com/rjeczalik/notify"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
)

const configFile = "./config.json"

func main() {
	watcher := NewCodeWatcher()
	watcher.Watch()
}

type watcher struct {
	home   string
	code   string
	config map[string]folderConfig
}

type folderConfig struct {
	Depth        int      `json:"depth"`
	ExcludedDirs []string `json:"excluded_dirs"`
}

func NewCodeWatcher() *watcher {
	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Fatal(err)
		return nil
	}
	var config map[string]folderConfig
	err = json.Unmarshal(data, &config)
	if err != nil {
		fmt.Println(err)
	}

	home := os.Getenv("HOME")
	return &watcher{
		home:   home,
		code:   path.Join(home, "Code"),
		config: config,
	}
}

// From the official documentation:
// "Running git init in an existing repository is safe. It will not overwrite things that are already there".
// This means we don't need to check if it already exists
func (w *watcher) Init(newFolder string, command string) {
	err := os.Chdir(newFolder)
	if err != nil {
		log.Println(err)
		return
	}
	cmdPieces := strings.Split(command, " ")
	name := cmdPieces[0]
	args := cmdPieces[1:]
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		log.Println(err)
	}
}

func IsUnixHiddenDir(name string) bool {
	return strings.HasPrefix(name, ".")
}

func Contains(slice []string, elements ...string) bool {
	for i := range slice {
		for _, elem := range elements {
			if slice[i] == elem {
				return true
			}
		}
	}
	return false
}

func (w *watcher) Watch() {
	_ = os.Chdir(w.code)
	c := make(chan notify.EventInfo, 1)
	err := notify.Watch("./...", c, notify.Create)
	if err != nil {
		log.Fatal(err)
	}
	defer notify.Stop(c)

	for {
		change := <-c
		switch change.Event() {
		case notify.Create:
			fi, err := os.Stat(change.Path())
			if err != nil {
				log.Println(err)
				continue
			}
			if fi.IsDir() {
				pathChunks := strings.Split(strings.TrimPrefix(change.Path(), w.code+"/"), "/")
				topParent := pathChunks[0]
				if len(pathChunks) == w.config[topParent].Depth && !Contains(pathChunks, w.config[topParent].ExcludedDirs...) && !IsUnixHiddenDir(path.Base(change.Path())) {
					// avoid hidden folders
					switch topParent {
					case "Go":
						w.Init(change.Path(), "go mod init")
						w.Init(change.Path(), "git init")
						break
					case "Rust":
						w.Init(change.Path(), "cargo init")
						w.Init(change.Path(), "git init")
						break
					default:
						w.Init(change.Path(), "git init")
					}
				}
			}
		}
	}
}
