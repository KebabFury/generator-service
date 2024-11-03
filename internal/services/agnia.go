package services

import (
	"fmt"
	"github.com/KebabFury/generator-service/internal/config"
	"github.com/KebabFury/generator-service/internal/domain"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
)

type AgniaServiceImp struct {
	conf config.AgniaConfig
}

var re = regexp.MustCompile(`(?m)sys_name.+\n.*`)

func (a *AgniaServiceImp) RegisterProviders(providers []domain.Provider) {
	for _, provider := range providers {
		dir := "./team_actions/src/actions/" + provider.Name

		// Проверка и создание папки, если её нет
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			err := os.Mkdir(dir, 0755) // 0755 - права доступа к папке
			if err != nil {
				fmt.Println("Ошибка при создании папки:", err)
				return
			}
			fmt.Println("Папка успешно создана:", dir)
		} else {
			fmt.Println("Папка уже существует:", dir)
		}

		// Запись двух файлов в папку
		files := map[string]string{
			"__init__.py":      "",
			"actions.py":       re.ReplaceAllString(provider.ActionCode, ""),
			"documentation.md": provider.Documentation,
		}

		for name, content := range files {
			filePath := filepath.Join(dir, name)
			err := os.WriteFile(filePath, []byte(content), 0644) // 0644 - права доступа к файлу
			if err != nil {
				fmt.Println("Ошибка при записи файла:", err)
				return
			}
			fmt.Printf("Файл %s успешно записан\n", filePath)
		}

		// Выполнение команды python -m team_actions.src.initial_setup
		cmd := exec.Command("python", "-m", "team_actions.src.initial_setup")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			fmt.Println("Ошибка при выполнении команды:", err)
		} else {
			fmt.Println("Команда выполнена успешно.")
		}
	}

}

func NewAgniaServiceImp(agnia config.AgniaConfig) *AgniaServiceImp {
	return &AgniaServiceImp{
		agnia,
	}
}
