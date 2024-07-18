# :baby: Baby abbreviator
Baby comes as an option to the default "alias" command and it's a simple program to abbreviate long prompts in GNU/Linux terminal.
You can easily set rules, delete them, list them and update them with a clear list of parameters. It should be functional in any GNU/Linux distribution.

⭐ **FEATURES**

* Simplify your long commands in terminal

* Improve your times at making repetitive tasks

* You don't need to open any config file

* Store, list, update and delete rules

* Run your rules in bulk, i.e. `baby <rule1> <rule2>`

* Import rules from a URL

* Import rules from a local file

* Backup your rules to a local file

* Feeding bottles (adding variables inside a command)


:white_check_mark: **PROGRAMMING LANGUAGE**

Baby is completly writen in Golang.


:ballot_box_with_check: **COMPILE YOURSELF**

If you preffer to compile yourself the source code you need to download the _main.go_ file and create a file named _baby.conf_ in ~.config/baby/, then run the following commands:

`go mod init baby`

`go mod tidy`

`go build -o baby`

**Previous requirements to compile:** golang ('gcc-go', 'golang-bin')

To install Golang in your system run

  `sudo dnf install golang` or `sudo apt install golang` depending on your GNU/Linux distribution.


:pencil: **CREATING RULES**

First step after install the program is run `baby -h` to know about how the script functions. Some examples to create rules in a Fedora system terminal:

  `baby -n update "sudo dnf update -y && sudo dnf upgrade -y"` this long command will run after with only type `baby update`.

  `baby -n ssh "ssh user@example.com"` will connect to your SSH server only typing `baby ssh`

  Running a block of rules is as easy as run `baby <name1> <name2>`. This command will run two rules continuously but you can set as many as your implementation let.

:pencil: **IMPORTING RULES**

  `baby -i <URL or path>` will import rules from a URL or a local file.

  Your URL must to point to a file extension, i.e: .txt, .md, .html, etc.

  The stored rules must follow this sintaxis: `b:<rule> = <command>:b`

:pencil: **EXPORTING RULES**

  `baby -e` will start the backup assistant.

:pencil: **LISTING RULES**

There are two options to list the rules stored in baby.conf file.

  `baby -l` will list all the rules stored in baby.conf file.

  `baby -ln <name>` will list an specific rule.

:pencil: **REMOVING RULES**

  `baby -r <name>` will remove an specific rule.

  `baby -r a` will remove all rules stored in baby.conf.

:pencil: **FEEDING BOTTLES**

  The feeding bottles help you adding a variable inside a command. Use only one bottle for command.

  The feeding bottle sintaxis is this `b%('bottle_name')%b` and you can add it into any part of the command.

  Usage examples: `baby -n ssh "ssh b%('ssh username')%b@example.com"`

  Execute the rule with: `baby ssh` and the system will prompt this:

  _The ssh username is?:_

  If the credentials are valid, you will get connection via ssh to example.com.


# 🤖 **TESTED ON**

🟢 Elementary OS

🟢 Debian

🟢 Linux Mint

🟢 MX Linux

🟢 Fedora

🟢 AlmaLinux

🟢 Zorin OS

🟢 Endeavour OS