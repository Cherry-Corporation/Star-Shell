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
		log.Printf("File  already exists. Skipping download.\n", filepath)
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
			log.Printf("Downloaded asset to \n", filepath)
			break
		}
	}

	return nil
}

func wget(url string, theme Theme) {
	output := path.Base(url)
	fmt.Printf("Downloading  to ...\n", url, output)

	resp, err := http.Get(url)
	if err != nil {
		getColor(theme.ErrorColor).Println("Error:", err)
		return
	}
	defer resp.Body.Close()

	out, err := os.Create(output)
	if err != nil {
		getColor(theme.ErrorColor).Println("Error:", err)
		return
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		getColor(theme.ErrorColor).Println("Error:", err)
		return
	}

	getColor(theme.TextColor).Println("Download completed!")
}

func ls(theme Theme) {
	files, err := ioutil.ReadDir(currentDir)
	if err != nil {
		getColor(theme.ErrorColor).Println("Error:", err)
		return
	}

	for _, file := range files {
		getColor(theme.TextColor).Println(file.Name())
	}
}

func cd(dir string, theme Theme) {
	if err := os.Chdir(dir); err != nil {
		getColor(theme.ErrorColor).Printf("Failed to change directory: \n", err)
		return
	}
	// Update the current directory
	currentDir, _ = os.Getwd()
	getColor(theme.OutputColor).Printf("Changed directory to \n", currentDir)
}

func help(theme Theme) {
	getColor(theme.TextColor).Println("Available commands:")
	fmt.Println("  help          - Show this help message")
	fmt.Println("  exit          - Exit the shell")
	fmt.Println("  ls            - List files in the current directory")
	fmt.Println("  cd <dir>      - Change directory")
	fmt.Println("  wget <url>    - Download a file from the web")
	fmt.Println("  verfetch      - Fetch system version information")
	fmt.Println("  ip            - Print the main IP address")
	fmt.Println("  pkg install <user/repo> - Install a package from GitHub")
}

func verfetch(theme Theme) {
	v, _ := host.Info()
	cpuStat, _ := cpu.Info()
	vMem, _ := mem.VirtualMemory()
	diskStat, _ := disk.Usage("/")

	getColor(theme.TextColor).Println("OS: ", v.Platform)
	getColor(theme.TextColor).Println("OS Version: ", v.PlatformVersion)
	getColor(theme.TextColor).Println("Kernel: ", v.KernelVersion)
	getColor(theme.TextColor).Println("Architecture: ", v.KernelArch)
	getColor(theme.TextColor).Println("CPU: ", cpuStat[0].ModelName)
	getColor(theme.TextColor).Println("Cores: ", cpuStat[0].Cores)
	getColor(theme.TextColor).Println("Total Memory:  GB", bToGb(vMem.Total))
	getColor(theme.TextColor).Println("Available Memory:  GB", bToGb(vMem.Available))
	getColor(theme.TextColor).Println("Used Memory:  GB", bToGb(vMem.Used))
	getColor(theme.TextColor).Println("Disk Total:  GB", bToGb(diskStat.Total))
	getColor(theme.TextColor).Println("Disk Used:  GB", bToGb(diskStat.Used))
	getColor(theme.TextColor).Println("Disk Free:  GB", bToGb(diskStat.Free))
}

// Converts bytes to gigabytes
func bToGb(b uint64) uint64 {
	return b / (1024 * 1024 * 1024)
}

func now(theme Theme) {
	currentTime := time.Now()
	getColor(theme.TextColor).Println("Current time: ", currentTime.Format("15:04:05"))
}

func printMainIP(theme Theme) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		color.Red("Oops: \n", err.Error())
		return
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	color.Green("IP address: ", localAddr.IP.String())
}
