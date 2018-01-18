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

type NewVersion struct {
	BuildNumber int64 `json:"build_number"`
}

func CheckForupdates() {
	netClient := helpers.NewNetClient()
	for {
		auth.FetchToken()
		fmt.Println("Checking for updates...")
		uri := fmt.Sprintf("%s%s", config.Config.BackendURI, "/api/proxyversions/getupdate/")

		val, err := strconv.ParseInt(config.BuildNumber, 10, 64)
		if err != nil {
			log.Println("Failed to convert build number", err.Error())
			time.Sleep(5 * time.Minute)
			continue
		}
		config.Config.BuildNumber = &val
		requestBody, err := json.Marshal(config.Config)
		if err != nil {
			log.Println("Failed to marshal config", err.Error())
			time.Sleep(5 * time.Minute)
			continue
		}
		req, err := http.NewRequest("POST", uri, bytes.NewBuffer(requestBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", auth.Token))
		resp, err := netClient.Do(req)
		if err != nil {
			log.Println("Failed to get update data", err.Error())
			time.Sleep(5 * time.Minute)
			continue
		}

		data, err := parseBody(resp)
		if err != nil {
			time.Sleep(5 * time.Minute)
			continue
		}

		fmt.Println(fmt.Sprintf("Configured version \"%d\"", data.BuildNumber))
		if data.BuildNumber != 0 && data.BuildNumber != *config.Config.BuildNumber {
			initiateUpdate(data.BuildNumber)
		}

		time.Sleep(5 * time.Minute)
	}
}

func parseBody(resp *http.Response) (NewVersion, error) {
	respBody, err := ioutil.ReadAll(resp.Body)
	data := NewVersion{}
	if err != nil {
		log.Println("Failed to read update data")
		return data, err
	}
	defer resp.Body.Close()

	err = json.Unmarshal(respBody, &data)
	if err != nil {
		log.Println("Failed to parse update data")
		return data, err
	}
	return data, nil
}

func initiateUpdate(buildNumber int64) error {
	dir, err := ioutil.TempDir("", "proxyupdates")
	fmt.Println("Creating staging area in ", dir)
	gsPath := fmt.Sprintf("gs://pos-proxy/%s/%d/update.sh", *config.Config.VirtualHost, buildNumber)
	cmd := exec.Command("gsutil", "-m", "cp", gsPath, dir)
	cmd.Env = append(os.Environ())
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	fmt.Println("Will download")
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
	fmt.Println("Done Downloading")
	update(buildNumber, updateCommand, dir)
	return nil
}

func update(buildNumber int64, updateCommand, updateDir string) {
	fmt.Println("Starting update process")
	cmd := exec.Command(updateCommand, *config.Config.VirtualHost, fmt.Sprintf("%d", buildNumber), updateDir)
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
