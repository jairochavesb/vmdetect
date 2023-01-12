package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/intel-go/cpuid"
)

type ArtifactList struct {
	vmStrings     string
	nicMacaddress string
	files         string
	directories   string
	cpuIDs        string
}

const (
	RED    = "\033[0;31m"
	GREEN  = "\033[0;32m"
	YELLOW = "\033[0;33m"
	WHITE  = "\033[1;37m"
	NORMAL = "\033[0m"
)

func main() {
	var artifacts = ArtifactList{}

	if len(os.Args) == 1 {
		showHelp(os.Args[0])

	} else if os.Args[1] == "-a" {
		showAbout()

	} else {
		artifacts = loadArtifacts(os.Args[1])
	}

	fmt.Printf("\033[H\033[2J")

	showBanner()

	fmt.Printf("%s[*] SEARCHING FOR HYPERVISOR NAMES IN FILES %s\n", WHITE, NORMAL)

	fileList := strings.Split(artifacts.files, ",")
	for _, file := range fileList {
		searchFileArtifacts(file, artifacts.vmStrings)
	}

	dirList := strings.Split(artifacts.directories, ",")
	for _, dir := range dirList {
		searchDirArtifacts(dir, artifacts.vmStrings)
	}

	fmt.Printf("\n%s[*] CHECKING SYSTEM INFORMATION %s\n", WHITE, NORMAL)
	getMacAddresses(artifacts.nicMacaddress)
	getSystemUptime()
	getRamSize()
	getCPUID(artifacts.cpuIDs)

	testCPUExtraFeatures()

}

func loadArtifacts(fileName string) ArtifactList {
	data, err := os.Open(fileName)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Can't open the file: ", err.Error())
	}
	defer data.Close()

	artifacts := new(ArtifactList)

	scanner := bufio.NewScanner(data)

	for scanner.Scan() {

		if strings.Contains(scanner.Text(), "vm_strings=") {
			artifacts.vmStrings = strings.Split(scanner.Text(), "=")[1]

		} else if strings.Contains(scanner.Text(), "mac_address=") {
			artifacts.nicMacaddress = strings.Split(scanner.Text(), "=")[1]

		} else if strings.Contains(scanner.Text(), "files=") {
			artifacts.files = strings.Split(scanner.Text(), "=")[1]

		} else if strings.Contains(scanner.Text(), "directories=") {
			artifacts.directories = strings.Split(scanner.Text(), "=")[1]

		} else if strings.Contains(scanner.Text(), "cpu_ids=") {
			artifacts.cpuIDs = strings.Split(scanner.Text(), "=")[1]
		}
	}

	return *artifacts
}

func testCPUExtraFeatures() {
	fmt.Printf("Hypervisor feature:")

	feaure := cpuid.HasFeature(cpuid.SVM)
	fmt.Println(feaure)
}

func getCPUID(cpuIds string) {

	ids := strings.Split(cpuIds, ",")

	vendorID := strings.ToLower(cpuid.VendorIdentificatorString)

	var match = ""

	for _, id := range ids {
		id = strings.ToLower(id)

		if vendorID == id {
			match = id
			break
		}
	}

	padding := strings.Repeat(" ", 45-len(" CPU ID "))

	if match != "" {
		fmt.Printf("    - CPU ID %s%s %s%s\n", padding, RED, match, NORMAL)
	} else {
		fmt.Printf("    - CPU ID %s%s%s%s\n", padding, GREEN, " Clean", NORMAL)
	}
}

func getRamSize() {
	data, err := os.Open("/proc/meminfo")
	if err != nil {
		fmt.Fprintf(os.Stderr, "      %s%s%s\n", YELLOW, err.Error(), NORMAL)
	}
	defer data.Close()

	scanner := bufio.NewScanner(data)
	memTotalStr := ""

	for scanner.Scan() {
		if strings.Contains(scanner.Text(), "MemTotal") {
			for _, c := range scanner.Text() {
				if c >= 48 && c <= 57 {
					memTotalStr += string(c)
				}
			}
			break
		}
	}

	memTotal, err := strconv.Atoi(memTotalStr)
	memTotal = memTotal / 976600

	if err != nil {
		fmt.Fprintf(os.Stderr, "      %s%s%s\n", YELLOW, err.Error(), NORMAL)
	}

	padding := strings.Repeat(" ", 44-len("Ram Size "))

	if memTotal < 8 {
		fmt.Printf("    - Ram Size  %s%s%d Gb%s\n", padding, RED, memTotal, NORMAL)

	} else if memTotal >= 8 && memTotal < 12 {
		fmt.Printf("    - Ram Size  %s%s%d Gb%s\n", padding, YELLOW, memTotal, NORMAL)

	} else {
		fmt.Printf("    - Ram Size  %s%s%d Gb%s\n", padding, GREEN, memTotal, NORMAL)
	}
}

func getSystemUptime() {
	data, err := os.Open("/proc/uptime")
	if err != nil {
		fmt.Fprintf(os.Stderr, "      %s%s%s\n", YELLOW, err.Error(), NORMAL)
	}
	defer data.Close()

	var uptime_str = ""

	scanner := bufio.NewScanner(data)
	for scanner.Scan() {
		uptime_str = strings.Split(scanner.Text(), " ")[0]
	}

	uptime, err := strconv.ParseFloat(uptime_str, 32)
	if err != nil {
		fmt.Fprintf(os.Stderr, "      %s%s%s\n", YELLOW, err.Error(), NORMAL)
	}

	padding := strings.Repeat(" ", 44-len("uptime"))

	if uptime < 3600 {
		fmt.Printf("    - Uptime %s%s%.2f minutes%s\n", padding, RED, uptime/60, NORMAL)

	} else if uptime >= 3600 && uptime <= 17999 {
		fmt.Printf("    - Uptime %s%s%.2f hours%s\n", padding, YELLOW, uptime/3600, NORMAL)

	} else {
		fmt.Printf("    - Uptime %s%s%.2f hours%s\n", padding, GREEN, uptime/3600, NORMAL)
	}
}

func getMacAddresses(artifacts string) {
	artifact_list := strings.Split(artifacts, ",")

	ifaces, err := net.Interfaces()
	if err != nil {
		fmt.Fprintf(os.Stderr, "      %s%s%s\n", YELLOW, err.Error(), NORMAL)
	}

	var match = ""

	for _, iface := range ifaces {

		if iface.Name == "lo" {
			continue
		}

		for _, artifact := range artifact_list {
			if strings.Contains(strings.ToLower(iface.HardwareAddr.String()), strings.ToLower(artifact)) {
				match = artifact
				break
			}
		}

		padding := strings.Repeat(" ", 45-len(iface.Name)-len(" mac address"))

		if match != "" {
			fmt.Printf("    - %s mac address%s%sFound: %s%s\n", iface.Name, padding, RED, match, NORMAL)
		} else {
			fmt.Printf("    - %s mac address%s%sClean%s\n", iface.Name, padding, GREEN, NORMAL)
		}
	}
}

func searchDirArtifacts(dirName, artifacts string) {
	artifact_list := strings.Split(artifacts, ",")

	dir, err := os.Open(dirName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "      %s%s%s\n", YELLOW, err.Error(), NORMAL)
	}
	defer dir.Close()

	files, err := dir.ReadDir(-1)
	if err != nil {
		fmt.Fprintf(os.Stderr, "      %s%s%s\n", YELLOW, err.Error(), NORMAL)
	}

	var match = ""

	for _, file := range files {
		for _, artifact := range artifact_list {
			if strings.Contains(strings.ToLower(file.Name()), artifact) {
				match = artifact
				break
			}
		}
	}

	padding := strings.Repeat(" ", 45-len(dirName))

	if match != "" {
		fmt.Printf("    - %s%s%sFound: %s%s\n", dirName, padding, RED, match, NORMAL)
	} else {
		fmt.Printf("    - %s%s%sClean%s\n", dirName, padding, GREEN, NORMAL)
	}
}

func searchFileArtifacts(fileName, artifacs string) {
	artifact_list := strings.Split(artifacs, ",")

	data, err := os.Open(fileName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "      %s%s%s\n", YELLOW, err.Error(), NORMAL)
		return
	}
	defer data.Close()

	scanner := bufio.NewScanner(data)
	var match = ""

	for scanner.Scan() {
		text_line := strings.ToLower(scanner.Text())

		for _, artifact := range artifact_list {
			if strings.Contains(text_line, artifact) {
				match = artifact
				break
			}
		}
	}

	padding := strings.Repeat(" ", 45-len(fileName))

	if match != "" {
		fmt.Printf("    - %s%s%sFound: %s%s\n", fileName, padding, RED, match, NORMAL)
	} else {
		fmt.Printf("    - %s%s%sClean%s\n", fileName, padding, GREEN, NORMAL)
	}

}

func showHelp(prog_name string) {
	fmt.Println("USAGE:")
	fmt.Println(prog_name + "[<FILE_NAME>] [-a]")
	fmt.Println("OPTIONS:")
	fmt.Println("<FILE_NAME>: File containing the VM artifacts.")

	os.Exit(1)
}

func showBanner() {
	fmt.Println("██    ██ ███    ███     ██████  ███████ ████████ ███████  ██████ ████████ ")
	fmt.Println("██    ██ ████  ████     ██   ██ ██         ██    ██      ██         ██    ")
	fmt.Println("██    ██ ██ ████ ██     ██   ██ █████      ██    █████   ██         ██    ")
	fmt.Println(" ██  ██  ██  ██  ██     ██   ██ ██         ██    ██      ██         ██    ")
	fmt.Println("  ████   ██      ██     ██████  ███████    ██    ███████  ██████    ██    ")
	fmt.Println("")
}

func showAbout() {
	fmt.Println("Sandbox Detect is a tool to check if it is running inside a vm or not.")
	fmt.Println("This tool use same techniques as implemented by malware to detect")
	fmt.Println("the presence of virtualization/analysis environments.")

	fmt.Println("Author: Jairo Chaves Brenes")
	fmt.Println("Dec 22, 2022")

	os.Exit(0)
}
