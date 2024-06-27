# :baby: Baby abbreviator
Baby comes as an option to the default "alias" command and it's a simple program to abbreviate long prompts in GNU/Linux terminal.
You can easily set rules, delete them, list them and update them with a clear list of parameters. It should be functional in any GNU/Linux distribution.

⭐ **FEATURES**

* Simplify your long commands in terminal

* Improve your times at making repetitive tasks

* You don't need to open any config file

* Store, list, update and delete rules

* Run a block of rules easy

* Coming soon: import rules from a public repository for a specific GNU/Linux distro
  

:white_check_mark: **INSTALLATION**

Go to release section and download the latest version, in that page you can find the installation instructions for the binary files.

  🔗 [Download the latest version](https://github.com/manuwarfare/baby/releases/latest)

:ballot_box_with_check: **COMPILE YOURSELF**

If you preffer to compile yourself the source code you need to download the _main.go_ file and create a file named _baby.conf_, then run the following commands:

`go mod init baby`

`go build -o baby`

**Previous requirements to compile:** golang ('gcc-go', 'golang-bin')

To install Golang in your system run

  `sudo dnf golang` or `sudo apt golang` depending on your GNU/Linux distribution.
  

:pencil: **CREATING RULES**

First step after install the program is run `baby -h` to know about how the script functions. Some examples to create rules in a Fedora system terminal:

  `baby -n update "sudo dnf update -y && sudo dnf upgrade -y"` this long command will run after with only type `baby update`.

  `baby -n ssh "ssh user@example.com"` will connect to your SSH server only typing `baby ssh`

  Running a block of rules is as easy as run `baby <name1> <name2>`. This command will run two rules continuously but you can set as many as your implementation let.
  

:pencil: **LISTING RULES**

There are two options to list the rules stored in baby.conf file.

  `baby -l` will list all the rules stored in baby.conf file.

  `baby -ln <name>` will list a specific rule.

:pencil: **REMOVING RULES**

  `baby -r <name>` will remove a specific rule.
  
  `baby -r a` will remove all rules stored in baby.conf.

# 🤖 **TESTED ON**

🟢 Elementary OS

🟢 Debian

🟢 Linux Mint

🟢 MX Linux

🟢 Fedora

🟢 AlmaLinux

🟢 Zorin OS

🟢 Endeavour OS


