package main

import (
	"bufio"
	"fmt"
	"os"
	"html"
	"net"
	"path/filepath"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
	"log"

	"golang.org/x/sys/unix"

)

const (
    configDir = "/.config/baby/"
    configFileName = "baby.conf"
    logDir = "/.local/share/baby/"
    logFileName = "baby.log"
    VERSION = "1.0.55"
)

var configFile = filepath.Join(os.Getenv("HOME"), configDir, configFileName)

var reservedNames = []string{
    "-h", "-l", "-n", "-r", "-c", "-ln", "-v", "-i", "-e", "-b",
    "-H", "-L", "-N", "-R", "-C", "-LN", "-V", "-I", "-E", "-B",
    "-lN", "-Ln",

    // Reserved for future implementations
    "-g", "-G", "-w", "-W", "-t", "-T", "-x", "-X", "-y", "-Y",
    "-z", "-Z", "-a", "-A",
}

func main() {

    // Initialize the config file
    homeDir, err := os.UserHomeDir()
    if err != nil {
        log.Fatalf("Failed to get home directory: %v", err)
    }
    configFile = filepath.Join(homeDir, ".config", "baby", "baby.conf")

    err = initConfigFile()
    if err != nil {
        log.Fatalf("Failed to initialize config file: %v", err)
    }

    args := os.Args[1:]

    bottleValues := make(map[string]string)
    var commands []string

    for i := 0; i < len(args); i++ {
        if strings.HasPrefix(args[i], "-b=") {
            parts := strings.SplitN(args[i], "=", 2)
            if len(parts) == 2 {
                bottleParts := strings.SplitN(parts[1], ":", 2)
                if len(bottleParts) == 2 {
                    bottleValues[bottleParts[0]] = bottleParts[1]
                }
            }
        } else {
            commands = append(commands, args[i])
        }
    }

    if len(commands) == 0 {
        showHelp()
        return
    }

    switch commands[0] {
    case "-h":
        showHelp()
    case "-l":
        listRules()
    case "-n":
        if len(commands) < 3 {
            fmt.Println("Error: Incorrect usage of -n. It should be: baby -n <name> '<command>'")
            return
        }
        name := commands[1]
        command := strings.Join(commands[2:], " ")
        createRule(name, command)
    case "-r":
        if len(commands) == 1 {
            fmt.Println("Error: Incorrect usage of -r. It should be: baby -r <name> [<name>...] or baby -r a")
            return
        }
        names := commands[1:]
        if len(names) == 1 && names[0] == "a" {
            deleteAllRules()
        } else {
            for _, name := range names {
                deleteRule(name)
            }
        }
    case "-c":
        if len(commands) < 3 {
            fmt.Println("Error: Incorrect usage of -c. It should be: baby -c <name> '<command>'")
            return
        }
        name := commands[1]
        command := strings.Join(commands[2:], " ")
        updateRule(name, command)
    case "-ln":
        if len(commands) != 2 {
            fmt.Println("Error: Incorrect usage of -ln. It should be: baby -ln <name>")
            return
        }
        name := commands[1]
        showRule(name)
    case "-v":
        fmt.Println("Baby version", VERSION)
    case "-i":
        if len(commands) != 2 {
            fmt.Println("Error: Incorrect usage of -i. It should be: baby -i <file path>")
            return
        }
        importSource := commands[1]
        importRulesFromFile(importSource)
    case "-e":
        exportRules()
    default:
        if strings.HasPrefix(commands[0], "-") {
            fmt.Println("Unrecognized option. Use baby -h to see the available options.")
        } else {
            runCommands(commands, bottleValues)
        }
    }
}

func showHelp() {
    fmt.Println("Usage: baby <option>")
    fmt.Println(" ")
    fmt.Println("Available options:")
    fmt.Println(" -n <name> '<command>'\tCreate a new rule")
    fmt.Println(" -l\t\t\tList stored rules")
    fmt.Println(" -r <name> [<name>...]\tDelete existing rules")
    fmt.Println(" -r a \t\t\tDelete all rules")
    fmt.Println(" -c <name> '<command>'\tUpdate the command of a rule")
    fmt.Println(" -ln <name>\t\tShow the contents of a specific rule")
    fmt.Println(" -h\t\t\tShow this help")
    fmt.Println(" -v\t\t\tShow the program version")
    fmt.Println(" -i <file path>\t\tImport rules from a local file")
    fmt.Println(" -e\t\t\tExport rules to a text file (backup)")
    fmt.Println(" -b=<variable:value>\tPre-define the content of a bottle")
    fmt.Println("\t\t\tSyntax for create bottles: b%('variable')%b")
    fmt.Println(" ")
    fmt.Println("Usage examples:")
    fmt.Println(" Create a new rule: baby -n update 'sudo apt update -y'")
    fmt.Println(" The next time just run: baby update")
    fmt.Println(" ")
    fmt.Println(" Create a new rule with bottle: baby -n ssh 'ssh -p 2222 b%('username')%b@example.com'")
    fmt.Println(" The next time you run 'baby ssh' the system will ask you for the username value")
    fmt.Println(" ")
    fmt.Println("For further help go to https://github.com/manuwarfare/baby")
    fmt.Println("Author: Manuel Guerra")
    fmt.Printf("V %s | This software is licensed under the GNU GPLv3\n", VERSION)
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

    if isReservedName(name) {
        fmt.Printf("Unable to create a rule with this name. '%s' is a reserved command name.\n", name)
        return
    }

    found := false
    for i, line := range lines {
        if strings.HasPrefix(line, name+" = ") {
            fmt.Printf("The rule '%s' already exists. Do you want to overwrite it? (y/n): ", name)
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

    // write the events in baby.log
    err = logEvent("CREATE_RULE", fmt.Sprintf("Name: %s, Command: %s", name, command))
    if err != nil {
        fmt.Printf("Warning: Failed to log event: %v\n", err)
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

    // write events in baby.log
    err = logEvent("DELETE_RULE", fmt.Sprintf("Name: %s", name))
    if err != nil {
        fmt.Printf("Warning: Failed to log event: %v\n", err)
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
	err := initConfigFile()
    if err != nil {
        fmt.Printf("Error initializing config file: %v\n", err)
        return
    }

    lines, err := readLines(configFile)
    if err != nil {
        fmt.Println("Error reading the configuration file:", err)
        return
    }

    if isReservedName(name) {
        fmt.Printf("Unable to update rule. '%s' is a reserved command name.\n", name)
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

    // write events in baby.log
    err = logEvent("UPDATE_RULE", fmt.Sprintf("Name: %s, New Command: %s", name, command))
    if err != nil {
        fmt.Printf("Warning: Failed to log event: %v\n", err)
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

func runCommands(commands []string, bottleValues map[string]string) {
    var processedCommands []string
    for _, cmd := range commands {
        rule, err := getCommand(cmd)
        if err != nil {
            fmt.Printf("Error: %s\n", err)
            continue
        }
        processedRule := processBottles(rule, bottleValues)
        processedCommands = append(processedCommands, processedRule)
    }
    if len(processedCommands) == 0 {
        fmt.Println("No rules found to execute.")
        return
    }
    for i, command := range processedCommands {
        start := time.Now()
        fmt.Printf("Executing command %d: %s\n", i+1, command)
        err := executeCommand(command)
        duration := time.Since(start)

        result := "Success"
        if err != nil {
            result = fmt.Sprintf("Error: %v", err)
            fmt.Printf("Error executing command %d: %s\n", i+1, err)
        }

        logDetails := fmt.Sprintf("Command: \"%s\", Result: %s in %v", command, result, duration)
        err = logEvent("EXECUTE_COMMAND", logDetails)
        if err != nil {
            fmt.Printf("Warning: Failed to log event: %v\n", err)
        }
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

    // Read existing rules
    existingRules, err := readLines(configFile)
    if err != nil {
        fmt.Println("Error reading existing rules:", err)
        return
    }

    // Process rules
    for _, rule := range rules {
        parts := strings.Split(rule, " = ")
        if len(parts) != 2 {
            fmt.Println("Error parsing rule:", rule)
            continue
        }
        name := strings.TrimSpace(parts[0])
        command := strings.TrimSpace(parts[1])

        // Check if the rule already exists
        exists := false
        for i, existingRule := range existingRules {
            if strings.HasPrefix(existingRule, name+" = ") {
                exists = true
                fmt.Printf("Rule '%s' already exists. Do you want to overwrite it? (y/n): ", name)
                var response string
                fmt.Scanln(&response)
                if response == "y" {
                    existingRules[i] = fmt.Sprintf("%s = %s", name, command)
                    fmt.Printf("Rule '%s' updated.\n", name)
                } else {
                    fmt.Printf("Skipping rule '%s'.\n", name)
                }
                break
            }
        }

        if !exists {
            existingRules = append(existingRules, fmt.Sprintf("%s = %s", name, command))
            fmt.Printf("Rule '%s' added.\n", name)
        }

        // Log the import event
        err := logEvent("IMPORT_RULE", fmt.Sprintf("From File: %s, Name: %s, Command: %s", filePath, name, command))
        if err != nil {
            fmt.Printf("Warning: Failed to log event: %v\n", err)
        }
    }

    // Write all rules back to the config file
    //err = writeLines(configFile, existingRules)
    err = writeLinesWithLock(configFile, existingRules)
    if err != nil {
        fmt.Println("Error writing rules to config file:", err)
        return
    }

    fmt.Println("Rules imported successfully.")
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
            // Reset the exportRules slice for re-selection
			exportRules = nil
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

        // Log the export event
        exportedRules := strings.Join(exportRules, ", ")
        err = logEvent("EXPORT_RULES", fmt.Sprintf("Exported rules: %s, To file: %s", exportedRules, exportFilePath))
        if err != nil {
        fmt.Printf("Warning: Failed to log event: %v\n", err)
        }

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
        if exitError, ok := err.(*exec.ExitError); ok {
            return fmt.Errorf("command failed with exit code %d: %v", exitError.ExitCode(), err)
        }
        return fmt.Errorf("failed to execute command: %v", err)
    }
    return nil
}

func processBottles(command string, bottleValues map[string]string) string {
    re := regexp.MustCompile(`b%\('([^']+)'\)%b`)
    return re.ReplaceAllStringFunc(command, func(match string) string {
        bottleName := re.FindStringSubmatch(match)[1]
        if value, ok := bottleValues[bottleName]; ok {
            return value
        }
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

func logEvent(eventType, details string) error {
    logPath := filepath.Join(os.Getenv("HOME"), logDir, logFileName)

    err := os.MkdirAll(filepath.Dir(logPath), 0755)
    if err != nil {
        return fmt.Errorf("failed to create log directory: %v", err)
    }

    file, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        return fmt.Errorf("failed to open log file: %v", err)
    }
    defer file.Close()

    user := os.Getenv("USER")
    timestamp := time.Now().Format("2006-01-02 15:04:05")
    ip := getIP()

    logMessage := fmt.Sprintf("[%s] %s %s at %s | %s\n", // i.e User:%s
                              timestamp, eventType, user, ip, details)

    _, err = file.WriteString(logMessage)
    if err != nil {
        return fmt.Errorf("failed to write to log file: %v", err)
    }

    return nil
}

func initConfigFile() error {
    homeDir, err := os.UserHomeDir()
    if err != nil {
        return fmt.Errorf("failed to get home directory: %v", err)
    }

    configDirPath := filepath.Join(homeDir, configDir)
    configPath := filepath.Join(configDirPath, configFileName)

    // Create the directory if it doesn't exist
    err = os.MkdirAll(configDirPath, 0755)
    if err != nil {
        return fmt.Errorf("failed to create config directory: %v", err)
    }

    // Open or create the configuration file with read-write permissions for the user
    file, err := os.OpenFile(configPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
    if err != nil {
        if os.IsPermission(err) {
            log.Printf("Permission error: %v\n", err)
            return fmt.Errorf("failed to open or create config file due to permissions: %v", err)
        }
        log.Printf("Failed to open or create config file: %v\n", err)
        return fmt.Errorf("failed to open or create config file: %v", err)
    }
    defer file.Close()

    return nil
}

func getIP() string {
    addrs, err := net.InterfaceAddrs()
    if err == nil {
        for _, addr := range addrs {
            if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
                if ipnet.IP.To4() != nil {
                    return ipnet.IP.String()
                }
            }
        }
    }
    return "Unknown IP"
}

func isReservedName(name string) bool {
    for _, reserved := range reservedNames {
        if name == reserved {
            return true
        }
    }
    return false
}

func writeLinesWithLock(filename string, lines []string) error {
    file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0644)
    if err != nil {
        return err
    }
    defer file.Close()

    // Lock the file
    if err := unix.Flock(int(file.Fd()), unix.LOCK_EX); err != nil {
        return err
    }
    defer unix.Flock(int(file.Fd()), unix.LOCK_UN)

    // Truncate the file
    if err := file.Truncate(0); err != nil {
        return err
    }
    if _, err := file.Seek(0, 0); err != nil {
        return err
    }

    writer := bufio.NewWriter(file)
    for _, line := range lines {
        _, err = fmt.Fprintln(writer, line)
        if err != nil {
            return err
        }
    }
    return writer.Flush()
}
