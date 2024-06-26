package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const (
	configFile = "baby.conf"
	version    = "1.0"
)

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		fmt.Println("Uso: ./baby <opción>")
		return
	}

	switch args[0] {
	case "-h":
		showHelp()
	case "-v":
		showVersion()
	case "-l":
		listRules()
	case "-n":
		if len(args) < 3 {
			fmt.Println("Error: Uso incorrecto de -n. Debe ser: ./baby -n <nombre> <comando>")
			return
		}
		name := args[1]
		command := strings.Join(args[2:], " ")
		createRule(name, command)
	case "-r":
		if len(args) == 1 {
			fmt.Println("Error: Uso incorrecto de -r. Debe ser: ./baby -r <nombre> o ./baby -r a")
			return
		}
		name := args[1]
		if name == "a" {
			deleteAllRules()
		} else {
			deleteRule(name)
		}
	case "-c":
		if len(args) < 3 {
			fmt.Println("Error: Uso incorrecto de -c. Debe ser: ./baby -c <nombre> <comando>")
			return
		}
		name := args[1]
		command := strings.Join(args[2:], " ")
		updateRule(name, command)
	case "-ln":
		if len(args) != 2 {
			fmt.Println("Error: Uso incorrecto de -ln. Debe ser: ./baby -ln <nombre>")
			return
		}
		name := args[1]
		showRule(name)
	default:
		if strings.HasPrefix(args[0], "-") {
			fmt.Println("Opción no reconocida. Usa ./baby -h para ver las opciones disponibles.")
		} else {
			runCommands(args)
		}
	}
}

func showHelp() {
	fmt.Println("Uso: ./baby <opción>")
	fmt.Println("Opciones disponibles:")
	fmt.Println("-l\t\t\tListar las reglas almacenadas")
	fmt.Println("-n <nombre> <comando>\tCrear una nueva regla")
	fmt.Println("-r <nombre>\t\tBorrar una regla existente")
	fmt.Println("-r a \t\t\tBorra todas las reglas")
	fmt.Println("-c <nombre> <comando>\tActualizar el comando de una regla")
	fmt.Println("-ln <nombre>\t\tLista el contenido de una regla específica")
	fmt.Println("-h\t\t\tMostrar esta ayuda")
	fmt.Println("-v\t\t\tMostrar la versión del programa")
	fmt.Println("Ejemplos de uso:")
	fmt.Println("Crea una nueva regla: baby -n update 'sudo apt update -y'")
	fmt.Println("Corre reglas en bloque: baby rule1 rule2")
}

func showVersion() {
	fmt.Printf("Baby versión %s\n", version)
}

func listRules() {
	file, err := os.Open(configFile)
	if err != nil {
		fmt.Println("No se pudo abrir el archivo de configuración:", err)
		fmt.Println("Aún no hay reglas creadas en Baby")
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var found bool
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		name := strings.TrimSpace(parts[0])
		command := strings.TrimSpace(parts[1])
		fmt.Printf("%s = %s\n", name, command)
		found = true
	}

	if !found {
		fmt.Println("Aún no hay reglas creadas en Baby")
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error al leer el archivo de configuración:", err)
	}
}

func createRule(name, command string) {
	file, err := os.OpenFile(configFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		fmt.Println("No se pudo abrir el archivo de configuración:", err)
		return
	}
	defer file.Close()

	_, err = file.WriteString(fmt.Sprintf("%s = %s\n", name, command))
	if err != nil {
		fmt.Println("Error al escribir en el archivo de configuración:", err)
		return
	}
	fmt.Printf("Regla '%s' añadida correctamente.\n", name)
}

func deleteRule(name string) {
	lines, err := readLines(configFile)
	if err != nil {
		fmt.Println("Error al leer el archivo de configuración:", err)
		return
	}

	found := false
	for i, line := range lines {
		if strings.HasPrefix(line, name+" = ") {
			lines = append(lines[:i], lines[i+1:]...)
			found = true
			break
		}
	}

	if !found {
		fmt.Printf("No se encontró la regla '%s'.\n", name)
		return
	}

	err = writeLines(configFile, lines)
	if err != nil {
		fmt.Println("Error al escribir en el archivo de configuración:", err)
		return
	}
	fmt.Printf("Regla '%s' eliminada correctamente.\n", name)
}

func deleteAllRules() {
	err := os.Remove(configFile)
	if err != nil {
		fmt.Println("Error al borrar todas las reglas:", err)
		return
	}
	fmt.Println("Todas las reglas han sido eliminadas correctamente.")
}

func updateRule(name, command string) {
	lines, err := readLines(configFile)
	if err != nil {
		fmt.Println("Error al leer el archivo de configuración:", err)
		return
	}

	found := false
	for i, line := range lines {
		if strings.HasPrefix(line, name+" = ") {
			lines[i] = fmt.Sprintf("%s = %s", name, command)
			found = true
			break
		}
	}

	if !found {
		fmt.Printf("No se encontró la regla '%s'.\n", name)
		return
	}

	err = writeLines(configFile, lines)
	if err != nil {
		fmt.Println("Error al escribir en el archivo de configuración:", err)
		return
	}
	fmt.Printf("Regla '%s' actualizada correctamente.\n", name)
}

func showRule(name string) {
	file, err := os.Open(configFile)
	if err != nil {
		fmt.Println("No se pudo abrir el archivo de configuración:", err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	found := false
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, name+" = ") {
			fmt.Println(line)
			found = true
			break
		}
	}

	if !found {
		fmt.Printf("La regla '%s' no existe.\n", name)
	}
}

func runCommands(commands []string) {
	var commandList []string

	for _, cmd := range commands {
		rule, err := getCommand(cmd)
		if err != nil {
			fmt.Printf("Error: %s\n", err)
			continue
		}
		commandList = append(commandList, rule)
	}

	if len(commandList) == 0 {
		fmt.Println("No se encontraron reglas para ejecutar.")
		return
	}

	fullCommand := strings.Join(commandList, " && ")
	fmt.Println("Ejecutando:", fullCommand)

	err := executeCommand(fullCommand)
	if err != nil {
		fmt.Printf("Error al ejecutar los comandos: %s\n", err)
	}
}

func getCommand(name string) (string, error) {
	file, err := os.Open(configFile)
	if err != nil {
		return "", fmt.Errorf("no se pudo abrir el archivo de configuración: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, name+" = ") {
			return strings.TrimSpace(strings.SplitN(line, "=", 2)[1]), nil
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error al leer el archivo de configuración: %w", err)
	}

	return "", fmt.Errorf("no se encontró una regla asociada a '%s'", name)
}

func executeCommand(command string) error {
	fmt.Println("Ejecutando comando:", command)

	cmd := exec.Command("bash", "-c", command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("error al ejecutar el comando: %w", err)
	}

	return nil
}

func readLines(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

func writeLines(filename string, lines []string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, line := range lines {
		_, err := writer.WriteString(line + "\n")
		if err != nil {
			return err
		}
	}
	return writer.Flush()
}
