package utils

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
	"unicode"

	gl "github.com/rafa-mori/gobe/logger"
)

// ValidateWorkerLimit valida o limite de workers
func ValidateWorkerLimit(value any) error {
	if limit, ok := value.(int); ok {
		if limit < 0 {
			return fmt.Errorf("worker limit cannot be negative")
		}
	} else {
		return fmt.Errorf("invalid type for worker limit")
	}
	return nil
}

func generateProcessFileName(processName string, pid int) string {
	bootID, err := GetBootID()
	if err != nil {
		gl.Log("error", fmt.Sprintf("Failed to get boot ID: %v", err))
		return ""
	}
	return fmt.Sprintf("%s_%d_%s.pid", processName, pid, bootID)
}

func createProcessFile(processName string, pid int) (*os.File, error) {
	fileName := generateProcessFileName(processName, pid)
	file, err := os.Create(fileName)
	if err != nil {
		return nil, err
	}

	// Escrever os detalhes do processo no arquivo
	_, err = file.WriteString(fmt.Sprintf("Process Name: %s\nPID: %d\nTimestamp: %d\n", processName, pid, time.Now().Unix()))
	if err != nil {
		file.Close()
		return nil, err
	}

	return file, nil
}

func removeProcessFile(file *os.File) {
	if file == nil {
		return
	}

	fileName := file.Name()
	file.Close()

	// Apagar o arquivo temporário
	if err := os.Remove(fileName); err != nil {
		gl.Log("error", fmt.Sprintf("Failed to remove process file %s: %v", fileName, err))
	} else {
		gl.Log("debug", fmt.Sprintf("Successfully removed process file: %s", fileName))
	}
}

func GetBootID() (string, error) {
	data, err := os.ReadFile("/proc/sys/kernel/random/boot_id")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

func GetBootTimeMac() (string, error) {
	cmd := exec.Command("sysctl", "-n", "kern.boottime")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func GetBootTimeWindows() (string, error) {
	cmd := exec.Command("powershell", "-Command", "(Get-WmiObject Win32_OperatingSystem).LastBootUpTime")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func IsBase64String(s string) bool {
	matched, _ := regexp.MatchString("^([A-Za-z0-9+/]{4})*([A-Za-z0-9+/]{3}=|[A-Za-z0-9+/]{2}==)?$", s)
	return matched
}

func IsBase64ByteSlice(s []byte) bool {
	matched, _ := regexp.Match("^([A-Za-z0-9+/]{4})*([A-Za-z0-9+/]{3}=|[A-Za-z0-9+/]{2}==)?$", s)
	return matched
}

func IsBase64ByteSliceString(s string) bool {
	matched, _ := regexp.Match("^([A-Za-z0-9+/]{4})*([A-Za-z0-9+/]{3}=|[A-Za-z0-9+/]{2}==)?$", []byte(s))
	return matched
}
func IsBase64ByteSliceStringWithPadding(s string) bool {
	matched, _ := regexp.Match("^([A-Za-z0-9+/]{4})*([A-Za-z0-9+/]{3}=|[A-Za-z0-9+/]{2}==)?$", []byte(s))
	return matched
}

func IsUrlEncodeString(s string) bool {
	matched, _ := regexp.MatchString("^[a-zA-Z0-9%_.-]+$", s)
	return matched
}
func IsUrlEncodeByteSlice(s []byte) bool {
	matched, _ := regexp.Match("^[a-zA-Z0-9%_.-]+$", s)
	return matched
}

func IsBase62String(s string) bool {
	if unicode.IsDigit(rune(s[0])) {
		return false
	}
	matched, _ := regexp.MatchString("^[a-zA-Z0-9_]+$", s)
	return matched

}
func IsBase62ByteSlice(s []byte) bool {
	matched, _ := regexp.Match("^[a-zA-Z0-9_]+$", s)
	return matched
}
