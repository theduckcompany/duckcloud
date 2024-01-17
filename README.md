<div align="center">

# DuckCloud

<h3 align='center'> The cloud for your family. </h3>

Keep your family data safe at home with a one click fully encrypted backup.<br/>

<p align="center">
    <a href="https://github.com/theduckcompany/duckcloud/commits/master">
    <img src="https://img.shields.io/github/last-commit/theduckcompany/duckcloud.svg?style=flat-square&logo=github&logoColor=white"
         alt="GitHub last commit">
    <a href="https://github.com/theduckcompany/duckcloud/issues">
    <img src="https://img.shields.io/github/issues-raw/theduckcompany/duckcloud.svg?style=flat-square&logo=github&logoColor=white"
         alt="GitHub issues">
    <a href="https://github.com/theduckcompany/duckcloud/pulls">
    <img src="https://img.shields.io/github/issues-pr-raw/theduckcompany/duckcloud.svg?style=flat-square&logo=github&logoColor=white"
         alt="GitHub pull requests">
</p>
      
<p align="center">
  <a href="#features">Features</a> •  
  <a href="#installation">Installation</a> •
  <a href="#configuration">Configuration</a> •
  <a href="#license">License</a>
</p>

</div>


> [!WARNING]  
> This project is in alpha stage. Things are moving fast and some occasional breaking changes could appears.

## Who's it for?

This project aims to propose an easy solution for the families who want to keep their data safe at home away from the GAFAM. This service aims to be a 
Nextcloud alternative with more simplicity for both the user and the administrator.



## Features / Roadmap
- [x] A virtual file systeme with a deduplication system and a data at rest encryption
- [x] A WebDAV integration to connect all your webdav compliante devices 
- [x] A web interface for managing the users, settings and navigate the files
- [] A contact registry with a CarDAV integration and a web interface
- [] An event registry with a CalDAV integration and a web interface
- [] A backup service with end-to-end encryption available with a few clicks

## Installation

| **OS/Distro** | **Command**   |
|---------------|---------------|
| Archlinux     | yay duckcloud |



#### From sources

Make sure you have Go installed, and that go is in your path.

Clone this repository and cd into the go directory. Then run:

```sh
go install github.com/theduckcompany/duckcloud@{{version}}
```


#### From binaries

Duckcloud is a single binary without any dependences. It's really easy to install. Download a binary from the realease page and put it on your PATH. We have a bunch 
of ways to make this even easier for most platforms. 

The [release page](https://github.com/theduckcompany/duckcloud/releases) includes precompiled binaries for Linux, macOS and Windows for every release. You can also get 
the latest binary of `master` branch from the "Coming soon" pre-release.

> This solution should be reserved for a quick test as no automatic updates are possible


## Configuration

#### Bootstrap











