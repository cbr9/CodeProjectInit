### Motivation
If you're like me, you always forget to run "git init" whenever you create a project you deem as not important. But then it grows, or it turns up to be more important than you thought, and you wish you could run "git reset" because you fucked up.

This a project setup tool. Every time you create a folder in your projects folder, it will run "git init" and other related tools (like "go mod init" for Go or "cargo init" for Rust)
It can be customized via a config JSON file. Note that this assumes that your projects folder is structured by languages, and they're not all mixed together, as in
* /home/x/Code
   - Python
      * ...
   - Rust
      * ...
   - ...      
### Configuration instructions:
#### projects_dir
The absolute path to your projects root folder. E.g. /home/X/Code/

#### languages
Every language defined here represents a directory in $projects_dir.
Inside each of these languages, you can define up to three parameters:
- **depth** (required): the depth of the new folder, starting from but not including $projects_dir. E.g. in /home/x/Code/Python/test_project, depth would equal 2.
- **excluded_dirs** (optional): a list defining already existing directory names in which no command will be run. Every Unix hidden folder (whose filename starts with a dot) is excluded.
- **extra_cmd** (optional): an extra command, useful for some languages that have their own project setup tools, like Go and Rust with Go Modules or Cargo. Note that if these tools already set up a git repository, there is no problem in running "git init". From the official documentation: *Running git init in an existing repository is safe. It will not overwrite things that are already there*.
