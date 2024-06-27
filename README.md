# :baby: Baby abbreviator
Baby comes as an option to the default "alias" command and it's a simple program to abbreviate long prompts in GNU/Linux terminal.
You can easily set rules, delete them, list them and update them with a clear list of parameters.

:white_check_mark: **INSTALLATION**

Go to release section and download the latest version, in that page you can find the installation instructions for the binary files.

:white_check_mark: **COMPILE YOURSELF**

To compile yourself the source code you need to download the _main.go_ file and create a file named _baby.conf_, then run the following commands:

`go mod init baby`

`go build -o baby`

Previous requirements to compile: golang ('gcc-go', 'golang-bin')

To install Golang in your system run

`sudo dnf golang` or `sudo apt golang` depending on your GNU/Linux distribution.

:pencil: **CREATING RULES**

First step after install the program is run `baby -h` to know about how the script functions. Some examples to create rules in a Fedora system terminal:

`baby -n update "sudo dnf update -y && sudo dnf upgrade -y"` this long command will run after with only type `baby update`

# 🤖 **TESTED ON**

🟢 Elementary OS
🟢 Linux Mint
🟢 MX Linux
🟢 Fedora
🟢 AlmaLinux
🟢 Zorin OS
🟢 Endeavour OS

