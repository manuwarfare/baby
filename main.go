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
		fmt.Println("Uso: ./baby <option>")
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
			fmt.Println("Error: Incorrect use of -n. Expected: baby -n <name> <command>")
			return
		}
		name := args[1]
		command := strings.Join(args[2:], " ")
		createRule(name, command)
	case "-r":
		if len(args) == 1 {
			fmt.Println("Error: Incorrect use of -r. Expected: baby -r <name> or baby -r a")
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
			fmt.Println("Error: Incorrect use of -c. Expected: baby -c <name> <command>")
			return
		}
		name := args[1]
		command := strings.Join(args[2:], " ")
		updateRule(name, command)
	case "-ln":
		if len(args) != 2 {
			fmt.Println("Error: Incorrect use of -ln. Expected: baby -ln <name>")
			return
		}
		name := args[1]
		showRule(name)
	default:
		if strings.HasPrefix(args[0], "-") {
			fmt.Println("Unknown option. Do baby -h to show help.")
		} else {
			runCommands(args)
		}
	}
}

func showHelp() {
	fmt.Println("Using: baby <option>")
	fmt.Println("Available options:")
	fmt.Println("-l\t\t\tList stored rules")
	fmt.Println("-n <name> <command>\tCreate a new rule")
	fmt.Println("-r <name>\t\tDelete a rule")
	fmt.Println("-r a \t\t\tDelete all rules")
	fmt.Println("-c <name> <command>\tUpdate a command")
	fmt.Println("-ln <name>\t\tList an specific rule")
	fmt.Println("-h\t\t\tShow the help")
	fmt.Println("-v\t\t\tShow the version of Baby")
	fmt.Println("Examples of use:")
	fmt.Println("Create a new rule: baby -n update 'sudo apt update -y'")
	fmt.Println("Run a block of rules: baby rule1 rule2")
}

func showVersion() {
	fmt.Printf("Baby version %s\n", version)
}

func listRules() {
	file, err := os.Open(configFile)
	if err != nil {
		fmt.Println("Unable to open config file:", err)
		fmt.Println("There are no rules in Baby, create the first one")
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
		fmt.Println("There are no rules yet in Baby")
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error opening the config file:", err)
	}
}

func createRule(name, command string) {
	file, err := os.OpenFile(configFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		fmt.Println("Error opening the config file:", err)
		return
	}
	defer file.Close()

	_, err = file.WriteString(fmt.Sprintf("%s = %s\n", name, command))
	if err != nil {
		fmt.Println("Error writing in config file:", err)
		return
	}
	fmt.Printf("Rule '%s' correctly added.\n", name)
}

func deleteRule(name string) {
	lines, err := readLines(configFile)
	if err != nil {
		fmt.Println("Error opening config file:", err)
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
		fmt.Printf("Unable to find the rule '%s'.\n", name)
		return
	}

	err = writeLines(configFile, lines)
	if err != nil {
		fmt.Println("Error writing in config file:", err)
		return
	}
	fmt.Printf("Rule '%s' deleted correctly.\n", name)
}

func deleteAllRules() {
	err := os.Remove(configFile)
	if err != nil {
		fmt.Println("Error deleting rules:", err)
		return
	}
	fmt.Println("All rules were deleted successfully.")
}

func updateRule(name, command string) {
	lines, err := readLines(configFile)
	if err != nil {
		fmt.Println("Error reading config file:", err)
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
		fmt.Printf("Unable to find the rule '%s'.\n", name)
		return
	}

	err = writeLines(configFile, lines)
	if err != nil {
		fmt.Println("Error writing the config file:", err)
		return
	}
	fmt.Printf("Rule'%s' correctly updated.\n", name)
}

func showRule(name string) {
	file, err := os.Open(configFile)
	if err != nil {
		fmt.Println("Error opening the config file:", err)
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
		fmt.Printf("The rule '%s' does not exist.\n", name)
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
		fmt.Println("There are no rules to run.")
		return
	}

	fullCommand := strings.Join(commandList, " && ")
	fmt.Println("Running:", fullCommand)

	err := executeCommand(fullCommand)
	if err != nil {
		fmt.Printf("Error executing commands: %s\n", err)
	}
}

func getCommand(name string) (string, error) {
	file, err := os.Open(configFile)
	if err != nil {
		return "", fmt.Errorf("unable to open the config file: %w", err)
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
		return "", fmt.Errorf("unable to read the config file: %w", err)
	}

	return "", fmt.Errorf("unable to find the rule '%s'", name)
}

func executeCommand(command string) error {
	fmt.Println("Running command:", command)

	cmd := exec.Command("bash", "-c", command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("error running command: %w", err)
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
