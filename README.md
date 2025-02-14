# desertBits

A blog on interesting AI, ML, Robotics, and random other topics. 

##### Blog Code
The blog is a static site generated from markdown files. The frontmatter in each post is in [toml](https://toml.io/en/). Syntax highlighting via [chroma](https://github.com/alecthomas/chroma). The rest is a bit of golang for generating the static files and vanilla html/css/js for the front end. It is meant to be unapologetically simple.

Building the static site
```shell
go run main.go
```

Building the static site with hot reloading and a dev server
```shell
go run main.go --dev
```