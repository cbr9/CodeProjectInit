package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/rjeczalik/notify"
)

const configFile = "./config.json"

func main() {
	watcher := newWatcher()
	watcher.watch()
}

type watcher struct {
	config config
}

type config struct {
	ProjectsDir string              `json:"projects_dir"`
	Languages   map[string]language `json:"languages"`
}

type language struct {
	Depth        int      `json:"depth"`
	ExcludedDirs []string `json:"excluded_dirs"`
	ExtraCmd     string   `json:"extra_cmd"`
}

func getConfig() *config {
	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Fatal(err)
		return nil
	}
	var config config
	err = json.Unmarshal(data, &config)
	if err != nil {
		fmt.Println(err)
	}
	return &config
}

func newWatcher() *watcher {
	config := getConfig()
	return &watcher{*config}
}

func (w *watcher) runCmd(newFolder string, command string) {
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
func (w *watcher) run(fi os.FileInfo, path string) {
	if fi.IsDir() {
		pathChunks := strings.Split(strings.TrimPrefix(path, w.config.ProjectsDir), string(os.PathSeparator))
		language := w.config.Languages[pathChunks[0]]
		if len(pathChunks) == language.Depth && !isUnixHiddenDir(path) { // avoid hidden folders in Unix systems
			if (len(language.ExcludedDirs) > 0 && !contains(pathChunks, language.ExcludedDirs...)) || len(language.ExcludedDirs) == 0 {
				if language.ExtraCmd != "" {
					w.runCmd(path, language.ExtraCmd)
				}
				w.runCmd(path, "git init")
			}
		}
	}
}

func isUnixHiddenDir(name string) bool {
	return strings.HasPrefix(name, ".")
}

func contains(slice []string, elements ...string) bool {
	for i := range slice {
		for _, elem := range elements {
			if slice[i] == elem {
				return true
			}
		}
	}
	return false
}

func (w *watcher) watch() {
	_ = os.Chdir(w.config.ProjectsDir)
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
			w.run(fi, change.Path())
		}
	}
}
