package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const configFile = "/etc/baby.conf"

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		fmt.Println("Usage: baby <option>")
		fmt.Println("")
		fmt.Println("Available options:")
		fmt.Println(" -l\t\t\tList stored rules")
		fmt.Println(" -n <name> <command>\tCreate a new rule")
		fmt.Println(" -r <name>\t\tDelete an existing rule")
		fmt.Println(" -r a \t\t\tDelete all rules")
		fmt.Println(" -c <name> <command>\tUpdate the command of a rule")
		fmt.Println(" -ln <name>\t\tShow the contents of a specific rule")
		fmt.Println(" -h\t\t\tShow this help")
		fmt.Println(" -v\t\t\tShow the program version")
		fmt.Println("")
		fmt.Println("Usage examples:")
		fmt.Println("Create a new rule: sudo baby -n update 'sudo apt update -y'")
		fmt.Println("The next time just run: 'baby update'")
		fmt.Println("")
		fmt.Println("This baby doesn't want to grow up!")
		return
	}

	switch args[0] {
	case "-h":
		showHelp()
	case "-l":
		listRules()
	case "-n":
		if len(args) < 3 {
			fmt.Println("Error: Incorrect usage of -n. It should be: baby -n <name> <command>")
			return
		}
		name := args[1]
		command := strings.Join(args[2:], " ")
		createRule(name, command)
	case "-r":
		if len(args) == 1 {
			fmt.Println("Error: Incorrect usage of -r. It should be: baby -r <name> or ./baby -r a")
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
			fmt.Println("Error: Incorrect usage of -c. It should be: baby -c <name> <command>")
			return
		}
		name := args[1]
		command := strings.Join(args[2:], " ")
		updateRule(name, command)
	case "-ln":
		if len(args) != 2 {
			fmt.Println("Error: Incorrect usage of -ln. It should be: baby -ln <name>")
			return
		}
		name := args[1]
		showRule(name)
	case "-v":
		fmt.Println("Baby version 1.0")
	default:
		if strings.HasPrefix(args[0], "-") {
			fmt.Println("Unrecognized option. Use baby -h to see the available options.")
		} else {
			runCommands(args)
		}
	}
}

func showHelp() {
	fmt.Println("Usage: baby <option>")
	fmt.Println(" ")
	fmt.Println("Available options:")
	fmt.Println(" -l\t\t\tList stored rules")
	fmt.Println(" -n <name> <command>\tCreate a new rule")
	fmt.Println(" -r <name>\t\tDelete an existing rule")
	fmt.Println(" -r a \t\t\tDelete all rules")
	fmt.Println(" -c <name> <command>\tUpdate the command of a rule")
	fmt.Println(" -ln <name>\t\tShow the contents of a specific rule")
	fmt.Println(" -h\t\t\tShow this help")
	fmt.Println(" -v\t\t\tShow the program version")
	fmt.Println("Usage examples:")
	fmt.Println("Create a new rule: sudo baby -n update 'sudo apt update -y'")
	fmt.Println("The next time just run: 'baby update'")
	fmt.Println(" ")
	fmt.Println("For further help go to https://salsa.debian.org/manuwarfare/baby")
	fmt.Println("Author: Manuel Guerra")
	fmt.Println("V 1.0 | This software is licensed under the GNU GPLv3")
}

func listRules() {
	file, err := os.Open(configFile)
	if err != nil {
		fmt.Println("Failed to open the configuration file:", err)
		fmt.Println("No rules have been created in Baby yet.")
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
		fmt.Println("No rules have been created in Baby yet.")
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading the configuration file:", err)
	}
}

func createRule(name, command string) {
	lines, err := readLines(configFile)
	if err != nil {
		fmt.Println("Error reading the configuration file:", err)
		return
	}

	for _, line := range lines {
		if strings.HasPrefix(line, name+" = ") {
			fmt.Printf("Rule '%s' already exists. Do you want to overwrite it? (y/n): ", name)
			var response string
			fmt.Scanln(&response)
			if response != "y" {
				fmt.Println("Operation cancelled.")
				return
			}
			break
		}
	}

	file, err := os.OpenFile(configFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		fmt.Println("Failed to open the configuration file:", err)
		return
	}
	defer file.Close()

	_, err = file.WriteString(fmt.Sprintf("%s = %s\n", name, command))
	if err != nil {
		fmt.Println("Error writing to the configuration file:", err)
		return
	}
	fmt.Printf("Rule '%s' successfully added.\n", name)
}

func deleteRule(name string) {
	lines, err := readLines(configFile)
	if err != nil {
		fmt.Println("Error reading the configuration file:", err)
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
		fmt.Printf("Rule '%s' not found.\n", name)
		return
	}

	err = writeLines(configFile, lines)
	if err != nil {
		fmt.Println("Error writing to the configuration file:", err)
		return
	}
	fmt.Printf("Rule '%s' successfully deleted.\n", name)
}

func deleteAllRules() {
	err := os.Remove(configFile)
	if err != nil {
		fmt.Println("Error deleting all rules:", err)
		return
	}
	fmt.Println("All rules have been successfully deleted.")
}

func updateRule(name, command string) {
	lines, err := readLines(configFile)
	if err != nil {
		fmt.Println("Error reading the configuration file:", err)
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
		fmt.Printf("Rule '%s' not found.\n", name)
		return
	}

	err = writeLines(configFile, lines)
	if err != nil {
		fmt.Println("Error writing to the configuration file:", err)
		return
	}
	fmt.Printf("Rule '%s' successfully updated.\n", name)
}

func showRule(name string) {
	file, err := os.Open(configFile)
	if err != nil {
		fmt.Println("Failed to open the configuration file:", err)
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
		fmt.Printf("Rule '%s' does not exist.\n", name)
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
		fmt.Println("No rules found to execute.")
		return
	}

	fullCommand := strings.Join(commandList, " && ")
	fmt.Println("Executing:", fullCommand)

	err := executeCommand(fullCommand)
	if err != nil {
		fmt.Printf("Error executing commands: %s\n", err)
	}
}

func getCommand(name string) (string, error) {
	file, err := os.Open(configFile)
	if err != nil {
		return "", fmt.Errorf("failed to open the configuration file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		ruleName := strings.TrimSpace(parts[0])
		command := strings.TrimSpace(parts[1])
		if ruleName == name {
			return command, nil
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading the configuration file: %w", err)
	}

	return "", fmt.Errorf("rule '%s' not found", name)
}

func executeCommand(cmd string) error {
	command := exec.Command("sh", "-c", cmd)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	command.Stdin = os.Stdin
	return command.Run()
}

func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
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

func writeLines(path string, lines []string) error {
	file, err := os.Create(path)
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
