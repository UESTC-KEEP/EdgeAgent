/*
Copyright 2016 The Kubernetes Authors.

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

package util

import (
	"fmt"
	"github.com/UESTC-KEEP/keep/constants"
	"github.com/UESTC-KEEP/keep/constants/cloud"
	"github.com/UESTC-KEEP/keep/constants/edge"
	"github.com/UESTC-KEEP/keep/pkg/util/loggerv1.0.1"
	"io"
	"io/ioutil"
	utilnet "k8s.io/apimachinery/pkg/util/net"
	"net"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

func GetLocalIP(hostName string) (string, error) {
	var ipAddr net.IP
	var err error
	addrs, _ := net.LookupIP(hostName)
	for _, addr := range addrs {
		if err := ValidateNodeIP(addr); err == nil {
			if addr.To4() != nil {
				ipAddr = addr
				break
			}
			if addr.To16() != nil && ipAddr == nil {
				ipAddr = addr
			}
		}
	}
	if ipAddr == nil {
		ipAddr, err = utilnet.ChooseHostInterface()
	}

	if err != nil {
		return "", err
	}

	return ipAddr.String(), nil
}

// ValidateNodeIP validates given node IP belongs to the current host
func ValidateNodeIP(nodeIP net.IP) error {
	// Honor IP limitations set in setNodeStatus()
	if nodeIP.To4() == nil && nodeIP.To16() == nil {
		return fmt.Errorf("nodeIP must be a valid IP address")
	}
	if nodeIP.IsLoopback() {
		return fmt.Errorf("nodeIP can't be loopback address")
	}
	if nodeIP.IsMulticast() {
		return fmt.Errorf("nodeIP can't be a multicast address")
	}
	if nodeIP.IsLinkLocalUnicast() {
		return fmt.Errorf("nodeIP can't be a link-local unicast address")
	}
	if nodeIP.IsUnspecified() {
		return fmt.Errorf("nodeIP can't be an all zeros address")
	}

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return err
	}
	for _, addr := range addrs {
		var ip net.IP
		switch v := addr.(type) {
		case *net.IPNet:
			ip = v.IP
		case *net.IPAddr:
			ip = v.IP
		}
		if ip != nil && ip.Equal(nodeIP) {
			return nil
		}
	}
	return fmt.Errorf("Node IP: %q not found in the host's network interfaces", nodeIP.String())
}

//Command executes command and returns output
func Command(name string, arg []string) (string, error) {
	cmd := exec.Command(name, arg...)
	ret, err := cmd.Output()
	if err != nil {
		logger.Error("exec command failed: %v", err)
		return string(ret), err
	}
	return strings.Trim(string(ret), "\n"), nil
}

//GetCurPath returns filepath
func GetCurPath() string {
	file, _ := exec.LookPath(os.Args[0])
	path, _ := filepath.Abs(file)
	rst := filepath.Dir(path)
	return rst
}

func SpliceErrors(errors []error) string {
	if len(errors) == 0 {
		return ""
	}
	var stb strings.Builder
	stb.WriteString("[\n")
	for _, err := range errors {
		stb.WriteString(fmt.Sprintf("  %s\n", err.Error()))
	}
	stb.WriteString("]\n")
	return stb.String()
}

//OrganizeConfigurationFile 组织自动下发配置文件 日志文件
func OrganizeConfigurationFile(agentName string) {
	fmt.Println(os.Getwd())
	var sourceConfDir, destinationConfDir string
	if agentName == edge.EdgeAgentName {
		sourceConfDir = edge.EdgeConfigeFilesSourceDir
	} else if agentName == cloud.CloudAgentName {
		sourceConfDir = cloud.CloudConfigeFilesSourceDir
	} else {
		logger.Error("系统中不存在模块：", agentName)
	}
	//path,_ := os.Getwd()
	//logger.Error(path)
	destinationConfDir = constants.DefaultConfigeFilesDestDir
	//fmt.Println(sourceConfDir,destinationConfDir)
	err := CopyDir(sourceConfDir, destinationConfDir)
	if err != nil {
		logger.Error(err)
	}
}

func CopyDir(src string, dst string) error {
	var err error
	var fds []os.FileInfo
	var srcinfo os.FileInfo

	if srcinfo, err = os.Stat(src); err != nil {
		return err
	}

	if err = os.MkdirAll(dst, srcinfo.Mode()); err != nil {
		return err
	}

	if fds, err = ioutil.ReadDir(src); err != nil {
		return err
	}
	for _, fd := range fds {
		srcfp := path.Join(src, fd.Name())
		dstfp := path.Join(dst, fd.Name())

		if fd.IsDir() {
			if err = CopyDir(srcfp, dstfp); err != nil {
				fmt.Println(err)
			}
		} else {
			if err = CopyFile(srcfp, dstfp); err != nil {
				fmt.Println(err)
			}
		}
	}
	return nil
}

func CopyFile(src, dst string) error {
	var err error
	var srcfd *os.File
	var dstfd *os.File
	var srcinfo os.FileInfo

	if srcfd, err = os.Open(src); err != nil {
		return err
	}
	defer srcfd.Close()

	if dstfd, err = os.Create(dst); err != nil {
		return err
	}
	defer dstfd.Close()

	if _, err = io.Copy(dstfd, srcfd); err != nil {
		return err
	}
	if srcinfo, err = os.Stat(src); err != nil {
		return err
	}
	return os.Chmod(dst, srcinfo.Mode())
}

func AllFilesOfDir(Dir string, files *[]string) error {
	files_, err := ioutil.ReadDir(Dir)
	if err != nil {
		return err
	}
	// 获取文件，并输出它们的名字
	for _, file := range files_ {
		*files = append(*files, file.Name())
	}
	logger.Trace(*files)
	return nil
}

func PrintKEEPLogo() {
	fmt.Printf("%c[1;5;34m%s%c[0m", 0x1B, "  __       ___    __________     ___________   ____________                                  __     __      \n", 0x1B)
	fmt.Printf("%c[1;5;34m%s%c[0m", 0x1B, " |  |     |  |   |  ________|   |  ________|  |   ______   |                                |  |   |  |   \n", 0x1B)
	fmt.Printf("%c[1;5;34m%s%c[0m", 0x1B, " |  |   |  |     |  |           |  |          |  |      |  | 	 _      _    _______       |  |   |  |     \n", 0x1B)
	fmt.Printf("%c[1;5;34m%s%c[0m", 0x1B, " |   | |  |      |  |_______    |  |_______   |  |______|  |    | |    | |  |  ___  |     |  |   |  |   \n", 0x1B)
	fmt.Printf("%c[1;5;34m%s%c[0m", 0x1B, " |   ||  |       |   _______|   |  _______|   |   _________|    | |    | |  | |___| |    |  |   |  |       \n", 0x1B)
	fmt.Printf("%c[1;5;34m%s%c[0m", 0x1B, " |  |   |  |     |  |           |  |          |  |              | |    | |  | ______|   |__|   |__|         \n", 0x1B)
	fmt.Printf("%c[1;5;34m%s%c[0m", 0x1B, " |  |    |  |    |  |_______    |  |_______   |  |              | |____| |  | |         __     __     \n", 0x1B)
	fmt.Printf("%c[1;5;34m%s%c[0m", 0x1B, " |__|     |__|   |__________|   |__________|  |__|    V 1.0.0   |________|  |_|        |__|   |__|          \n", 0x1B)
}

//// GetPodSandboxImage return snadbox image name based on arch, default image is for amd64.
//func GetPodSandboxImage() string {
//	switch runtime.GOARCH {
//	case "arm":
//		return constants.DefaultArmPodSandboxImage
//	case "arm64":
//		return constants.DefaultArm64PodSandboxImage
//	default:
//		return constants.DefaultPodSandboxImage
//	}
//}
