package main

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"io"
	"path/filepath"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

const (
    configDir = "/.config/baby/"
    configFileName = "baby.conf"
)

var configFile = filepath.Join(os.Getenv("HOME"), configDir, configFileName)

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		showHelp()
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
			fmt.Println("Error: Incorrect usage of -r. It should be: baby -r <name> [<name>...] or baby -r a")
			return
		}
		names := args[1:]
		if len(names) == 1 && names[0] == "a" {
			deleteAllRules()
		} else {
			for _, name := range names {
				deleteRule(name)
			}
		}
	case "-c":
		if len(args) < 3 {
			fmt.Println("Error: Incorrect usage of -c. It should be: baby -c <name> '<command>'")
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
	case "-i":
        if len(args) != 2 {
            fmt.Println("Error: Incorrect usage of -i. It should be: baby -i <url or file path>")
            return
        }
        importSource := args[1]
        if strings.HasPrefix(importSource, "http://") || strings.HasPrefix(importSource, "https://") {
            importRulesFromURL(importSource)
        } else {
            importRulesFromFile(importSource)
        }
	case "-e":
		exportRules()
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
	fmt.Println(" -n <name> '<command>'\tCreate a new rule")
	fmt.Println(" -r <name> [<name>...]\tDelete existing rules")
	fmt.Println(" -r a \t\t\tDelete all rules")
	fmt.Println(" -c <name> '<command>'\tUpdate the command of a rule")
	fmt.Println(" -ln <name>\t\tShow the contents of a specific rule")
	fmt.Println(" -h\t\t\tShow this help")
	fmt.Println(" -v\t\t\tShow the program version")
	fmt.Println(" -i <url or path>\tImport rules from a URL or file")
	fmt.Println(" -e\t\t\tExport rules to a text file")
	fmt.Println(" ")
	fmt.Println("Usage examples:")
	fmt.Println("Create a new rule: baby -n update 'sudo apt update -y'")
	fmt.Println("The next time just run: 'baby update'")
	fmt.Println(" ")
	fmt.Println("For further help go to https://github.com/manuwarfare/baby")
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

	found := false
	for i, line := range lines {
		if strings.HasPrefix(line, name+" = ") {
			fmt.Printf("Rule '%s' already exists. Do you want to overwrite it? (y/n): ", name)
			var response string
			fmt.Scanln(&response)
			if response != "y" {
				fmt.Println("Operation cancelled.")
				return
			}
			lines[i] = fmt.Sprintf("%s = %s", name, command)
			found = true
			break
		}
	}

	if !found {
		lines = append(lines, fmt.Sprintf("%s = %s", name, command))
	}

	err = writeLines(configFile, lines)
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

func deleteAllRules() error {
    file, err := os.OpenFile(configFile, os.O_WRONLY, 0644)
    if err != nil {
        return fmt.Errorf("failed to open baby.conf: %v", err)
    }
    defer file.Close()

    err = file.Truncate(0)
    if err != nil {
        return fmt.Errorf("failed to truncate baby.conf: %v", err)
    }

    fmt.Println("All rules have been successfully deleted.")
    return nil
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
        processedRule := processBottles(rule)
        commandList = append(commandList, processedRule)
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
		return "", fmt.Errorf("failed to open the configuration file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, name+" = ") {
			return strings.TrimSpace(strings.TrimPrefix(line, name+" = ")), nil
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading the configuration file: %v", err)
	}

	return "", fmt.Errorf("rule '%s' not found", name)
}

func importRulesFromURL(url string) {
    resp, err := http.Get(url)
    if err != nil {
        fmt.Println("Error fetching rules from URL:", err)
        return
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        fmt.Println("Failed to fetch rules from URL. Status code:", resp.StatusCode)
        return
    }

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        fmt.Println("Error reading response body:", err)
        return
    }

    rulesText := string(body)

    // Use HTML unescaping to convert special characters
    rulesText = html.UnescapeString(rulesText)

    // Process the rules
    rules := extractRules(rulesText)

    // Import the rules into baby.conf
    existingRules := make(map[string]bool)
    fileLines, err := readLines(configFile)
    if err == nil {
        for _, line := range fileLines {
            parts := strings.SplitN(line, "=", 2)
            if len(parts) == 2 {
                existingRules[strings.TrimSpace(parts[0])] = true
            }
        }
    }

    for _, rule := range rules {
        parts := strings.Split(rule, " = ")
        if len(parts) != 2 {
            fmt.Println("Error parsing rule:", rule)
            continue
        }
        name := strings.TrimSpace(parts[0])
        command := strings.TrimSpace(parts[1])

        // Check if the rule already exists
        if existingRules[name] {
            fmt.Printf("Rule '%s' already exists. Do you want to overwrite it? (y/n): ", name)
            var response string
            fmt.Scanln(&response)
            if response != "y" {
                fmt.Printf("Skipping rule '%s'.\n", name)
                continue
            }
            updateRule(name, command)
        } else {
            createRule(name, command)
        }
    }
}

func importRulesFromFile(filePath string) {
    file, err := os.Open(filePath)
    if err != nil {
        fmt.Println("Error opening file:", err)
        return
    }
    defer file.Close()

    scanner := bufio.NewScanner(file)
    var rulesText string
    for scanner.Scan() {
        rulesText += scanner.Text() + "\n"
    }
    if err := scanner.Err(); err != nil {
        fmt.Println("Error reading file:", err)
        return
    }

    rules := extractRules(rulesText)

    // Import the rules into baby.conf
    existingRules := make(map[string]bool)
    fileLines, err := readLines(configFile)
    if err == nil {
        for _, line := range fileLines {
            parts := strings.SplitN(line, "=", 2)
            if len(parts) == 2 {
                existingRules[strings.TrimSpace(parts[0])] = true
            }
        }
    }

    for _, rule := range rules {
        parts := strings.Split(rule, " = ")
        if len(parts) != 2 {
            fmt.Println("Error parsing rule:", rule)
            continue
        }
        name := strings.TrimSpace(parts[0])
        command := strings.TrimSpace(parts[1])

        // Check if the rule already exists
        if existingRules[name] {
            fmt.Printf("Rule '%s' already exists. Do you want to overwrite it? (y/n): ", name)
            var response string
            fmt.Scanln(&response)
            if response != "y" {
                fmt.Printf("Skipping rule '%s'.\n", name)
                continue
            }
            updateRule(name, command)
        } else {
            createRule(name, command)
        }
    }
}

func extractRules(text string) []string {
	var rules []string

	re := regexp.MustCompile(`b:([^=]+) = (.*?):b`)
	matches := re.FindAllStringSubmatch(text, -1)
	for _, match := range matches {
		ruleName := strings.TrimSpace(match[1])
		ruleCommand := strings.TrimSpace(match[2])

		// Replace HTML entities with their actual characters
		ruleCommand = html.UnescapeString(ruleCommand)

		rule := fmt.Sprintf("%s = %s", ruleName, ruleCommand)
		rules = append(rules, rule)
	}

	return rules
}

func exportRules() {
	fmt.Println("Exporting rules in progress... Press ctrl+c to quit")
	fmt.Println("You can export rules in bulk, e.g., <rule1> <rule2>")

	var exportRules []string
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Println("Which rule(s) do you want to export? Leave blank to export all:")
		scanner.Scan()
		text := scanner.Text()

		if text == "" {
			exportRules = getAllRules()
			break
		} else {
			rules := strings.Fields(text)
			exportRules = nil // Reset the exportRules slice for re-selection
			var invalidRules []string
			for _, rule := range rules {
				if ruleExists(rule) {
					exportRules = append(exportRules, rule)
				} else {
					invalidRules = append(invalidRules, rule)
				}
			}

			if len(invalidRules) > 0 {
				fmt.Printf("The following rules were not found: %v\n", invalidRules)
				fmt.Println("Please re-enter the correct rules or leave blank to export all.")
			} else {
				break
			}
		}
	}

	if len(exportRules) == 0 {
		fmt.Println("No valid rules selected for export.")
		return
	}

	fmt.Println("Do you want to add a comment? Leave blank to continue:")
	scanner.Scan()
	comment := scanner.Text()

	// Prepare export content
	var exportContent []string
	if comment != "" {
		exportContent = append(exportContent, fmt.Sprintf("#%s", comment))
	}

	for _, rule := range exportRules {
		command, err := getCommand(rule)
		if err != nil {
			fmt.Printf("Error getting command for rule '%s': %v\n", rule, err)
			continue
		}
		exportContent = append(exportContent, fmt.Sprintf("b:%s = %s:b", rule, command))
	}

	for {
		fmt.Println("Where do you want to store your file? Leave blank to store in $HOME")
		fmt.Println("Select a folder for your file:")
		scanner.Scan()
		exportPath := scanner.Text()

		if exportPath == "" {
			exportPath = os.Getenv("HOME")
		}

		// Check if the path is valid
		fileInfo, err := os.Stat(exportPath)
		if err != nil || !fileInfo.IsDir() {
			fmt.Println("Location not found or not a directory.")
			continue
		}

		// Write to file
		exportFilePath := fmt.Sprintf("%s/baby-rules.txt", exportPath)
		err = writeToFile(exportFilePath, exportContent)
		if err != nil {
			fmt.Println("Error writing rules to file:", err)
			return
		}

		fmt.Printf("Rules successfully exported to: %s\n", exportFilePath)
		break
	}
}

func ruleExists(name string) bool {
	file, err := os.Open(configFile)
	if err != nil {
		return false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, name+" = ") {
			return true
		}
	}

	return false
}

func getAllRules() []string {
	var rules []string

	file, err := os.Open(configFile)
	if err != nil {
		fmt.Println("Failed to open the configuration file:", err)
		return rules
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			rules = append(rules, strings.TrimSpace(parts[0]))
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading the configuration file:", err)
	}

	return rules
}

func writeToFile(filePath string, content []string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, line := range content {
		_, err := fmt.Fprintln(writer, line)
		if err != nil {
			return fmt.Errorf("failed to write to file: %v", err)
		}
	}
	if err := writer.Flush(); err != nil {
		return fmt.Errorf("failed to flush writer: %v", err)
	}

	return nil
}

func executeCommand(command string) error {
	cmd := exec.Command("bash", "-c", command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to execute command: %v", err)
	}
	return nil
}

func processBottles(command string) string {
    re := regexp.MustCompile(`b%\('([^']+)'\)%b`)
    return re.ReplaceAllStringFunc(command, func(match string) string {
        bottleName := re.FindStringSubmatch(match)[1]
        fmt.Printf("The %s is?: ", bottleName)
        var value string
        fmt.Scanln(&value)
        return value
    })
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
		// Write the line after unquoting it to interpret escaped characters
		unquotedLine, err := strconv.Unquote(`"` + line + `"`)
		if err != nil {
			fmt.Println("Error unquoting line:", line, "Error:", err)
			return err
		}
		_, err = fmt.Fprintln(writer, unquotedLine)
		if err != nil {
			return err
		}
	}
	return writer.Flush()
}
