# PROMPTO

## Overview

`prompto` is a streamlined command-line tool designed to efficiently generate context for prompting Large Language
Models (LLMs). With its innovative approach, it bridges the gap between the limited context size of LLMs and the
necessity for substantial API and documentation during code generation.

## Why did I write `prompto`?

Prompting LLMs with an exhaustive context derived from API and other technical documents aids in generating precise
code. Manually copying and pasting the appropriate context from source files can be laborious. Though tools
like [oak](https://github.com/go-go-golems/oak) facilitate tree-sitter queries on codebases, there's often a need to run
diverse sets of queries on distinct parts of the codebase. Using custom shell scripts to compile this information became
a cumbersome chore, especially when extracting contexts from multiple repositories. Enter `prompto`, crafted to
alleviate these concerns.

## How does `prompto` work?

`prompto` functions by scanning a predefined list of repositories mentioned in `~/.prompto/config.yaml`. For each
repository, the tool searches for a `prompto/` directory. All files within this directory are treated as prompts,
retrievable using the `prompto get` command. In cases where the file is executable, the `prompto get` command will run
it. Consequently, this allows users to present both static files (documentation, example data) and dynamic prompts (live
commands).

Safety is paramount. As `prompto` doesn’t intrinsically verify the safety of the commands it executes, users are
strongly cautioned against employing `prompto` on repositories not under their direct control.

## Usage

### Configuring repositories

To configure the repositories for scanning, edit the `~/.prompto/config.yaml`:

```yaml
repositories:
  - /path/to/repo1
  - /path/to/repo2
  ...
```

### Creating prompts in a repository

For `prompto` to recognize prompts within a repository, create a `prompto/` directory at the root of the repository.
Place any desired files (documentation, example data) or executable scripts within this directory.

### Listing available prompts

To view all available prompts, use:

```
❯ prompto list
```

### Getting a prompt context

To retrieve the context from a specific prompt, use:

```
❯ prompto get [prompt-name]
```

### Safety considerations

Always ensure that repositories added to `prompto` are safe and trusted. As `prompto` doesn’t inherently verify the
safety of executed commands, it's vital to be cautious and avoid using repositories that may contain malicious content.

## Examples

To get a glimpse of available prompts:

```
❯ prompto list
/home/manuel/code/wesen/corporate-headquarters/glazed glazed/definitions
/home/manuel/code/wesen/corporate-headquarters/common-sense cms/data/example-flags.md
/home/manuel/code/wesen/corporate-headquarters/common-sense cms/data/form-dsl.yaml
/home/manuel/code/wesen/corporate-headquarters/common-sense cms/data/glazed-types.md
...
```

Retrieve specific contexts:

```
❯ prompto get cms/sql
// Schema is a struct that represents the schema of a CMS object.
  // This contains all the necessary tables, as well as the main table for the object
  // to which all other tables are joined on its `id` using the `parent_id` field.
  //
  // A CMS object is represents by a main table and multiple additional tables used
...
```

```
❯ prompto get glazed/definitions
File: pkg/cmds/cmds.go
  // CommandDescription contains the necessary information for registering
  // a command with cobra. Because a command gets registered in a verb tree,
  // a full list of Parents all the way to the root needs to be provided.
type CommandDescription struct {
...
```

```
❯ prompto get cms/data/plant-dsl.yaml
# This is the main table that represents the plant object
#
#plant:
  help: This table represents a plant object in the CMS
  fields:
...
```


```
 _______  _______    _______  _______ 
|       ||       |  |       ||       |
|    ___||   _   |  |    ___||   _   |
|   | __ |  | |  |  |   | __ |  | |  |
|   ||  ||  |_|  |  |   ||  ||  |_|  |
|   |_| ||       |  |   |_| ||       |
|_______||_______|  |_______||_______|
 _______  _______  ___      _______  __   __  _______ 
|       ||       ||   |    |       ||  |_|  ||       |
|    ___||   _   ||   |    |    ___||       ||  _____|
|   | __ |  | |  ||   |    |   |___ |       || |_____ 
|   ||  ||  |_|  ||   |___ |    ___||       ||_____  |
|   |_| ||       ||       ||   |___ | ||_|| | _____| |
|_______||_______||_______||_______||_|   |_||_______|
 __   __  __    _  ___      _______  _______  _______  __   __ 
|  | |  ||  |  | ||   |    |       ||   _   ||       ||  | |  |
|  | |  ||   |_| ||   |    |    ___||  |_|  ||  _____||  |_|  |
|  |_|  ||       ||   |    |   |___ |       || |_____ |       |
|       ||  _    ||   |___ |    ___||       ||_____  ||       |
|       || | |   ||       ||   |___ |   _   | _____| ||   _   |
|_______||_|  |__||_______||_______||__| |__||_______||__| |__|
 _______  __   __  _______    _______  _______  _     _  _______  ______  
|       ||  | |  ||       |  |       ||       || | _ | ||       ||    _ | 
|_     _||  |_|  ||    ___|  |    _  ||   _   || || || ||    ___||   | || 
  |   |  |       ||   |___   |   |_| ||  | |  ||       ||   |___ |   |_|| 
  |   |  |       ||    ___|  |    ___||  |_|  ||       ||    ___||    __ |
  |   |  |   _   ||   |___   |   |    |       ||   _   ||   |___ |   |  ||
  |___|  |__| |__||_______|  |___|    |_______||__| |__||_______||___|  ||
 _______  _______ 
|       ||       |
|   _   ||    ___|
|  | |  ||   |___ 
|  |_|  ||    ___|
|       ||   |    
|_______||___|    
 _______  _______  __    _  _______  _______  __   __  _______ 
|       ||       ||  |  | ||       ||       ||  |_|  ||       |
|       ||   _   ||   |_| ||_     _||    ___||       ||_     _|
|       ||  | |  ||       |  |   |  |   |___ |       |  |   |  
|      _||  |_|  ||  _    |  |   |  |    ___| |     |   |   |  
|     |_ |       || | |   |  |   |  |   |___ |   _   |  |   |  
|_______||_______||_|  |__|  |___|  |_______||__| |__|  |___|  
 _______  ______   _______  _______  _______  ___   _______  __    _       
|       ||    _ | |       ||   _   ||       ||   | |       ||  |  | |      
|       ||   | || |    ___||  |_|  ||_     _||   | |   _   ||   |_| |      
|       ||   |_|| |   |___ |       |  |   |  |   | |  | |  ||       |      
|      _||    __ ||    ___||       |  |   |  |   | |  |_|  ||  _    | ___  
|     |_ |   |  |||   |___ |   _   |  |   |  |   | |       || | |   ||   | 
|_______||___|  |||_______||__| |__|  |___|  |___| |_______||_|  |__||___| 
```
