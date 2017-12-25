package proxy

import (
	"encoding/json"
	"strconv"
	//"errors"
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"pos-proxy/auth"
	"pos-proxy/config"
	"pos-proxy/helpers"
	"syscall"
	"time"
)

func CheckForupdates() {
	auth.FetchToken()
	type NewVersion struct {
		BuildNumber int64 `json:"build_number"`
	}
	netClient := helpers.NewNetClient()
	for {
		log.Println("Checking for updates...")
		uri := fmt.Sprintf("%s%s", config.Config.BackendURI, "/api/proxyversions/getupdate/")

		val, err := strconv.ParseInt(config.BuildNumber, 10, 64)
		if err != nil {
			log.Println("Failed to convert build number", err.Error())
		}
		config.Config.BuildNumber = &val
		requestBody, err := json.Marshal(config.Config)
		if err != nil {
			log.Println("Failed to marshal config", err.Error())
		}
		req, err := http.NewRequest("POST", uri, bytes.NewBuffer(requestBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", auth.Token))
		resp, err := netClient.Do(req)
		if err != nil {
			log.Println("Failed to get update data", err.Error())
			return
		}

		respBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println("Failed to read update data", err.Error())
			return
		}

		data := NewVersion{}
		err = json.Unmarshal(respBody, &data)
		if err != nil {
			log.Println("Failed to parse update data", string(respBody), err.Error())
			return
		}
		log.Println(fmt.Sprintf("New version \"%d\"", data.BuildNumber))
		if data.BuildNumber != 0 && data.BuildNumber != *config.Config.BuildNumber {
			initiateUpdate(data.BuildNumber)
		}

		resp.Body.Close()
		time.Sleep(5 * time.Minute)
	}
}

func initiateUpdate(buildNumber int64) error {
	dir, err := ioutil.TempDir("", "proxyupdates")
	log.Println("Creating staging area in ", dir)
	gsPath := fmt.Sprintf("gs://pos-proxy/%s/%d/update.sh", config.VirtualHost, buildNumber)
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
	err = os.Chmod(dir, 0777)
	if err != nil {
		log.Println("Failed to chmod on update folder", err.Error())
		os.RemoveAll(dir)
		return err
	}
	updateCommand := fmt.Sprintf("%s/update.sh", dir)
	err = os.Chmod(updateCommand, 0777)
	if err != nil {
		log.Println("Failed to chmod on update command", err.Error())
		os.RemoveAll(dir)
		return err
	}
	log.Println("Done Downloading")
	update(buildNumber, updateCommand, dir)
	return nil
}

func update(buildNumber int64, updateCommand, updateDir string) {
	log.Println("Starting update process")
	cmd := exec.Command(updateCommand, config.VirtualHost, fmt.Sprintf("%d", buildNumber), updateDir)
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
