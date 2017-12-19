package proxy

import (
	"encoding/json"
	//"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"pos-proxy/config"
	"pos-proxy/helpers"
	"strings"
	"syscall"
	"time"
)

func CheckForupdates() {
	type NewVersion struct {
		BuildNumber string `json:"build_number"`
	}
	netClient := helpers.NewNetClient()
	for {
		log.Println("Checking for updates...")
		uri := fmt.Sprintf("%s%s", config.Config.BackendURI, "/api/pos/proxy/update/")
		requestBody := fmt.Sprintf("{\"build_number\": \"%s\"}", config.BuildNumber)
		log.Println("Request", uri, requestBody)
		req, err := http.NewRequest("POST", uri, strings.NewReader(requestBody))
		req = helpers.PrepareRequestHeaders(req)
		resp, err := netClient.Do(req)
		if err != nil {
			log.Println("Failed to get update data", err.Error())
		}

		respBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println("Failed to read update data", err.Error())
		}

		data := NewVersion{}
		err = json.Unmarshal(respBody, &data)
		if err != nil {
			log.Println("Failed to parse update data", string(respBody), err.Error())
		}
		log.Println(fmt.Sprintf("New version \"%s\"", data.BuildNumber))
		if data.BuildNumber != "" {
			initiateUpdate(data.BuildNumber)
		}

		resp.Body.Close()
		time.Sleep(5 * time.Minute)
	}
}

func initiateUpdate(buildNumber string) error {
	dir, err := ioutil.TempDir("", "example")
	log.Println("Creating staging area in ", dir)
	gsPath := fmt.Sprintf("gs://pos-proxy/test/%s/update.sh", buildNumber)
	cmd := exec.Command("gsutil", "-m", "cp", gsPath, dir)
	cmd.Env = append(os.Environ())
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	log.Println("Will download")
	err = cmd.Run()
	if err != nil {
		log.Println("Failed to fetch update script")
		os.RemoveAll(dir)
		return err
	}
	updateCommand := fmt.Sprintf("%s/update.sh", dir)
	err = os.Chmod(updateCommand, 0555)
	if err != nil {
		log.Println("Failed to chmod on update command", err.Error())
		os.RemoveAll(dir)
		return err
	}
	log.Println("Done Downloading")
	update(buildNumber, updateCommand, dir)
	return nil
}

func update(buildNumber, updateCommand, updateDir string) {
	log.Println("Starting update process")
	cmd := exec.Command(updateCommand, "test", buildNumber, updateDir)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	err := cmd.Start()
	if err != nil {
		log.Println("Failed to start update process", err.Error())
		os.RemoveAll(updateDir)
	}
}
