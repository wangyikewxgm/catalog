/*
Copyright 2021 The KubeVela Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"sigs.k8s.io/yaml"
)

type AddonMeta struct {
	Dependencies []Dependency `json:"dependencies"`
}

type Dependency struct {
	Name string `json:"name"`
}

var file = "addons/velaux/template.yaml"
var regexPattern = "^addons.*"
var globalRexPattern = "^.github.*|makefile|test/e2e-test/addon-test/main.go"

// This can be used for pending some error addon temporally, Please fix it as soon as posible.
var pendingAddon = map[string]bool{"ocm-gateway-manager-addon": true, "model-serving": true}

func main() {
	changedFile := os.Args[1:]
	changedAddon := determineNeedEnableAddon(changedFile)
	if len(changedAddon) == 0 {
		return
	}
	if err := enableAddonsByOrder(changedAddon); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}


func determineNeedEnableAddon (changedFile []string) map[string]bool {
	changedAddon := map[string]bool{}
	globalRex := regexp.MustCompile(globalRexPattern)
	regx := regexp.MustCompile(regexPattern)
	for _, s := range changedFile {
		regRes := regx.Find([]byte(s))
		if len(regRes) != 0 {
			fmt.Println(string(regRes))
			list := strings.Split(string(regRes), "/")
			if len(list) > 1 {
				addon := list[1]
				changedAddon[addon] = true
			}
		}

		if regRes := globalRex.Find([]byte(s)); len(regRes) != 0 {
			// change CI related file, must test all addons
			err := putInAllAddons(changedAddon)
			if err != nil {
				return nil
			} else {
				fmt.Println("This pr need checkAll addons")
				return changedAddon
			}
		}
	}

	for addon := range changedAddon {
		checkAddonDependency(addon, changedAddon)
	}

	fmt.Printf("This pr need test addons: ")
	for ca := range changedAddon {
		fmt.Printf("%s,", ca)
	}
	fmt.Printf("\n")
	return changedAddon
}

func putInAllAddons (addons map[string]bool) error {
	dir, err := ioutil.ReadDir("./addons")
	if err != nil {
		return err
	}
	for _, subDir := range  dir {
		if subDir.IsDir() {
			fmt.Println(subDir.Name())
			addons[subDir.Name()] = true
		}
	}
	return nil
}


func checkAddonDependency(addon string, changedAddon map[string]bool ) {
	metaFile, err := os.ReadFile(filepath.Join([]string{"addons", addon, "metadata.yaml"}...))
	if err != nil {
		panic(err)
	}
	meta := AddonMeta{}
	err = yaml.Unmarshal(metaFile, &meta)
	if err != nil {
		panic(err)
	}

	for _, dep := range meta.Dependencies {
		changedAddon[dep.Name] = true
		checkAddonDependency(dep.Name, changedAddon)
	}
}

// This func will enable addon by order rely-on addon's relationShip dependency,
// this func is so dummy now that the order is written manually, we can generated a dependency DAG workflow in the furture.
func enableAddonsByOrder (changedAddon map[string]bool)  error {
	dirPattern := "addons/%s"
	if changedAddon["fluxcd"] {
		if err := enableOneAddon(fmt.Sprintf(dirPattern, "fluxcd")); err != nil {
			return err
		}
		changedAddon["fluxcd"] = false
	}
	if changedAddon["terraform"] {
		if err := enableOneAddon(fmt.Sprintf(dirPattern, "terraform")); err != nil {
			return err
		}
		changedAddon["terraform"] = false
	}
	if changedAddon["velaux"] {
		if err := enableOneAddon(fmt.Sprintf(dirPattern, "velaux")); err != nil {
			return err
		}
		changedAddon["velaux"] = false
	}
	if changedAddon["cert-manager"] {
		if err := enableOneAddon(fmt.Sprintf(dirPattern, "cert-manager")); err != nil {
			return err
		}
		changedAddon["cert-manager"] = false
	}
	for s, b := range changedAddon {
		if b && !pendingAddon[s] {
			if err := enableOneAddon(fmt.Sprintf(dirPattern, s)); err != nil {
				return err
			}
			if err := disableOneAddon(s); err != nil {
				return err
			}
			switch s {
			case "dex":
				if err := disableOneAddon("velaux"); err != nil {
					return err
				}
			case "flink-kubernetes-operator":
				if err := disableOneAddon("cert-manager"); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func enableOneAddon(dir string) error {
	cmd := exec.Command("vela","addon", "enable", dir)
	fmt.Println(cmd.String())
	stdout, err := cmd.StdoutPipe()
	cmd.Stderr = cmd.Stdout
	if err != nil {
		panic(err)
	}
	if err = cmd.Start(); err != nil {
		return err
	}
	for {
		tmp := make([]byte, 1024)
		_, err := stdout.Read(tmp)
		fmt.Print(string(tmp))
		if err != nil {
			break
		}
	}
	if err = cmd.Wait(); err != nil {
		return err
	}
	return nil
}

func disableOneAddon (addonName string) error {
	cmd := exec.Command("vela","addon", "disable", addonName)
	fmt.Println(cmd.String())
	stdout, err := cmd.StdoutPipe()
	cmd.Stderr = cmd.Stdout
	if err != nil {
		panic(err)
	}
	if err = cmd.Start(); err != nil {
		return err
	}
	for {
		tmp := make([]byte, 1024)
		_, err := stdout.Read(tmp)
		fmt.Print(string(tmp))
		if err != nil {
			break
		}
	}
	if err = cmd.Wait(); err != nil {
		return err
	}
	return nil
}


// this func can be used for debug when addon enable failed.
func checkAppStatus(addonName string)  {
	cmd := exec.Command("vela","status", "-n", "vela-system", "addon-" + addonName)
	fmt.Println(cmd.String())
	stdout, err := cmd.StdoutPipe()
	cmd.Stderr = cmd.Stdout
	if err != nil {
		panic(err)
	}
	if err = cmd.Start(); err != nil {
		fmt.Println(err)
	}
	for {
		tmp := make([]byte, 1024)
		_, err := stdout.Read(tmp)
		fmt.Print(string(tmp))
		if err != nil {
			break
		}
	}
	if err = cmd.Wait(); err != nil {
		fmt.Println(err)
	}
}

func checkPodStatus(namespace string) {
	cmd := exec.Command("kubectl","get pods", "-n", namespace)
	fmt.Println(cmd.String())
	stdout, err := cmd.StdoutPipe()
	cmd.Stderr = cmd.Stdout
	if err != nil {
		panic(err)
	}
	if err = cmd.Start(); err != nil {
		fmt.Println(err)
	}
	for {
		tmp := make([]byte, 1024)
		_, err := stdout.Read(tmp)
		fmt.Print(string(tmp))
		if err != nil {
			break
		}
	}
	if err = cmd.Wait(); err != nil {
		fmt.Println(err)
	}
}
