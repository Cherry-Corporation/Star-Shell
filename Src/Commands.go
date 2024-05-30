package main

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/google/go-github/github"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
)

type PackageManager struct {
	client *github.Client
}

func NewPackageManager() *PackageManager {
	return &PackageManager{
		client: github.NewClient(nil),
	}
}

func DownloadFile(url string, filepath string) error {
	dir := path.Dir(filepath)

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.MkdirAll(dir, 0755)
	}

	if _, err := os.Stat(filepath); err == nil {
		log.Printf("File %s already exists. Skipping download.\n", filepath)
		return nil
	} else if !os.IsNotExist(err) {
		// Some other error occurred
		return fmt.Errorf("Failed to check if file exists: %w", err)
	}

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func (pm *PackageManager) Install(user, repo string) error {
	release, _, err := pm.client.Repositories.GetLatestRelease(context.Background(), user, repo)
	if err != nil {
		return fmt.Errorf("Failed to get latest release: %w", err)
	}

	for _, asset := range release.Assets {
		if strings.HasSuffix(asset.GetName(), ".tar.gz") || strings.HasSuffix(asset.GetName(), ".zip") {
			url := asset.GetBrowserDownloadURL()
			filepath := path.Join("downloads", asset.GetName())

			err := DownloadFile(url, filepath)
			if err != nil {
				return fmt.Errorf("Failed to download asset: %w", err)
			}
			log.Printf("Downloaded asset to %s\n", filepath)
			break
		}
	}

	return nil
}

func wget(url string) {
	output := path.Base(url)
	fmt.Printf("Downloading %s to %s...\n", url, output)

	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer resp.Body.Close()

	out, err := os.Create(output)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("Download completed!")
}

func ls() {
	files, err := ioutil.ReadDir(currentDir)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	for _, file := range files {
		fmt.Println(file.Name())
	}
}

func cd(dir string, theme Theme) {
	if err := os.Chdir(dir); err != nil {
		getColor(theme.ErrorColor).Printf("Failed to change directory: %v\n", err)
		return
	}
	// Update the current directory
	currentDir, _ = os.Getwd()
	getColor(theme.OutputColor).Printf("Changed directory to %s\n", currentDir)
}

func help() {
	fmt.Println("Available commands:")
	fmt.Println("  help          - Show this help message")
	fmt.Println("  exit          - Exit the shell")
	fmt.Println("  ls            - List files in the current directory")
	fmt.Println("  cd <dir>      - Change directory")
	fmt.Println("  wget <url>    - Download a file from the web")
	fmt.Println("  verfetch      - Fetch system version information")
	fmt.Println("  ip            - Print the main IP address")
	fmt.Println("  pkg install <user/repo> - Install a package from GitHub")
}

func verfetch() {
	v, _ := host.Info()
	cpuStat, _ := cpu.Info()
	vMem, _ := mem.VirtualMemory()
	diskStat, _ := disk.Usage("/")

	color.Magenta("OS: %s", v.Platform)
	color.Magenta("OS Version: %s", v.PlatformVersion)
	color.Magenta("Kernel: %s", v.KernelVersion)
	color.Magenta("Architecture: %s", v.KernelArch)
	color.Magenta("CPU: %s", cpuStat[0].ModelName)
	color.Magenta("Cores: %d", cpuStat[0].Cores)
	color.Magenta("Total Memory: %v GB", bToGb(vMem.Total))
	color.Magenta("Available Memory: %v GB", bToGb(vMem.Available))
	color.Magenta("Used Memory: %v GB", bToGb(vMem.Used))
	color.Magenta("Disk Total: %v GB", bToGb(diskStat.Total))
	color.Magenta("Disk Used: %v GB", bToGb(diskStat.Used))
	color.Magenta("Disk Free: %v GB", bToGb(diskStat.Free))
}

// Converts bytes to gigabytes
func bToGb(b uint64) uint64 {
	return b / (1024 * 1024 * 1024)
}

func hello() {
	fmt.Println("Hello, welcome to Cherry Terminal!")
}

func now() {
	currentTime := time.Now()
	fmt.Println("Current time: ", currentTime.Format("15:04:05"))
}

func printMainIP() {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		color.Red("Oops: %v\n", err.Error())
		return
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	color.Green("IP address: %s", localAddr.IP.String())
}
