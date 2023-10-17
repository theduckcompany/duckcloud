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



## Features
- A file systeme accessible via a browser application
- A Webdav integration to connect all your devices
- An backup service with end-to-end encryption available with a few clicks
- A quick an easy installation, 5mn max.
- A lot will come soon.

## Installation



#### Download a binary

> This solution should be reserved for a quick test as no automatic updates are possible

Duckcloud is a single binary without any dependences. It's really easy to install. Download a binary from the realease page and put it on your PATH. We have a bunch 
of ways to make this even easier for most platforms. 


#### From sources

> This solution should be reserved for a quick test as no automatic updates are possible

Make sure you have Go installed, and that go is in your path.

Clone this repository and cd into the go directory. Then run:
```sh
git clone https://github.com/theduckcompany/duckcloud
go install ./cmd/duckserver
```


## Configuration

#### Bootstrap











