package types

import (
	"fmt"
	"io"
	"os"

	ci "github.com/rafa-mori/gobe/internal/contracts/interfaces"
	gl "github.com/rafa-mori/gobe/internal/module/logger"
)

func IsShellSpecialVar(c uint8) bool {
	switch c {
	case '*', '#', '$', '@', '!', '?', '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		return true
	}
	return false
}
func IsAlphaNum(c uint8) bool {
	return c == '_' || '0' <= c && c <= '9' || 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z'
}
func screeningByRAMSize(env ci.IEnvironment, filePath string) string {
	memAvailable := env.MemTotal()

	fileInfo, err := os.Stat(filePath)
	if err != nil {
		gl.Log("error", fmt.Sprintf("Erro ao obter tamanho do arquivo: %v", err))
		return "fallback"
	}

	// Se a memória disponível for menor que 100MB E arquivo for grande (>5MB), usa strings
	if memAvailable < 100 && fileInfo.Size() > 5*1024*1024 {
		return "strings"
	}
	return "json"
}
func asyncCopyFile(src, dst string) error {
	//go func() {
	_, err := copyFile(src, dst)
	if err != nil {
		fmt.Printf("Erro ao fazer backup do arquivo: %v\n", err)
	}
	//}()
	return nil
}
func copyFile(src, dst string) (int64, error) {
	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer func(source *os.File) {
		_ = source.Close()
	}(source)

	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer func(destination *os.File) {
		_ = destination.Close()
	}(destination)

	return io.Copy(destination, source)
}
